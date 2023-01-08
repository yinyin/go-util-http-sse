package httpsse

import (
	"math/rand"
	"testing"
)

func BenchmarkLoop2(b *testing.B) {
	d := make([]byte, 1024)
	t := make([]byte, 512)
	for idx := 0; idx < 1024; idx++ {
		d[idx] = byte(rand.Int31n(0x0100))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := t[:0]
		for _, c := range d {
			if c == '\n' {
				continue
			}
			u = append(u, c)
		}
		_ = u
	}
}

func BenchmarkLoop3(b *testing.B) {
	d := make([]byte, 1024)
	t := make([]byte, 512)
	for idx := 0; idx < 1024; idx++ {
		d[idx] = byte(rand.Int31n(0x0100))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := t
		targetIndex := 0
		for _, c := range d {
			if c == '\n' {
				continue
			}
			u[targetIndex] = c
			targetIndex++
			if targetIndex >= len(u) {
				u2 := make([]byte, len(u)+512)
				copy(u2, u)
				u = u2
			}
		}
		u = u[:targetIndex]
		_ = u
	}
}

func TestPrepareEventDataBuffer1(t *testing.T) {
	result := prepareEventDataBuffer([]byte("id: 123\n"), []byte("123"))
	if txt := string(result); txt != "id: 123\ndata: 123\n\n" {
		t.Errorf("unexpect result: [%s]", txt)
	}
}

func TestPrepareEventDataBuffer2(t *testing.T) {
	result := prepareEventDataBuffer([]byte("id: 123\n"), []byte("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"))
	if txt := string(result); txt != "id: 123\ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \ndata: \n\n" {
		t.Errorf("unexpect result: [%s]", txt)
	}
}
