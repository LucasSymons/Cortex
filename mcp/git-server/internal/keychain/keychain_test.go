package keychain

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

// resetBackend clears the cached backend so the next call re-selects one.
func resetBackend() {
	storeMu.Lock()
	activeStore = nil
	storeMu.Unlock()
}

// TestKeyringBackendRoundTrip exercises the public API against the in-memory
// keyring mock, mirroring how the keychain backend behaves in CI.
func TestKeyringBackendRoundTrip(t *testing.T) {
	keyring.MockInit()
	resetBackend()
	t.Cleanup(resetBackend)

	if got := Backend(); got != "keychain" {
		t.Fatalf("Backend() = %q, want keychain", got)
	}

	const host = "gitlab.com"
	if err := SetCredentials(host, "alice", "token-aaa"); err != nil {
		t.Fatalf("SetCredentials: %v", err)
	}

	user, token, err := GetCredentials(host)
	if err != nil {
		t.Fatalf("GetCredentials: %v", err)
	}
	if user != "alice" || token != "token-aaa" {
		t.Fatalf("got (%q, %q), want (alice, token-aaa)", user, token)
	}

	if err := DeleteCredentials(host); err != nil {
		t.Fatalf("DeleteCredentials: %v", err)
	}
	if _, _, err := GetCredentials(host); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete, GetCredentials err = %v, want ErrNotFound", err)
	}
}

// TestKeyringUnavailableSelectsFile verifies that a Secret Service failure
// causes the file backend to be selected.
func TestKeyringUnavailableSelectsFile(t *testing.T) {
	keyring.MockInitWithError(errors.New("org.freedesktop.secrets not available"))
	resetBackend()
	t.Cleanup(func() {
		keyring.MockInit()
		resetBackend()
	})

	if got := Backend(); got != "file" {
		t.Fatalf("Backend() = %q, want file", got)
	}
}
