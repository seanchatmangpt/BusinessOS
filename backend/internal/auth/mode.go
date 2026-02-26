// Package auth implements the three-tier authentication system for BusinessOS.
//
// Tier 1: single — Zero-auth mode. Default for fresh installs and personal use.
//
//	A default owner user is auto-created on first boot. No login screen.
//
// Tier 2: local — Email/password with bcrypt. Invite-only after first user.
//
//	Works completely offline.
//
// Tier 3: oauth — Google and/or GitHub OAuth. Requires credentials in .env.
//
// Tiers can be combined: "local+oauth" enables both email and social login.
package auth

import "strings"

// AuthMode controls which authentication strategy is active.
type AuthMode string

const (
	// AuthModeSingle is the default. No login screen; a permanent owner
	// session is injected automatically on every request.
	AuthModeSingle AuthMode = "single"

	// AuthModeLocal enables email/password authentication.
	AuthModeLocal AuthMode = "local"

	// AuthModeOAuth enables OAuth providers only (Google, GitHub).
	AuthModeOAuth AuthMode = "oauth"

	// AuthModeHybrid enables both email/password and OAuth.
	AuthModeHybrid AuthMode = "local+oauth"
)

// ParseAuthMode converts a string from .env to an AuthMode.
// Unknown values fall back to AuthModeSingle.
func ParseAuthMode(s string) AuthMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "local":
		return AuthModeLocal
	case "oauth":
		return AuthModeOAuth
	case "local+oauth", "oauth+local":
		return AuthModeHybrid
	default:
		return AuthModeSingle
	}
}

// RequiresLogin returns true when the mode requires the user to authenticate
// before accessing protected routes.
func (m AuthMode) RequiresLogin() bool {
	return m != AuthModeSingle
}

// AllowsLocalAuth returns true when email/password login is supported.
func (m AuthMode) AllowsLocalAuth() bool {
	return m == AuthModeLocal || m == AuthModeHybrid
}

// AllowsOAuth returns true when OAuth providers are supported.
func (m AuthMode) AllowsOAuth() bool {
	return m == AuthModeOAuth || m == AuthModeHybrid
}
