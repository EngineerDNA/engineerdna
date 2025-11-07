package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// KeyStore manages encryption keys
type KeyStore struct {
	masterKey []byte
}

// NewKeyStore creates a new key store
// It tries to load the master key from:
// 1. Environment variable ENGINEERDNA_MASTER_KEY
// 2. Generate a new key if not found
func NewKeyStore() (*KeyStore, error) {
	var key []byte

	// Try environment variable first
	keyB64 := os.Getenv("ENGINEERDNA_MASTER_KEY")
	if keyB64 != "" {
		var err error
		key, err = base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ENGINEERDNA_MASTER_KEY: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("ENGINEERDNA_MASTER_KEY must be 32 bytes (256 bits)")
		}
	} else {
		// Generate new key
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate master key: %w", err)
		}
		fmt.Printf("Generated new master key. Set this environment variable to persist:\n")
		fmt.Printf("export ENGINEERDNA_MASTER_KEY=%s\n", base64.StdEncoding.EncodeToString(key))
	}

	return &KeyStore{masterKey: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (k *KeyStore) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (k *KeyStore) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(k.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptPluginConfig encrypts secret fields in plugin configuration
func (k *KeyStore) EncryptPluginConfig(config map[string]string, secretFields []string) (map[string]string, error) {
	encrypted := make(map[string]string)

	for key, value := range config {
		if contains(secretFields, key) {
			enc, err := k.Encrypt(value)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", key, err)
			}
			encrypted[key] = enc
		} else {
			encrypted[key] = value
		}
	}

	return encrypted, nil
}

// DecryptPluginConfig decrypts secret fields in plugin configuration
func (k *KeyStore) DecryptPluginConfig(config map[string]string, secretFields []string) (map[string]string, error) {
	decrypted := make(map[string]string)

	for key, value := range config {
		if contains(secretFields, key) {
			dec, err := k.Decrypt(value)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", key, err)
			}
			decrypted[key] = dec
		} else {
			decrypted[key] = value
		}
	}

	return decrypted, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
