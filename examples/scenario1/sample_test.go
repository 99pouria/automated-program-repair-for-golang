package mymath

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipleSum(t *testing.T) {
	require.Equal(t, 55, MultipleSum(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
}
