package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.InsertAll([]*int{&[]int{1}[0], &[]int{2}[0], &[]int{3}[0]})
	assert.Equal(t, rb.All(), []*int{&[]int{1}[0], &[]int{2}[0], &[]int{3}[0]})
	rb.Insert(&[]int{4}[0])
	assert.Equal(t, rb.All(), []*int{&[]int{2}[0], &[]int{3}[0], &[]int{4}[0]})
	rb.Insert(&[]int{5}[0])
	assert.Equal(t, rb.All(), []*int{&[]int{3}[0], &[]int{4}[0], &[]int{5}[0]})
	rb.Insert(&[]int{6}[0])
	assert.Equal(t, rb.All(), []*int{&[]int{4}[0], &[]int{5}[0], &[]int{6}[0]})
	rb.Insert(&[]int{7}[0])
	assert.Equal(t, rb.All(), []*int{&[]int{5}[0], &[]int{6}[0], &[]int{7}[0]})
}

func TestRingBufferCurrent(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.InsertAll([]*int{&[]int{1}[0], &[]int{2}[0], &[]int{3}[0]})
	assert.Equal(t, rb.Current(), &[]int{1}[0])
	rb.Next()
	assert.Equal(t, rb.Current(), &[]int{2}[0])
	rb.Next()
	assert.Equal(t, rb.Current(), &[]int{3}[0])
	rb.Next()
	assert.Equal(t, rb.Current(), &[]int{1}[0])
	rb.Next()
	assert.Equal(t, rb.Current(), &[]int{2}[0])
}

func TestRingBufferConcurrent(t *testing.T) {
	rb := NewRingBuffer[int](3)
	doneChan := make(chan bool, 3)

	go func() {
		for i := 0; i < 100000; i++ {
			rb.Insert(&[]int{i}[0])
		}
		doneChan <- true
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			rb.Next()
		}
		doneChan <- true
	}()

	go func() {
		for i := 0; i < 100000; i++ {
			rb.Current()
		}
		doneChan <- true
	}()

	for i := 0; i < 3; i++ {
		<-doneChan
	}
}
