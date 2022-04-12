// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rate

import (
	"sync/atomic"
	"testing"
	"time"

	lockfree "github.com/bpowers/lockfree-rate"
	"go.uber.org/ratelimit"
	golang "golang.org/x/time/rate"
)

const (
	rate  = 100000
	burst = 1
)

func maybeReport(t testing.TB, d time.Duration, ok, total uint64) {
	if d > 1*time.Second {
		allowedRPS := float64(ok) / (float64(d) / float64(time.Second))
		overallRPS := float64(total) / (float64(d) / float64(time.Second))
		percent := 100 * float64(ok) / float64(total)
		t.Logf("%.1f RPS allowed, %.1f RPS overall. %.3f%% of %d requests in %v", allowedRPS, overallRPS, percent, total, d)
	}
}

func BenchmarkLockfree(b *testing.B) {
	var total = uint64(0)
	var numOK = uint64(0)

	lim := lockfree.NewLimiter(rate, burst)

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		localOK := uint64(0)
		localTotal := uint64(0)

		for pb.Next() {
			if lim.Allow() {
				localOK++
			}
			localTotal++
		}

		atomic.AddUint64(&numOK, localOK)
		atomic.AddUint64(&total, localTotal)
	})

	maybeReport(b, time.Now().Sub(start), atomic.LoadUint64(&numOK), atomic.LoadUint64(&total))
}

func BenchmarkGolang(b *testing.B) {
	var total = uint64(0)
	var numOK = uint64(0)

	lim := golang.NewLimiter(rate, burst)

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		localOK := uint64(0)
		localTotal := uint64(0)

		for pb.Next() {
			if lim.Allow() {
				localOK++
			}
			localTotal++
		}

		atomic.AddUint64(&numOK, localOK)
		atomic.AddUint64(&total, localTotal)
	})

	maybeReport(b, time.Now().Sub(start), atomic.LoadUint64(&numOK), atomic.LoadUint64(&total))
}

func BenchmarkUber(b *testing.B) {
	var total = uint64(0)
	var numOK = uint64(0)

	lim := ratelimit.New(rate, ratelimit.WithSlack(burst))

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		localOK := uint64(0)
		localTotal := uint64(0)

		for pb.Next() {
			_ = lim.Take()
			localOK++
			localTotal++
		}

		atomic.AddUint64(&numOK, localOK)
		atomic.AddUint64(&total, localTotal)
	})

	maybeReport(b, time.Now().Sub(start), atomic.LoadUint64(&numOK), atomic.LoadUint64(&total))
}

// global var so the comparison below can't be optimized away
var Then int64

// this tests the absolute minimum time a call to Accept() (which internally
// calls time.Now) could take.
func BenchmarkTimeNow(b *testing.B) {
	var total = uint64(0)
	var numOK = uint64(0)

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()
	b.RunParallel(func(pb *testing.PB) {
		localOK := uint64(0)
		localTotal := uint64(0)

		for pb.Next() {
			if time.Now().UnixMicro() > Then {
				localOK++
			}
			localTotal++
		}

		atomic.AddUint64(&numOK, localOK)
		atomic.AddUint64(&total, localTotal)
	})

	maybeReport(b, time.Now().Sub(start), atomic.LoadUint64(&numOK), atomic.LoadUint64(&total))
}
