package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"manage_system/pkg/config"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
	jwt.RegisteredClaims
}

type Service struct {
	secret      string
	expire      time.Duration
	issuer      string
	redisClient *redis.Client
}

func NewService(cfg config.JWTConfig, rdb *redis.Client) *Service {
	return &Service{
		secret:      cfg.Secret,
		expire:      time.Duration(cfg.Expire) * time.Second,
		issuer:      cfg.Issuer,
		redisClient: rdb,
	}
}

func (s *Service) GenerateToken(userID uint, username string, roleID uint, roleName string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(s.expire)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", 0, fmt.Errorf("签发Token失败: %w", err)
	}

	return tokenStr, expiresAt.Unix(), nil
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", t.Header["alg"])
			}
			return []byte(s.secret), nil
		})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Token无效")
	}

	return claims, nil
}

func (s *Service) AddToBlacklist(tokenString string, expireAt time.Time) error {
	hash := sha256hex(tokenString)
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return nil
	}

	key := "jwt:blacklist:" + hash
	return s.redisClient.Set(context.Background(), key, "1", ttl).Err()
}

func (s *Service) IsInBlacklist(tokenString string) bool {
	hash := sha256hex(tokenString)
	key := "jwt:blacklist:" + hash
	val, err := s.redisClient.Get(context.Background(), key).Result()
	if err != nil || val != "1" {
		return false
	}
	return true
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
