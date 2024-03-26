package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	hash := PassHash("pass#$%word")

	assert.Equal(t, "59e5bf62c83b5b521dc91f5baff2eb17215b43e877739571df4cd80f1b3d9d29", hash)
}
