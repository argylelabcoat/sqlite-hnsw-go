package hnsw

import "container/heap"

type candidate struct {
	id   int
	dist float32
}

type candidateMinHeap []candidate

func (h candidateMinHeap) Len() int           { return len(h) }
func (h candidateMinHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h candidateMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *candidateMinHeap) Push(x any)        { *h = append(*h, x.(candidate)) }
func (h *candidateMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	if h.Len() > 0 {
		heap.Fix(h, 0)
	}
	return x
}

type candidateMaxHeap []candidate

func (h candidateMaxHeap) Len() int           { return len(h) }
func (h candidateMaxHeap) Less(i, j int) bool { return h[i].dist > h[j].dist }
func (h candidateMaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *candidateMaxHeap) Push(x any)        { *h = append(*h, x.(candidate)) }
func (h *candidateMaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	if h.Len() > 0 {
		heap.Fix(h, 0)
	}
	return x
}
