package handler_test

import (
	"testing"

	"gin-tutorial/app/handler"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error_ReturnsMessage(t *testing.T) {
	assert.Equal(t, "resource not found", handler.ErrNotFound.Error())
	assert.Equal(t, "authentication required", handler.ErrUnauthorized.Error())
	assert.Equal(t, "invalid request", handler.ErrBadRequest.Error())
}

func TestErrNotFound_Fields(t *testing.T) {
	assert.Equal(t, 404, handler.ErrNotFound.Status)
	assert.Equal(t, "NOT_FOUND", handler.ErrNotFound.Code)
	assert.Equal(t, "resource not found", handler.ErrNotFound.Message)
}

func TestErrUnauthorized_Fields(t *testing.T) {
	assert.Equal(t, 401, handler.ErrUnauthorized.Status)
	assert.Equal(t, "UNAUTHORIZED", handler.ErrUnauthorized.Code)
	assert.Equal(t, "authentication required", handler.ErrUnauthorized.Message)
}

func TestErrBadRequest_Fields(t *testing.T) {
	assert.Equal(t, 400, handler.ErrBadRequest.Status)
	assert.Equal(t, "BAD_REQUEST", handler.ErrBadRequest.Code)
	assert.Equal(t, "invalid request", handler.ErrBadRequest.Message)
}
