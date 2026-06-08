package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// encryptedPayload wraps the AES-GCM output so we can tell encrypted and
// plaintext token files apart on disk. The magic prefix "v1:" is a forward-
// compatible version tag.
const encryptedMagic = "v1:"

// Overridable OS shims for test injection.
var (
	osUserHomeDir = os.UserHomeDir
	osMkdirAll    = os.MkdirAll
	osWriteFile   = os.WriteFile
)

// machineKeyPath is the on-disk key file (next to auth.json). It is created
// with 0600 permissions on first use and never leaves the user's home dir.
func machineKeyPath() (string, error) {
	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claw-code", "auth.key"), nil
}

var (
	keyOnce sync.Once
	keyData []byte
	keyErr  error
)

// loadOrCreateMachineKey returns a 32-byte AES key, creating a random one on
// first use. The key is stored in a 0600 file in the user's claw-code dir
// so only that user can read it. The OS file permissions are the only line
// of defence — anyone with root can still recover the key.
func loadOrCreateMachineKey() ([]byte, error) {
	keyOnce.Do(func() {
		path, err := machineKeyPath()
		if err != nil {
			keyErr = err
			return
		}
		if existing, err := os.ReadFile(path); err == nil && len(existing) == 64 {
			raw, decErr := hex.DecodeString(string(existing))
			if decErr == nil && len(raw) == 32 {
				keyData = raw
				return
			}
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			keyErr = fmt.Errorf("generate machine key: %w", err)
			return
		}
		if err := osMkdirAll(filepath.Dir(path), 0o700); err != nil {
			keyErr = err
			return
		}
		if err := osWriteFile(path, []byte(hex.EncodeToString(raw)), 0o600); err != nil {
			keyErr = err
			return
		}
		keyData = raw
	})
	return keyData, keyErr
}

// encryptTokenData returns the AES-GCM ciphertext of plaintext, prefixed
// with the version tag so LoadTokens can decrypt transparently.
func encryptTokenData(plaintext []byte) ([]byte, error) {
	key, err := loadOrCreateMachineKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	out := make([]byte, 0, len(encryptedMagic)+len(sealed))
	out = append(out, []byte(encryptedMagic)...)
	out = append(out, sealed...)
	return out, nil
}

// decryptTokenData reverses encryptTokenData. Returns an error if the file
// is not in the encrypted format (e.g. a legacy plaintext auth.json) so
// the caller can decide whether to migrate or surface a clear message.
func decryptTokenData(blob []byte) ([]byte, error) {
	if len(blob) < len(encryptedMagic) || string(blob[:len(encryptedMagic)]) != encryptedMagic {
		return nil, fmt.Errorf("auth: token file is not in encrypted format")
	}
	payload := blob[len(encryptedMagic):]
	key, err := loadOrCreateMachineKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm: %w", err)
	}
	if len(payload) < gcm.NonceSize() {
		return nil, fmt.Errorf("auth: ciphertext truncated")
	}
	nonce, ct := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("auth: decrypt token file (key may have changed): %w", err)
	}
	return pt, nil
}

// fingerprintKey returns a non-secret, non-cryptographic fingerprint of the
// machine key for use in debug log lines. Truncated SHA-256 hex, first 8 chars.
func fingerprintKey() string {
	key, err := loadOrCreateMachineKey()
	if err != nil {
		return "unknown"
	}
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:])[:8]
}
