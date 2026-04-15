package sqlitehnsw

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeVector_RoundTrip(t *testing.T) {
	original := []float32{1.0, -2.5, 0.0, 3.14, math.MaxFloat32}
	encoded := encodeVector(original)
	assert.Len(t, encoded, len(original)*4)

	decoded := decodeVector(encoded)
	assert.InDeltaSlice(t, original, decoded, 1e-6)
}

func TestEncodeVector_Length(t *testing.T) {
	vec := make([]float32, 384)
	encoded := encodeVector(vec)
	assert.Len(t, encoded, 384*4, "384-dim vector should be 1536 bytes")
}

func TestDecodeVector_InvalidLength_ReturnsNil(t *testing.T) {
	result := decodeVector([]byte{0x01, 0x02, 0x03})
	assert.Nil(t, result, "non-multiple-of-4 bytes should return nil")
}

func TestEncodeDecodeVector_Empty(t *testing.T) {
	encoded := encodeVector([]float32{})
	require.Empty(t, encoded)
	decoded := decodeVector(encoded)
	assert.Empty(t, decoded)
}
