package common

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Secrets struct {
	JWTSecret               string
	EncryptionKey           string
	AdminUsernameCredential string
	AdminPasswordCredential string
	IsAllowSignup           bool
}

var secretsFilePath = filepath.Join("..", "..", ".env")

var once sync.Once
var instance *Secrets

func GetSecrets() *Secrets {
	once.Do(func() {
		instance = loadOrGenerateSecrets()
	})
	return instance
}

func loadOrGenerateSecrets() *Secrets {
	err := godotenv.Load(secretsFilePath)
	if err != nil {
		log.Println("[WARN] No .env file found, will generate secrets if missing")
	}

	jwtSecret := getOrGenerateSecret("JWT_SECRET")
	encryptionKey := getOrGenerateSecret("ENCRYPTION_KEY")
	adminUsernameCredential, _ := getWithoutGenerateSecret("ADMIN_USERNAME_CREDENTIAL")
	adminPasswordCredential, _ := getWithoutGenerateSecret("ADMIN_PASSWORD_CREDENTIAL")
	isAllowSignup := getOrGenerateSecret("ALLOW_REGISTER")

	saveSecretsToFile(jwtSecret, encryptionKey, isAllowSignup)

	return &Secrets{
		JWTSecret:               jwtSecret,
		EncryptionKey:           encryptionKey,
		AdminUsernameCredential: adminUsernameCredential,
		AdminPasswordCredential: adminPasswordCredential,
		IsAllowSignup:           isAllowSignup == "true",
	}
}

// getWithoutGenerateSecret retrieves a secret from the environment variables without generating a new one
func getWithoutGenerateSecret(envVar string) (string, error) {
	if secret := os.Getenv(envVar); secret != "" {
		return secret, nil
	}
	return "", fmt.Errorf("secret not found in env")
}

func getOrGenerateSecret(envVar string) string {
	if secret := os.Getenv(envVar); secret != "" {
		return secret
	}

	var newSecret string
	if envVar == "ENCRYPTION_KEY" {
		newSecret = generateSecureHexKey()
	} else if envVar == "ALLOW_REGISTER" {
		newSecret = generateAllowSingup()
	} else {
		newSecret = generateSecureBase64Key()
	}

	fmt.Printf("[INFO] %s not found in env. Generated new key.\n", envVar)
	return newSecret
}

// Generate allow signup
func generateAllowSingup() string {
	return "true"
}

// Generate a secure random hex key (for AES encryption)
func generateSecureHexKey() string {
	key := make([]byte, 32) // 32 bytes = 64 hex characters
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Failed to generate encryption key:", err)
	}
	return hex.EncodeToString(key) // Convert to hex string
}

// Generate a secure random base64 key (for JWT)
func generateSecureBase64Key() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		panic("Failed to generate secure key")
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func saveSecretsToFile(jwtSecret, encryptionKey, isAllowSignup string) {
	existingContent, _ := os.ReadFile(secretsFilePath)
	existingEnv := string(existingContent)

	if !containsLine(existingEnv, "JWT_SECRET=") {
		existingEnv += fmt.Sprintf("\nJWT_SECRET=%s", jwtSecret)
	}

	if !containsLine(existingEnv, "ENCRYPTION_KEY=") {
		existingEnv += fmt.Sprintf("\nENCRYPTION_KEY=%s", encryptionKey)
	}
	if !containsLine(existingEnv, "ALLOW_REGISTER=") {
		existingEnv += fmt.Sprintf("\nALLOW_REGISTER=%s", isAllowSignup)
	}

	os.WriteFile(secretsFilePath, []byte(existingEnv), 0600)
}

func containsLine(envContent, prefix string) bool {
	lines := strings.Split(envContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
