package v5_test

import (
	"fmt"
	"testing"

	v5 "gin-tutorial/app/domain/v5"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSeed_SkipsWhenDataExists(t *testing.T) {
	// Arrange: count > 0 なのでINSERTは実行されない
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	// Act
	err := v5.RunSeed()

	// Assert
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_SeedsWhenEmpty(t *testing.T) {
	// Arrange: count = 0 なのでINSERTが実行される
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO`).
		WillReturnResult(sqlmock.NewResult(3, 3))
	mock.ExpectCommit()

	// Act
	err := v5.RunSeed()

	// Assert
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_ReturnsErrorOnInsertFailure(t *testing.T) {
	// Arrange: INSERTがエラーになる
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO`).
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	// Act
	err := v5.RunSeed()

	// Assert
	assert.Error(t, err)
}
