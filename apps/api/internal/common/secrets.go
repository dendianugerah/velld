package common

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
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

var once sync.Once
var instance *Secrets

func GetSecrets() *Secrets {
	once.Do(func() {
		instance = loadSecrets()
	})
	return instance
}

func loadSecrets() *Secrets {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

	jwtSecret, err := getRequiredSecret("JWT_SECRET")
	if err != nil {
		log.Fatal(err)
	}

	encryptionKey, err := getRequiredSecret("ENCRYPTION_KEY")
	if err != nil {
		log.Fatal(err)
	}

	if err := validateEncryptionKey(encryptionKey); err != nil {
		log.Fatal(err)
	}

	// Optional admin credentials (for initial setup)
	adminUsernameCredential := os.Getenv("ADMIN_USERNAME_CREDENTIAL")
	adminPasswordCredential := os.Getenv("ADMIN_PASSWORD_CREDENTIAL")

	isAllowSignup := getWithDefault("ALLOW_REGISTER", "true")

	return &Secrets{
		JWTSecret:               jwtSecret,
		EncryptionKey:           encryptionKey,
		AdminUsernameCredential: adminUsernameCredential,
		AdminPasswordCredential: adminPasswordCredential,
		IsAllowSignup:           strings.ToLower(isAllowSignup) == "true",
	}
}

func getRequiredSecret(envVar string) (string, error) {
	secret := strings.TrimSpace(os.Getenv(envVar))
	if secret == "" {
		return "", fmt.Errorf("[ERROR] %s is required but not set. Please set it in your environment or .env file", envVar)
	}

	if strings.HasPrefix(secret, "$(") && strings.HasSuffix(secret, ")") {
		return "", fmt.Errorf("[ERROR] %s appears to be a shell command: %s\n"+
			"Shell commands in .env files are not executed.\n"+
			"Please run the command manually and paste the output:\n"+
			"  Example: openssl rand -hex 32", envVar, secret)
	}

	return secret, nil
}

func getWithDefault(envVar, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(envVar))
	if value == "" {
		return defaultValue
	}
	return value
}

func validateEncryptionKey(key string) error {
	key = strings.TrimSpace(key)

	if len(key) != 64 {
		return fmt.Errorf("[ERROR] ENCRYPTION_KEY must be exactly 64 hexadecimal characters (32 bytes), got %d characters.\n"+
			"Generate a valid key with: openssl rand -hex 32", len(key))
	}

	_, err := hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("[ERROR] ENCRYPTION_KEY must be a valid hexadecimal string.\n"+
			"Generate a valid key with: openssl rand -hex 32\n"+
			"Error: %v", err)
	}

	return nil
}
