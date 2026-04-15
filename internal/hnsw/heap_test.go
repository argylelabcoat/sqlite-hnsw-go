package hnsw

import (
	"testing"

	"container/heap"

	"github.com/stretchr/testify/assert"
)

func TestMinHeap_PopReturnsMinimum(t *testing.T) {
	h := &candidateMinHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 1, dist: 3.0})
	heap.Push(h, candidate{id: 2, dist: 1.0})
	heap.Push(h, candidate{id: 3, dist: 2.0})

	assert.Equal(t, 3, h.Len())

	item := heap.Pop(h).(candidate)
	assert.Equal(t, 2, item.id)
	assert.Equal(t, float32(1.0), item.dist)

	item = heap.Pop(h).(candidate)
	assert.Equal(t, 3, item.id)
	assert.Equal(t, float32(2.0), item.dist)

	item = heap.Pop(h).(candidate)
	assert.Equal(t, 1, item.id)
	assert.Equal(t, float32(3.0), item.dist)

	assert.Equal(t, 0, h.Len())
}

func TestMaxHeap_PopReturnsMaximum(t *testing.T) {
	h := &candidateMaxHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 1, dist: 1.0})
	heap.Push(h, candidate{id: 2, dist: 3.0})
	heap.Push(h, candidate{id: 3, dist: 2.0})

	assert.Equal(t, 3, h.Len())

	item := heap.Pop(h).(candidate)
	assert.Equal(t, 2, item.id)
	assert.Equal(t, float32(3.0), item.dist)

	item = heap.Pop(h).(candidate)
	assert.Equal(t, 3, item.id)
	assert.Equal(t, float32(2.0), item.dist)

	item = heap.Pop(h).(candidate)
	assert.Equal(t, 1, item.id)
	assert.Equal(t, float32(1.0), item.dist)

	assert.Equal(t, 0, h.Len())
}

func TestMinHeap_SingleItem(t *testing.T) {
	h := &candidateMinHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 42, dist: 99.0})

	item := heap.Pop(h).(candidate)
	assert.Equal(t, 42, item.id)
	assert.Equal(t, float32(99.0), item.dist)
	assert.Equal(t, 0, h.Len())
}

func TestMaxHeap_SingleItem(t *testing.T) {
	h := &candidateMaxHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 42, dist: 99.0})

	item := heap.Pop(h).(candidate)
	assert.Equal(t, 42, item.id)
	assert.Equal(t, float32(99.0), item.dist)
	assert.Equal(t, 0, h.Len())
}

func TestMinHeap_EqualDistances(t *testing.T) {
	h := &candidateMinHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 1, dist: 1.0})
	heap.Push(h, candidate{id: 2, dist: 1.0})

	item := heap.Pop(h).(candidate)
	assert.Equal(t, float32(1.0), item.dist)
}

func TestMaxHeap_EqualDistances(t *testing.T) {
	h := &candidateMaxHeap{}
	heap.Init(h)

	heap.Push(h, candidate{id: 1, dist: 1.0})
	heap.Push(h, candidate{id: 2, dist: 1.0})

	item := heap.Pop(h).(candidate)
	assert.Equal(t, float32(1.0), item.dist)
}
