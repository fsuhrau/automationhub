package mcpproxy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrNoToken is returned by a source that has no token.
var ErrNoToken = errors.New("no token")

// TokenSource supplies the hub access token.
type TokenSource interface {
	Token() (string, error)
}

// TokenFunc adapts a function to TokenSource.
type TokenFunc func() (string, error)

func (f TokenFunc) Token() (string, error) { return f() }

// EnvToken reads the token from an environment variable.
func EnvToken(name string) TokenSource {
	return TokenFunc(func() (string, error) {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("$%s: %w", name, ErrNoToken)
	})
}

// FileToken reads the token from a file only its owner can read.
func FileToken(path string) TokenSource {
	return TokenFunc(func() (string, error) {
		info, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%s: %w", path, ErrNoToken)
		}
		if err != nil {
			return "", err
		}
		if info.Mode().Perm()&0o077 != 0 {
			return "", fmt.Errorf("%s is readable by others (mode %o); chmod 600 it", path, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if v := strings.TrimSpace(string(data)); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("%s is empty: %w", path, ErrNoToken)
	})
}

// KeychainToken reads a generic password from the macOS Keychain. On other
// systems it never has a token.
func KeychainToken(service, account string) TokenSource {
	return TokenFunc(func() (string, error) {
		if runtime.GOOS != "darwin" {
			return "", fmt.Errorf("keychain: %w", ErrNoToken)
		}
		args := []string{"find-generic-password", "-s", service, "-w"}
		if account != "" {
			args = append(args, "-a", account)
		}
		out, err := exec.Command("security", args...).Output()
		if err != nil {
			// exit code 44: no such item
			return "", fmt.Errorf("keychain item %q: %w", service, ErrNoToken)
		}
		if v := strings.TrimSpace(string(out)); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("keychain item %q is empty: %w", service, ErrNoToken)
	})
}

// FirstToken uses the first source that has a token. A source that fails
// for another reason than ErrNoToken stops the search, so a misconfigured
// file isn't silently skipped.
func FirstToken(sources ...TokenSource) TokenSource {
	return TokenFunc(func() (string, error) {
		var missing []string
		for _, s := range sources {
			token, err := s.Token()
			if err == nil {
				return token, nil
			}
			if !errors.Is(err, ErrNoToken) {
				return "", err
			}
			missing = append(missing, strings.TrimSuffix(err.Error(), ": "+ErrNoToken.Error()))
		}
		return "", fmt.Errorf("%w (looked in %s)", ErrNoToken, strings.Join(missing, ", "))
	})
}

// CachedToken keeps a token for ttl, so a rotated token is picked up without
// a restart while the Keychain isn't asked on every request.
func CachedToken(source TokenSource, ttl time.Duration) TokenSource {
	var (
		mu      sync.Mutex
		token   string
		fetched time.Time
	)
	return TokenFunc(func() (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if token != "" && time.Since(fetched) < ttl {
			return token, nil
		}
		t, err := source.Token()
		if err != nil {
			return "", err
		}
		token, fetched = t, time.Now()
		return token, nil
	})
}
