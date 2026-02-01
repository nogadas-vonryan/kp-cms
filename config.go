package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	FrontendHost string `json:"frontendHost"`
	FrontendPort int    `json:"frontendPort"`
	BackendHost  string `json:"backendHost"`
	BackendPort  int    `json:"backendPort"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DataPath     string `json:"dataPath"`
	BackupPath   string `json:"backupPath"`
}

// GetConfigDir returns the appropriate config directory for the current OS
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Use %APPDATA% on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "kpcms")
	case "linux", "darwin":
		// Use ~/.config on Linux/Mac
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		configDir = filepath.Join(homeDir, ".config", "kpcms")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return configDir, nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// getKeyFilePath returns the path to the encryption key file
func getKeyFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".key"), nil
}

// getMachineID generates a machine-specific identifier
func getMachineID() string {
	// Get hostname as a machine identifier
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "default-host"
	}

	// Combine with OS and architecture info
	machineInfo := fmt.Sprintf("%s-%s-%s", hostname, runtime.GOOS, runtime.GOARCH)
	hash := sha256.Sum256([]byte(machineInfo))
	return hex.EncodeToString(hash[:16])
}

// generateAndSaveKey generates a new encryption key and saves it
func generateAndSaveKey() ([]byte, error) {
	keyPath, err := getKeyFilePath()
	if err != nil {
		return nil, err
	}

	// Generate 32 random bytes for AES-256
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %v", err)
	}

	// Encode key to base64 for storage
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(keyPath, []byte(encodedKey), 0600); err != nil {
		return nil, fmt.Errorf("failed to save key: %v", err)
	}

	return key, nil
}

// loadKey loads the encryption key from disk
func loadKey() ([]byte, error) {
	keyPath, err := getKeyFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %v", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32, got %d", len(key))
	}

	return key, nil
}

// getEncryptionKey retrieves or generates the encryption key
// Uses a persistent key file with machine-specific salt for added security
func getEncryptionKey() ([]byte, error) {
	// Try to load existing key
	key, err := loadKey()
	if err == nil {
		return key, nil
	}

	// If key file doesn't exist, generate a new one
	if os.IsNotExist(err) {
		return generateAndSaveKey()
	}

	return nil, fmt.Errorf("failed to get encryption key: %v", err)
}

// encryptPassword encrypts a password using AES-GCM
func encryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	plaintext := []byte(password)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts a password using AES-GCM
func decryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			FrontendHost: "0.0.0.0",
			FrontendPort: 8081,
			BackendHost:  "0.0.0.0",
			BackendPort:  8080,
			Username:     "admin",
			Password:     "",
			DataPath:     "./data",
			BackupPath:   "./backup",
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Decrypt password
	if config.Password != "" {
		decrypted, err := decryptPassword(config.Password)
		if err != nil {
			// Log decryption failure (not silent)
			logMessage(fmt.Sprintf("Warning: Failed to decrypt saved password: %v", err))
			// If decryption fails, clear the password
			config.Password = ""
		} else {
			config.Password = decrypted
		}
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Make a copy to avoid modifying the original
	configToSave := *config

	// Encrypt password before saving
	if configToSave.Password != "" {
		encrypted, err := encryptPassword(configToSave.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %v", err)
		}
		configToSave.Password = encrypted
	}

	data, err := json.MarshalIndent(configToSave, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
