package db_test

import (
	"errors"
	"testing"

	"gin-tutorial/app/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mysqldriver "gorm.io/driver/mysql"
)

func TestBuildDSN_DefaultValues(t *testing.T) {
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(key, "")
	}
	dsn := db.BuildDSN()
	assert.Equal(t, "root:root@tcp(localhost:3306)/gin_tutorial?charset=utf8mb4&parseTime=True&loc=Local", dsn)
}

func TestBuildDSN_CustomEnvVars(t *testing.T) {
	t.Setenv("DB_HOST", "myhost")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "myuser")
	t.Setenv("DB_PASSWORD", "mypass")
	t.Setenv("DB_NAME", "mydb")
	dsn := db.BuildDSN()
	assert.Equal(t, "myuser:mypass@tcp(myhost:3307)/mydb?charset=utf8mb4&parseTime=True&loc=Local", dsn)
}

func TestInitWithDialector_Success(t *testing.T) {
	// Arrange
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
		db.DB = nil
	})
	mock.ExpectQuery("SELECT VERSION()").
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))

	dialector := mysqldriver.New(mysqldriver.Config{Conn: sqlDB})

	// Act
	err = db.InitWithDialector(dialector)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, db.DB)
}

func TestInitWithDialector_ReturnsErrorOnConnectionFailure(t *testing.T) {
	// Arrange: VERSION()クエリが失敗 → GORM初期化エラー
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	mock.ExpectQuery("SELECT VERSION()").
		WillReturnError(errors.New("connection refused"))

	dialector := mysqldriver.New(mysqldriver.Config{Conn: sqlDB})

	// Act
	err = db.InitWithDialector(dialector)

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to connect to database")
}
