package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"
)

type fakeUsers struct {
	byLogin map[string]*model.User
	nextID  int64
}

func (f *fakeUsers) Create(ctx context.Context, u *model.User) error {
	if f.byLogin == nil {
		f.byLogin = map[string]*model.User{}
	}
	f.nextID++
	u.ID = f.nextID
	f.byLogin[u.Login] = &model.User{ID: u.ID, Login: u.Login, PasswordHash: u.PasswordHash}
	return nil
}
func (f *fakeUsers) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u, ok := f.byLogin[login]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}
func (f *fakeUsers) GetByID(ctx context.Context, id int64) (*model.User, error) {
	for _, u := range f.byLogin {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func TestRegisterLogin(t *testing.T) {
	f := &fakeUsers{}
	svc := NewAuthService(f, "jwt-secret")
	ctx := context.Background()

	require.NoError(t, svc.Register(ctx, "user", "pass"))

	tok, uid, err := svc.Login(ctx, "user", "pass")
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	require.NotZero(t, uid)
}

func TestLoginWrongPassword(t *testing.T) {
	f := &fakeUsers{}
	svc := NewAuthService(f, "jwt-secret")
	ctx := context.Background()

	require.NoError(t, svc.Register(ctx, "user", "pass"))
	_, _, err := svc.Login(ctx, "user", "bad")
	require.Error(t, err)
}
