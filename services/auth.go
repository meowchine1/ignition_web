package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"flasher/config"
	"flasher/db"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrAdminExists        = errors.New("admin already exists")
)

// Argon2id параметры (OWASP recommended)
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
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

// ---------- Хэширование паролей (Argon2id) ----------

// hashPassword хэширует пароль через Argon2id и возвращает строку в формате:
// $argon2id$v=19$m=65536,t=1,p=4$<salt-base64>$<hash-base64>
func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// verifyPassword проверяет пароль против сохранённого хэша.
// Поддерживает Argon2id (новые) и bcrypt (старые, для обратной совместимости).
func verifyPassword(password, encodedHash string) (bool, error) {
	// Argon2id формат: $argon2id$v=19$m=...,t=...,p=...$salt$hash
	if strings.HasPrefix(encodedHash, "$argon2id$") {
		return verifyArgon2id(password, encodedHash)
	}

	// Обратная совместимость: bcrypt-хэши от старых версий
	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") || strings.HasPrefix(encodedHash, "$2y$") {
		if err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password)); err != nil {
			return false, nil
		}
		return true, nil
	}

	return false, errors.New("unsupported hash format")
}

// verifyArgon2id проверяет пароль против Argon2id-хэша.
func verifyArgon2id(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}

	var memory, iterations, parallelism int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}
	return false, nil
}

// ---------- Регистрация / Логин ----------

// Register всегда создаёт пользователя с ролью "user".
// Роль admin через публичный API назначить нельзя — только через CLI bootstrap.
func (s *AuthService) Register(username, password string) (db.User, error) {
	if len(username) < 3 || len(password) < 8 {
		return db.User{}, errors.New("username or password too short")
	}

	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return db.User{}, ErrUsernameTaken
	}

	hash, err := hashPassword(password)
	if err != nil {
		return db.User{}, err
	}

	user := db.User{
		Username:     username,
		PasswordHash: hash,
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

	ok, err := verifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
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

// ---------- CLI bootstrap первого admin ----------

// CreateFirstAdmin создаёт первого администратора.
// Если admin уже существует — возвращает ErrAdminExists и ничего не перезаписывает.
func (s *AuthService) CreateFirstAdmin(username, password string) (db.User, error) {
	if len(username) < 3 || len(password) < 8 {
		return db.User{}, errors.New("username or password too short")
	}

	// Проверяем, существует ли уже admin
	count, err := s.repo.CountUsersByRole(db.RoleAdmin)
	if err != nil {
		return db.User{}, err
	}
	if count > 0 {
		return db.User{}, ErrAdminExists
	}

	hash, err := hashPassword(password)
	if err != nil {
		return db.User{}, err
	}

	user := db.User{
		Username:     username,
		PasswordHash: hash,
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

// ---------- Admin API: управление пользователями ----------

// ListUsers возвращает всех пользователей (без password_hash).
func (s *AuthService) ListUsers() ([]db.User, error) {
	return s.repo.ListUsers()
}

// CreateUserByAdmin создаёт пользователя с указанной ролью (только admin).
func (s *AuthService) CreateUserByAdmin(username, password string, role db.Role) (db.User, error) {
	if len(username) < 3 || len(password) < 8 {
		return db.User{}, errors.New("username or password too short")
	}
	if role != db.RoleUser && role != db.RoleAdmin {
		return db.User{}, errors.New("invalid role")
	}

	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return db.User{}, ErrUsernameTaken
	}

	hash, err := hashPassword(password)
	if err != nil {
		return db.User{}, err
	}

	user := db.User{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    time.Now(),
	}

	id, err := s.repo.CreateUser(user)
	if err != nil {
		return db.User{}, err
	}
	user.ID = id
	return user, nil
}

// UpdateUserByAdmin обновляет роль пользователя (только admin).
func (s *AuthService) UpdateUserByAdmin(id int64, role db.Role) (db.User, error) {
	if role != db.RoleUser && role != db.RoleAdmin {
		return db.User{}, errors.New("invalid role")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return db.User{}, ErrUserNotFound
	}

	user.Role = role
	if err := s.repo.UpdateUser(user); err != nil {
		return db.User{}, err
	}

	return user, nil
}

// DeleteUserByAdmin удаляет пользователя (только admin).
// Нельзя удалить самого себя.
func (s *AuthService) DeleteUserByAdmin(id int64, currentUserID int64) error {
	if id == currentUserID {
		return errors.New("cannot delete yourself")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	// Нельзя удалить последнего admin
	if user.Role == db.RoleAdmin {
		count, err := s.repo.CountUsersByRole(db.RoleAdmin)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("cannot delete the last admin")
		}
	}

	return s.repo.DeleteUser(id)
}

// ---------- Смена / reset пароля ----------

// ChangePassword — смена пароля текущим пользователем (требует старый пароль).
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("new password too short")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	ok, err := verifyPassword(oldPassword, user.PasswordHash)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.repo.UpdateUser(user)
}

// ResetPassword — принудительный сброс пароля администратором.
func (s *AuthService) ResetPassword(userID int64, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("new password too short")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	return s.repo.UpdateUser(user)
}