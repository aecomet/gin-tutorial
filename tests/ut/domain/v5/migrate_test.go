package v5_test

import (
	"testing"

	v5 "gin-tutorial/app/domain/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations_PropagatesError(t *testing.T) {
	// Arrange: VERSION()消費後はExpectationがないため、
	// AutoMigrateが発行するクエリはsqlmockに弾かれてエラーになる
	setupV5MockDB(t)

	// Act
	err := v5.RunMigrations()

	// Assert: migration failed: のプレフィックスで包まれて返ること
	require.Error(t, err)
	assert.ErrorContains(t, err, "migration failed")
}
