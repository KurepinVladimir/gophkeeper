package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"gophkeeper/internal/model"
)

func TestGetByLoginNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	st := New(db)

	mock.ExpectQuery("SELECT id, login, password_hash, created_at").
		WithArgs("u").
		WillReturnError(sql.ErrNoRows)

	_, err = st.GetByLogin(context.Background(), "u")
	require.Error(t, err)
}

func TestUpsertInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	st := New(db)

	sec := &model.Secret{
		UserID:        1,
		Type:          model.SecretText,
		Title:         "t",
		Meta:          "m",
		EncryptedData: []byte{1, 2},
		Version:       1,
		UpdatedAt:     time.Now().UTC(),
		Deleted:       false,
	}
	mock.ExpectQuery("INSERT INTO secrets").
		WithArgs(sec.UserID, string(sec.Type), sec.Title, sec.Meta, sec.EncryptedData, sec.Version, sec.UpdatedAt, sec.Deleted).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	out, err := st.Upsert(context.Background(), sec)
	require.NoError(t, err)
	require.Equal(t, int64(10), out.ID)
}
