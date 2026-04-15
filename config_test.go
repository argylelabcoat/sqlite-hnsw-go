package sqlitehnsw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DefaultsApplied(t *testing.T) {
	cfg := Config{DBPath: "test.db", Dimension: 384}
	cfg.applyDefaults()
	assert.Equal(t, "default", cfg.Collection)
	assert.Equal(t, 16, cfg.M)
	assert.Equal(t, 200, cfg.EfConstruction)
	assert.Equal(t, 64, cfg.EfSearch)
	assert.Equal(t, 1000, cfg.FlushThreshold)
	assert.Equal(t, 10240, cfg.CacheSize)
	assert.Equal(t, Cosine, cfg.Metric)
}

func TestConfig_CustomValuesPreserved(t *testing.T) {
	cfg := Config{
		DBPath:         "test.db",
		Dimension:      384,
		Collection:     "mycollection",
		Metric:         Euclidean,
		M:              32,
		EfConstruction: 400,
		EfSearch:       128,
		FlushThreshold: 500,
		CacheSize:      2048,
	}
	cfg.applyDefaults()
	assert.Equal(t, "test.db", cfg.DBPath)
	assert.Equal(t, 384, cfg.Dimension)
	assert.Equal(t, "mycollection", cfg.Collection)
	assert.Equal(t, Euclidean, cfg.Metric)
	assert.Equal(t, 32, cfg.M)
	assert.Equal(t, 400, cfg.EfConstruction)
	assert.Equal(t, 128, cfg.EfSearch)
	assert.Equal(t, 500, cfg.FlushThreshold)
	assert.Equal(t, 2048, cfg.CacheSize)
}
