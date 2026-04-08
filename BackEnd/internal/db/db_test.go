package db

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInitMongo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := InitMongo(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, DB)
}
