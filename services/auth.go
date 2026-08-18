package services

import (
	"errors"
	"flasher/config"
	"flasher/db"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already taken")
)

type AuthService struct {
	cfg  *config.Config
	repo db.Repository
}

func NewAuthService(cfg *config.Config, repo db.Repository) *AuthService {
	return &AuthService{cfg: cfg, repo: repo}
}

type Claims struct {
	UserID int64   `json:"uid"`
	Role   db.Role `json:"role"`
	jwt.RegisteredClaims
}

// Register всегда создаёт пользователя с ролью "user".
// Роль admin через публичный API назначить нельзя — только через bootstrap/CLI.
func (s *AuthService) Register(username, password string) (db.User, error) {
	if len(username) < 3 || len(password) < 8 {
		return db.User{}, errors.New("username or password too short")
	}

	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return db.User{}, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	user := db.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         db.RoleUser,
		CreatedAt:    time.Now(),
	}

	id, err := s.repo.CreateUser(user)
	if err != nil {
		return db.User{}, err
	}
	user.ID = id
	return user, nil
}

func (s *AuthService) Login(username, password string) (string, db.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", db.User{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", db.User{}, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return "", db.User{}, err
	}

	return token, user, nil
}

func (s *AuthService) issueToken(user db.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// CreateAdmin используется только при bootstrap-е сервера или из CLI-команды.
// Никогда не вызывается из HTTP-хендлера.
func (s *AuthService) CreateAdmin(username, password string) (db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	user := db.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         db.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	id, err := s.repo.CreateUser(user)
	if err != nil {
		return db.User{}, err
	}
	user.ID = id
	return user, nil
}