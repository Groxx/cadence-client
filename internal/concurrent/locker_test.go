// Copyright (c) 2017-2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package concurrent

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocker(t *testing.T) {
	t.Run("basic lock", func(t *testing.T) {
		l := NewLocker(5)
		m, unlock := l.Lock()
		defer unlock()
		assert.Equal(t, *m, 5)
		*m = 3
		assert.Equal(t, *m, 3)
	})
	t.Run("basic change", func(t *testing.T) {
		l := NewLocker(-1)
		var wg sync.WaitGroup
		wg.Add(100)
		for i := 0; i < 100; i++ {
			i := i
			go func() {
				l.Change(func(value *int) {
					*value = i
				})
				wg.Done()
			}()
		}
		wg.Wait()
		assert.NotEqual(t, l.Get(), -1)
	})
}

func TestBenchmark(t *testing.T) {
	// both serial and parallel tests are essentially identical,
	// but they can't reasonably share code because it's a benchmark.
	//
	// this is set up as a test rather than a benchmark because it allows
	// asserting on the results.  they only take about 10 seconds to run.
	// if it turns out to be flaky, increase epsilons or simply skip this
	// in CI.
	//
	// I would set this up as a `func BenchmarkSomething`, but it deadlocks
	// when using `testing.Benchmark` internally for some reason.

	// results:
	// - concurrent speed is marginally worse, which makes sense since it's running
	//   more functions -> more opportunities to be preempted?  it's very small though.
	// - serial speed is basically identical, and they frequently trade which one is
	//   the fastest.
	// - running with the race detector makes the differences clearer, but are still
	//   probably acceptable.
	//
	// so all in all pretty good!  it very likely optimizes away to identical code
	// (or close enough) in most cases, since generics inline pretty well.

	/*
		On an M1 mac:
		primitive: 10036098       114.2 ns/op        0 B/op       0 allocs/op
		lock:       9842462       120.3 ns/op        0 B/op       0 allocs/op
		change:     9509110       128.3 ns/op        0 B/op       0 allocs/op

		With race:
		primitive:  2228901       510.3 ns/op        0 B/op       0 allocs/op
		lock:       2193153       554.4 ns/op        0 B/op       0 allocs/op
		change:     2078850       567.5 ns/op        0 B/op       0 allocs/op
	*/
	t.Run("parallel", func(t *testing.T) {
		primitive := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			var mut sync.Mutex
			val := -1
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					mut.Lock()
					val = i
					mut.Unlock()
					i++
				}
			})
			assert.NotEqual(b, val, -1)
		})
		lock := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			l := NewLocker(-1)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					val, unlock := l.Lock()
					*val = i
					unlock()
					i++
				}
			})
			assert.NotEqual(b, l.Get(), -1)
		})
		change := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			l := NewLocker(-1)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					l.Change(func(value *int) {
						*value = i
					})
					i++
				}
			})
			assert.NotEqual(b, l.Get(), -1)
		})
		t.Log("primitive:", primitive.String(), primitive.MemString())
		t.Log("lock:     ", lock.String(), lock.MemString())
		t.Log("change:   ", change.String(), change.MemString())
		require.NotZero(t, primitive.NsPerOp()) // sanity check
		assert.InEpsilon(t, primitive.NsPerOp(), lock.NsPerOp(), 1.1, "Lock should be within 10% of a primitive mutex")
		assert.InEpsilon(t, primitive.NsPerOp(), change.NsPerOp(), 1.1, "Change should be within 10% of a primitive mutex")
	})
	/*
		On an M1 mac:
		primitive: 177453584         6.782 ns/op        0 B/op       0 allocs/op
		lock:      176782773         6.818 ns/op        0 B/op       0 allocs/op
		change:    173444262         6.854 ns/op        0 B/op       0 allocs/op

		With race:
		primitive: 13319624        94.00 ns/op       0 B/op       0 allocs/op
		lock:      10605433       106.0 ns/op        0 B/op       0 allocs/op
		change:     8294510       138.7 ns/op        0 B/op       0 allocs/op
	*/
	t.Run("serial", func(t *testing.T) {
		primitive := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			var mut sync.Mutex
			val := -1
			for i := 0; i < b.N; i++ {
				mut.Lock()
				val = i
				mut.Unlock()
				i++
			}
			assert.NotEqual(b, val, -1)
		})
		lock := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			l := NewLocker(-1)
			for i := 0; i < b.N; i++ {
				val, unlock := l.Lock()
				*val = i
				unlock()
				i++
			}
			assert.NotEqual(b, l.Get(), -1)
		})
		change := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			l := NewLocker(-1)
			for i := 0; i < b.N; i++ {
				l.Change(func(value *int) {
					*value = i
				})
				i++
			}
			assert.NotEqual(b, l.Get(), -1)
		})
		t.Log("primitive:", primitive.String(), primitive.MemString())
		t.Log("lock:     ", lock.String(), lock.MemString())
		t.Log("change:   ", change.String(), change.MemString())
		require.NotZero(t, primitive.NsPerOp()) // sanity check

		// bm.NsPerOp is an integer, which is not precise enough for these single-digit-ns tests.
		// so compute the higher precision value it prints instead.
		nsop := func(bm testing.BenchmarkResult) float64 {
			return float64(bm.T) / float64(bm.N)
		}
		assert.InEpsilon(t, nsop(primitive), nsop(lock), 1.1, "Lock should be within 10% of a primitive mutex")
		assert.InEpsilon(t, nsop(primitive), nsop(change), 1.1, "Change should be within 10% of a primitive mutex")
	})
}
