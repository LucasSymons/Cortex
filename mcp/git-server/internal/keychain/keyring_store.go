package keychain

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// keyringStore stores credentials in the OS keychain via zalando/go-keyring.
// Each host occupies two entries: one for the username and one for the token.
type keyringStore struct{}

func (keyringStore) kind() string { return "keychain" }

// Set stores the username and token for the given host.
func (keyringStore) Set(host, username, token string) error {
	if err := keyring.Set(service, hostUsernameKey(host), username); err != nil {
		return fmt.Errorf("storing username: %w", err)
	}
	if err := keyring.Set(service, hostTokenKey(host), token); err != nil {
		return fmt.Errorf("storing token: %w", err)
	}
	return nil
}

// Get retrieves the username and token for the given host.
func (keyringStore) Get(host string) (username, token string, err error) {
	username, err = keyring.Get(service, hostUsernameKey(host))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", fmt.Errorf("retrieving username for %s: %w", host, err)
	}
	token, err = keyring.Get(service, hostTokenKey(host))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", "", ErrNotFound
		}
		return "", "", fmt.Errorf("retrieving token for %s: %w", host, err)
	}
	return username, token, nil
}

// Delete removes the stored credentials for the given host.
func (keyringStore) Delete(host string) error {
	if err := keyring.Delete(service, hostUsernameKey(host)); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("deleting username: %w", err)
	}
	if err := keyring.Delete(service, hostTokenKey(host)); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("deleting token: %w", err)
	}
	return nil
}

func hostUsernameKey(host string) string { return host + ":username" }
func hostTokenKey(host string) string    { return host + ":token" }
