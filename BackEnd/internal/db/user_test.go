package db

import (
	"CommentClassifier/internal/data"
	"CommentClassifier/internal/utils"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	user = data.User{
		Username: "admin",
		Password: "123456",
		Role:     utils.RoleAdmin,
	}
)

func setup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := InitMongo(ctx)
	if err != nil {
		panic(err)
	}
}

func TestRegisterUser(t *testing.T) {
	setup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := RegisterUser(ctx, user.Username, user.Password, user.Role)
	assert.NoError(t, err)
}

func TestFindUserByUsername(t *testing.T) {
	setup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u, err := FindUserByUsername(ctx, user.Username)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, user.Username, u.Username)
}

func TestCountUsers(t *testing.T) {
	setup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := CountUsers(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
