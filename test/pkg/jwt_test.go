package pkg_test

import (
	"testing"
	"time"

	"manage_system/pkg/config"
	jwtpkg "manage_system/pkg/jwt"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupJWTTest(t *testing.T) (*jwtpkg.Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test-issuer",
	}
	svc := jwtpkg.NewService(cfg, rdb)
	return svc, mr
}

func TestJWT_GenerateAndParse_Success(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, expiresIn, err := svc.GenerateToken(1, "testuser", 3, "member")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.Greater(t, expiresIn, time.Now().Unix())

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, uint(3), claims.RoleID)
	assert.Equal(t, "member", claims.RoleName)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

func TestJWT_GenerateToken_Expiration(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, expiresIn, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ParseToken(token)
	require.NoError(t, err)

	// exp should be ~7200s from iat
	diff := claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()
	assert.InDelta(t, int64(7200), diff, 2)
	assert.Equal(t, expiresIn, claims.ExpiresAt.Unix())
}

func TestJWT_ParseToken_Expired(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	// Create service with very short expiration
	cfg := config.JWTConfig{
		Secret: "test-secret-key-at-least-32-chars-long!!",
		Expire: -1, // Already expired
		Issuer: "test",
	}
	svc := jwtpkg.NewService(cfg, rdb)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	_, err = svc.ParseToken(token)
	assert.Error(t, err, "expired token should fail parsing")
}

func TestJWT_ParseToken_WrongSecret(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	cfg1 := config.JWTConfig{
		Secret: "secret-key-1-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test",
	}
	svc1 := jwtpkg.NewService(cfg1, rdb)

	cfg2 := config.JWTConfig{
		Secret: "secret-key-2-at-least-32-chars-long!!",
		Expire: 7200,
		Issuer: "test",
	}
	svc2 := jwtpkg.NewService(cfg2, rdb)

	token, _, err := svc1.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	_, err = svc2.ParseToken(token)
	assert.Error(t, err, "token signed with different secret should fail")
}

func TestJWT_ParseToken_InvalidFormat(t *testing.T) {
	svc, _ := setupJWTTest(t)

	_, err := svc.ParseToken("not.a.valid.jwt")
	assert.Error(t, err)
}

func TestJWT_ParseToken_EmptyString(t *testing.T) {
	svc, _ := setupJWTTest(t)

	_, err := svc.ParseToken("")
	assert.Error(t, err)
}

func TestJWT_AddToBlacklist_Success(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	// Initially not blacklisted
	assert.False(t, svc.IsInBlacklist(token))

	// Add to blacklist
	err = svc.AddToBlacklist(token, time.Now().Add(time.Hour))
	require.NoError(t, err)

	// Now should be blacklisted
	assert.True(t, svc.IsInBlacklist(token))
}

func TestJWT_IsInBlacklist_NotAdded(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	assert.False(t, svc.IsInBlacklist(token))
}

func TestJWT_AddToBlacklist_Idempotent(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	// Add twice - should not error
	err = svc.AddToBlacklist(token, time.Now().Add(time.Hour))
	require.NoError(t, err)
	err = svc.AddToBlacklist(token, time.Now().Add(time.Hour))
	require.NoError(t, err, "double blacklist should be idempotent")

	assert.True(t, svc.IsInBlacklist(token))
}

func TestJWT_Blacklist_TTLExpires(t *testing.T) {
	svc, mr := setupJWTTest(t)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	// Add with short TTL
	err = svc.AddToBlacklist(token, time.Now().Add(1*time.Second))
	require.NoError(t, err)
	assert.True(t, svc.IsInBlacklist(token))

	// Fast forward time past TTL
	mr.FastForward(2 * time.Second)

	assert.False(t, svc.IsInBlacklist(token), "blacklist entry should expire after TTL")
}

func TestJWT_Blacklist_AlreadyExpiredToken(t *testing.T) {
	svc, _ := setupJWTTest(t)

	token, _, err := svc.GenerateToken(1, "user", 1, "role")
	require.NoError(t, err)

	// Try to add with already-passed expiry
	err = svc.AddToBlacklist(token, time.Now().Add(-1*time.Hour))
	require.NoError(t, err, "adding expired token to blacklist should not error")

	assert.False(t, svc.IsInBlacklist(token), "token with expired TTL should not be blacklisted")
}
