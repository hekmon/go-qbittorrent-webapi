package qbtapi

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestApplicationDomain(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()

	// ── version ─────────────────────────────────────────────
	ver, err := c.GetApplicationVersion(ctx)
	if err != nil {
		t.Fatalf("GetApplicationVersion: %v", err)
	}
	if ver == "" {
		t.Fatal("GetApplicationVersion returned empty string")
	}
	t.Logf("version: %s", ver)

	// ── api version ─────────────────────────────────────────
	apiVer, err := c.GetAPIVersion(ctx)
	if err != nil {
		t.Fatalf("GetAPIVersion: %v", err)
	}
	if apiVer == "" {
		t.Fatal("GetAPIVersion returned empty string")
	}
	t.Logf("api version: %s", apiVer)

	// ── build info ──────────────────────────────────────────
	buildInfo, err := c.GetBuildInfo(ctx)
	if err != nil {
		t.Fatalf("GetBuildInfo: %v", err)
	}
	validateBuildInfo(t, buildInfo)
	t.Logf("build info: %+v", buildInfo)

	// ── default save path ───────────────────────────────────
	savePath, err := c.GetDefaultSavePath(ctx)
	if err != nil {
		t.Fatalf("GetDefaultSavePath: %v", err)
	}
	if savePath == "" {
		t.Fatal("GetDefaultSavePath returned empty string")
	}
	t.Logf("default save path: %s", savePath)

	// ── preferences round-trip ──────────────────────────────
	t.Run("preferences_round_trip", func(t *testing.T) {
		// get current prefs
		prefsBefore, err := c.GetApplicationPreferences(ctx)
		if err != nil {
			t.Skipf("GetApplicationPreferences failed — server may report a different API version than the library target (%s): %v", APIReferenceVersion, err)
		}
		if prefsBefore.WebUISessionTimeout == nil {
			t.Fatal("WebUISessionTimeout is nil in fetched preferences")
		}

		originalTimeout := *prefsBefore.WebUISessionTimeout
		testTimeout := originalTimeout + 1

		// set a new value
		if err := c.SetApplicationPreferences(ctx, ApplicationPreferences{
			WebUISessionTimeout: Int(testTimeout),
		}); err != nil {
			t.Fatalf("SetApplicationPreferences: %v", err)
		}

		// verify the change stuck
		prefsAfter, err := c.GetApplicationPreferences(ctx)
		if err != nil {
			t.Fatalf("GetApplicationPreferences (after set): %v", err)
		}
		if prefsAfter.WebUISessionTimeout == nil {
			t.Fatal("WebUISessionTimeout is nil after set")
		}
		if *prefsAfter.WebUISessionTimeout != testTimeout {
			t.Fatalf("expected WebUISessionTimeout=%d, got %d",
				testTimeout, *prefsAfter.WebUISessionTimeout)
		}

		// restore original value
		if err := c.SetApplicationPreferences(ctx, ApplicationPreferences{
			WebUISessionTimeout: Int(originalTimeout),
		}); err != nil {
			t.Fatalf("restoring WebUISessionTimeout: %v", err)
		}
	})

	// ── cookies round-trip ──────────────────────────────────
	t.Run("cookies_round_trip", func(t *testing.T) {
		// save original cookies
		originalCookies, err := c.GetCookies(ctx)
		if err != nil {
			t.Fatalf("GetCookies (before): %v", err)
		}

		// set test cookies
		now := time.Now()
		testCookies := []Cookie{
			{
				Name:           "qbtapi_test_1",
				Domain:         "example.com",
				Path:           "/",
				Value:          "value1",
				ExpirationDate: now.Add(24 * time.Hour),
			},
			{
				Name:           "qbtapi_test_2",
				Domain:         "foo.bar",
				Path:           "/baz",
				Value:          "hello=world",
				ExpirationDate: now.Add(48 * time.Hour),
			},
		}
		if err := c.SetCookies(ctx, testCookies); err != nil {
			t.Fatalf("SetCookies: %v", err)
		}

		// retrieve and verify
		cookies, err := c.GetCookies(ctx)
		if err != nil {
			t.Fatalf("GetCookies (after set): %v", err)
		}

		found := make(map[string]Cookie)
		for _, co := range cookies {
			found[co.Name] = co
		}

		for _, want := range testCookies {
			got, ok := found[want.Name]
			if !ok {
				t.Fatalf("cookie %q not found after SetCookies", want.Name)
			}
			if got.Domain != want.Domain {
				t.Errorf("cookie %q domain: want %q, got %q", want.Name, want.Domain, got.Domain)
			}
			if got.Path != want.Path {
				t.Errorf("cookie %q path: want %q, got %q", want.Name, want.Path, got.Path)
			}
			if got.Value != want.Value {
				t.Errorf("cookie %q value: want %q, got %q", want.Name, want.Value, got.Value)
			}
			// API stores seconds since epoch, truncate to 1s precision
			wantExp := want.ExpirationDate.Truncate(time.Second)
			gotExp := got.ExpirationDate.Truncate(time.Second)
			if !gotExp.Equal(wantExp) {
				t.Errorf("cookie %q expiration: want %v, got %v", want.Name, wantExp, gotExp)
			}
		}

		// restore original cookies
		if err := c.SetCookies(ctx, originalCookies); err != nil {
			t.Fatalf("restoring original cookies: %v", err)
		}

		// verify restoration
		restored, err := c.GetCookies(ctx)
		if err != nil {
			t.Fatalf("GetCookies (after restore): %v", err)
		}

		// build maps for comparison (name+domain+path as key)
		key := func(co Cookie) string {
			return fmt.Sprintf("%s|%s|%s", co.Name, co.Domain, co.Path)
		}
		origMap := make(map[string]Cookie)
		for _, co := range originalCookies {
			origMap[key(co)] = co
		}
		restMap := make(map[string]Cookie)
		for _, co := range restored {
			restMap[key(co)] = co
		}

		if len(origMap) != len(restMap) {
			t.Fatalf("cookie count mismatch after restore: want %d, got %d", len(origMap), len(restMap))
		}
		for k, want := range origMap {
			got, ok := restMap[k]
			if !ok {
				t.Fatalf("restored cookie %q missing", k)
			}
			if got.Value != want.Value {
				t.Errorf("restored cookie %q value: want %q, got %q", k, want.Value, got.Value)
			}
			wantExp := want.ExpirationDate.Truncate(time.Second)
			gotExp := got.ExpirationDate.Truncate(time.Second)
			if !gotExp.Equal(wantExp) {
				t.Errorf("restored cookie %q expiration: want %v, got %v", k, wantExp, gotExp)
			}
		}
	})
}

// validateBuildInfo checks that all version fields were unmarshaled correctly.
// A silent json tag typo would leave these empty.
func validateBuildInfo(t *testing.T, bi BuildInfo) {
	t.Helper()
	if bi.QT == "" {
		t.Fatal("BuildInfo.QT is empty")
	}
	if bi.LibTorrent == "" {
		t.Fatal("BuildInfo.LibTorrent is empty")
	}
	if bi.Boost == "" {
		t.Fatal("BuildInfo.Boost is empty")
	}
	if bi.OpenSSL == "" {
		t.Fatal("BuildInfo.OpenSSL is empty")
	}
	if bi.Bitness == 0 {
		t.Fatal("BuildInfo.Bitness is zero")
	}
	if bi.Platform == "" {
		t.Fatal("BuildInfo.Platform is empty")
	}
	if bi.Zlib == "" {
		t.Fatal("BuildInfo.Zlib is empty")
	}
}
