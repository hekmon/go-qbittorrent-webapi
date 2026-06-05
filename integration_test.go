package qbtapi

import (
	"net/url"
	"os"
	"testing"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()

	addr := os.Getenv("QBT_ADDR")
	user := os.Getenv("QBT_USER")
	pass := os.Getenv("QBT_PASS")
	if addr == "" || user == "" || pass == "" {
		t.Skip("QBT_ADDR, QBT_USER and QBT_PASS must be set to run integration tests")
	}

	u, err := url.Parse(addr)
	if err != nil {
		t.Fatalf("parsing QBT_ADDR: %v", err)
	}

	c, err := New(u, user, pass)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	return c
}
