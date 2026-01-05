package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims are JWT claims used by the server.
type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// IssueToken issues a JWT for given user id.
func IssueToken(secret string, userID int64, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseToken parses and validates token string and returns claims.
func ParseToken(secret, tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
