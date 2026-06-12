package keychain

import (
	"errors"
	"os"
	"path/filepath"
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

// TestConfigDirOverrideForcesFileBackend verifies that CORTEX_CONFIG_DIR pins
// the encrypted-file backend at the given directory even when a working OS
// keyring is present, and that the keyring is never written to.
func TestConfigDirOverrideForcesFileBackend(t *testing.T) {
	keyring.MockInit() // a working keyring that the override must ignore
	dir := t.TempDir()
	t.Setenv("CORTEX_CONFIG_DIR", dir)
	resetBackend()
	t.Cleanup(resetBackend)

	if got := Backend(); got != "file" {
		t.Fatalf("Backend() = %q, want file", got)
	}

	const host = "gitlab.com"
	if err := SetCredentials(host, "alice", "token-aaa"); err != nil {
		t.Fatalf("SetCredentials: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "credentials.enc")); err != nil {
		t.Fatalf("credentials file not created in CORTEX_CONFIG_DIR: %v", err)
	}
	if _, err := keyring.Get(service, host); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("keyring was touched despite the override (err = %v)", err)
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

// TestConfigDirBlankIsIgnored verifies that a whitespace-only CORTEX_CONFIG_DIR
// does not force the file backend.
func TestConfigDirBlankIsIgnored(t *testing.T) {
	keyring.MockInit()
	t.Setenv("CORTEX_CONFIG_DIR", "   ")
	resetBackend()
	t.Cleanup(resetBackend)

	if got := Backend(); got != "keychain" {
		t.Fatalf("Backend() = %q, want keychain", got)
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
