package storage

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestIsUserCreated проверяет, создавался ли пользователь
func TestIsUserCreated(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store = db

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("testuser").WillReturnRows(rows)

	result := IsUserCreated("testuser")
	assert.True(t, result)
}

// TestIsUserAuthenticated проверяет, аутентифицирован ли пользователь
func TestIsUserAuthenticated(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store = db

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS`).WithArgs("testuser", "hashedpassword").WillReturnRows(rows)

	result := IsUserAuthenticated("testuser", "hashedpassword")
	assert.True(t, result)
}

// TestCreateUser проверяет создание пользователя
func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store = db

	mock.ExpectExec(`INSERT INTO users`).WithArgs("testuser", "hashedpassword", "token").WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateUser("testuser", "hashedpassword", "token")
	assert.NoError(t, err)
}

// TestSetNewTokenByUser проверяет обновление токена пользователя
func TestSetNewTokenByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store = db

	mock.ExpectExec(`UPDATE users SET token`).WithArgs("newtoken", "testuser").WillReturnResult(sqlmock.NewResult(1, 1))

	err = SetNewTokenByUser("testuser", "newtoken")
	assert.NoError(t, err)
}
