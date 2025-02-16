package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *AuthRepository
	jwtSecret []byte
}

func NewAuthService(repo *AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	payload := User{
		ID:        uuid.New(),
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return s.repo.CreateUser(payload)
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) GetProfile(ctx context.Context) (string, error) {
	claims, ok := ctx.Value("user").(jwt.MapClaims)

	if !ok {
		return "", errors.New("invalid token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return username, nil
}

func (s *AuthService) CreateNewUserByEnvData(username, password string) (bool, error) {
	adminUser, err := s.repo.GetUserByUsername(username)
	if err != nil && err.Error() != "invalid credentials" {
		return false, err
	}
	if adminUser != nil {
		return false, errors.New("admin user already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	payload := User{
		ID:        uuid.New(),
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	return true, s.repo.CreateUser(payload)
}
