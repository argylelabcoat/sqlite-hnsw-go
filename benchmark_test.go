package sqlitehnsw

import (
	"math"
	"math/rand"
	"testing"
)

func BenchmarkInsert(b *testing.B) {
	s := benchStore(b, 384)
	defer s.Close()

	rng := rand.New(rand.NewSource(42))
	vecs := benchRandomVectors(rng, b.N, 384)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Upsert([]Point{{ID: i + 1, Vector: vecs[i]}})
	}
}

func BenchmarkSearch(b *testing.B) {
	s := benchStore(b, 384)
	defer s.Close()

	rng := rand.New(rand.NewSource(42))
	const n = 1000
	vecs := benchRandomVectors(rng, n, 384)
	for i := 0; i < n; i++ {
		s.Upsert([]Point{{ID: i + 1, Vector: vecs[i]}})
	}

	query := benchRandomVectors(rng, 1, 384)[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Search(query, 10)
	}
}

func BenchmarkSerialize(b *testing.B) {
	s := benchStore(b, 384)
	defer s.Close()

	rng := rand.New(rand.NewSource(42))
	const n = 1000
	vecs := benchRandomVectors(rng, n, 384)
	for i := 0; i < n; i++ {
		s.Upsert([]Point{{ID: i + 1, Vector: vecs[i]}})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.FlushGraph()
	}
}

func benchStore(b *testing.B, dim int) *Store {
	b.Helper()
	path := b.TempDir() + "/bench.db"
	s, err := NewStore(Config{DBPath: path, Dimension: dim, M: 16, EfConstruction: 200, EfSearch: 64})
	if err != nil {
		b.Fatal(err)
	}
	return s
}

func benchRandomVectors(rng *rand.Rand, n, dim int) [][]float32 {
	vecs := make([][]float32, n)
	for i := 0; i < n; i++ {
		v := make([]float32, dim)
		for j := 0; j < dim; j++ {
			v[j] = float32(rng.NormFloat64())
		}
		norm := float32(0)
		for _, x := range v {
			norm += x * x
		}
		norm = float32(math.Sqrt(float64(norm)))
		for j := range v {
			v[j] /= norm
		}
		vecs[i] = v
	}
	return vecs
}
