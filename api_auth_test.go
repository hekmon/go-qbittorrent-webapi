package qbtapi

import (
	"context"
	"testing"
)

func TestAuthenticationDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── explicit login ──────────────────────────────────────
	if err := c.Login(ctx); err != nil {
		t.Fatalf("Login: %v", err)
	}

	// sanity check: session is usable
	ver, err := c.GetApplicationVersion(ctx)
	if err != nil {
		t.Fatalf("GetApplicationVersion after Login: %v", err)
	}
	if ver == "" {
		t.Fatal("GetApplicationVersion returned empty string after Login")
	}
	t.Logf("version after explicit login: %s", ver)

	// ── logout ──────────────────────────────────────────────
	if err := c.Logout(ctx); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// auto-auth should transparently re-establish the session
	ver, err = c.GetApplicationVersion(ctx)
	if err != nil {
		t.Fatalf("GetApplicationVersion after Logout (auto-auth): %v", err)
	}
	if ver == "" {
		t.Fatal("GetApplicationVersion returned empty string after auto-auth")
	}
	t.Logf("version after logout + auto-auth: %s", ver)
}
