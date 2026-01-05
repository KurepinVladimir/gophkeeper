package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIssueParse(t *testing.T) {
	secret := "s"
	tok, err := IssueToken(secret, 123, time.Minute)
	require.NoError(t, err)

	claims, err := ParseToken(secret, tok)
	require.NoError(t, err)
	require.Equal(t, int64(123), claims.UserID)
}
