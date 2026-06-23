package jwt

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
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
	secret       string
	expire       time.Duration
	issuer       string
	redisClient  *redis.Client
	memBlacklist sync.Map // key: hash, value: expireAt (time.Time) — in-memory fallback when Redis is nil
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

	// Unique JWT ID prevents identical tokens when refreshed in the same second.
	// HMAC-SHA256 is deterministic — without jti, two GenerateToken calls with
	// the same claims + same timestamp produce the same token string, which
	// would already be blacklisted (refresh → logout loop).
	jti, err := genJTI()
	if err != nil {
		return "", 0, fmt.Errorf("生成JTI失败: %w", err)
	}

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RoleID:   roleID,
		RoleName: roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
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

// genJTI generates a unique JWT ID: unix nano timestamp + 4 random bytes (hex).
func genJTI() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1<<32))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%x-%x", time.Now().UnixNano(), b, n), nil
}

// ParseTokenForRefresh 解析 token 但不验证过期时间 — 刷新端点专用。
// 签名、黑名单等检查由 RefreshToken handler 单独完成。
func (s *Service) ParseTokenForRefresh(tokenString string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, err := parser.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("非预期的签名方法: %v", t.Header["alg"])
			}
			return []byte(s.secret), nil
		})

	if err != nil {
		return nil, err
	}
	if claims.ID == "" {
		return nil, fmt.Errorf("Token无效")
	}

	return claims, nil
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

func (s *Service) AddToBlacklist(ctx context.Context, tokenString string, expireAt time.Time) error {
	hash := sha256hex(tokenString)
	ttl := time.Until(expireAt)
	if ttl <= 0 {
		return nil
	}

	if s.redisClient == nil {
		// In-memory fallback when no Redis (e.g. integration tests)
		s.memBlacklist.Store(hash, expireAt)
		return nil
	}

	key := "jwt:blacklist:" + hash
	// Use caller's context (request-scoped), with a short timeout so a slow
	// Redis cannot hold the goroutine hostage after the HTTP client has
	// already disconnected.
	redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.redisClient.Set(redisCtx, key, "1", ttl).Err()
}

func (s *Service) IsInBlacklist(ctx context.Context, tokenString string) bool {
	hash := sha256hex(tokenString)

	if s.redisClient == nil {
		// In-memory fallback when no Redis (e.g. integration tests)
		if expireAt, ok := s.memBlacklist.Load(hash); ok {
			if time.Now().Before(expireAt.(time.Time)) {
				return true
			}
			// Expired, clean up
			s.memBlacklist.Delete(hash)
		}
		return false
	}

	key := "jwt:blacklist:" + hash
	redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	val, err := s.redisClient.Get(redisCtx, key).Result()
	if err != nil || val != "1" {
		return false
	}
	return true
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
