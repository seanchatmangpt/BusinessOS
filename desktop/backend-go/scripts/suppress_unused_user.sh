#!/bin/bash
# Add _ = user after GetCurrentUser in functions where user is genuinely unused

cat > /tmp/suppress_user.py << 'EOF'
#!/usr/bin/env python3
import re
import sys

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Pattern: Find GetCurrentUser followed by comment, where user is NOT used later
    # We'll add `_ = user` right after the comment

    pattern = re.compile(
        r'(user := middleware\.GetCurrentUser\(c\)\n'
        r'\s*// Auth guaranteed by middleware - user cannot be nil here\n)',
        re.MULTILINE
    )

    def check_and_replace(match):
        start_pos = match.end()
        # Check if 'user' appears later (not just in comments)
        rest = content[start_pos:]

        # Find the end of the current function
        brace_count = 1
        func_end = 0
        for i, char in enumerate(rest):
            if char == '{':
                brace_count += 1
            elif char == '}':
                brace_count -= 1
                if brace_count == 0:
                    func_end = i
                    break

        func_body = rest[:func_end] if func_end > 0 else rest

        # Check if user is used (as user.Something or user,)
        if not re.search(r'\buser\.[a-zA-Z]|\buser[,\)]', func_body):
            # User is unused - add suppression
            return match.group(0) + '\t_ = user // Suppress unused variable warning\n'
        return match.group(0)

    modified_content = pattern.sub(check_and_replace, content)

    if modified_content != content:
        with open(filepath, 'w') as f:
            f.write(modified_content)
        return True
    return False

if __name__ == '__main__':
    changed = process_file(sys.argv[1])
    if changed:
        print(f"✅ {sys.argv[1]}")
EOF

chmod +x /tmp/suppress_user.py

echo "Adding unused variable suppressions..."
python3 /tmp/suppress_user.py internal/handlers/agents.go
python3 /tmp/suppress_user.py internal/handlers/crm.go

go build ./cmd/server 2>&1 | head -15
