package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempToSpeed(t *testing.T) {
	a, b := tempToSpeed(0)
	assert.True(t, a)
	assert.Equal(t, 5, int(b))

	a, b = tempToSpeed(90)
	assert.False(t, a)
	assert.Equal(t, 100, int(b))
}

func TestBoundedTempisOk(t *testing.T) {
	for i := -5000; i < 5000; i++ {
		_, speed := tempToSpeed(float64(i))
		assert.True(t, speed <= 100)
		assert.True(t, speed >= 0)
	}
}
