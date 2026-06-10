// Package enricher 的 RateLimitHandler 单测。
//
// 覆盖：
//  1. Wait 第一次调用不 sleep
//  2. 第二次调用距离第一次 < minInterval 时 sleep 补齐
//  3. 第二次调用距离第一次 >= minInterval 时不 sleep
//  4. Pause 之后下一次 Wait 会 sleep 到 pausedUntil
//  5. Pause 之后 pausedUntil 之前 → 后续 Wait 不会再 sleep（因为已经过了）
//  6. Reset 清空 pausedUntil
//  7. 并发安全：多个 goroutine 并发调 Wait 总时间 ≈ (N-1) * minInterval
package enricher

import (
	"sync"
	"testing"
	"time"
)

// TestRateLimit_FirstCallNoSleep 第一次调 Wait 不应 sleep（lastReq 是 zero time）。
func TestRateLimit_FirstCallNoSleep(t *testing.T) {
	rl := NewRateLimitHandler(100 * time.Millisecond)
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("first Wait should be near-instant, got %v", elapsed)
	}
}

// TestRateLimit_SecondCallSleepToInterval 验证最小间隔。
func TestRateLimit_SecondCallSleepToInterval(t *testing.T) {
	minInterval := 200 * time.Millisecond
	rl := NewRateLimitHandler(minInterval)

	rl.Wait() // 第一次

	start := time.Now()
	rl.Wait() // 第二次应 sleep ~minInterval
	elapsed := time.Since(start)

	// 允许一些时钟误差
	if elapsed < minInterval-50*time.Millisecond {
		t.Errorf("second Wait should sleep at least %v, got %v", minInterval, elapsed)
	}
	if elapsed > minInterval+200*time.Millisecond {
		t.Errorf("second Wait should not oversleep too much, got %v", elapsed)
	}
}

// TestRateLimit_ElapsedBeyondInterval 距离上次够远时,不 sleep。
func TestRateLimit_ElapsedBeyondInterval(t *testing.T) {
	minInterval := 50 * time.Millisecond
	rl := NewRateLimitHandler(minInterval)

	rl.Wait()
	time.Sleep(100 * time.Millisecond) // 远大于 minInterval

	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Wait should be near-instant when elapsed > minInterval, got %v", elapsed)
	}
}

// TestRateLimit_Pause 验证 Pause 期间 Wait 会 sleep 到 pausedUntil。
func TestRateLimit_Pause(t *testing.T) {
	rl := NewRateLimitHandler(10 * time.Millisecond)
	pauseUntil := time.Now().Add(150 * time.Millisecond)
	rl.Pause(pauseUntil)

	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("Wait should sleep until pausedUntil, got %v (want ~150ms)", elapsed)
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("Wait should not oversleep too much, got %v", elapsed)
	}
}

// TestRateLimit_PauseAlreadyPassed 验证 pausedUntil 早过期时,Wait 立即返回。
func TestRateLimit_PauseAlreadyPassed(t *testing.T) {
	rl := NewRateLimitHandler(10 * time.Millisecond)
	rl.Pause(time.Now().Add(-1 * time.Second)) // 已过期

	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Wait should not sleep when pausedUntil passed, got %v", elapsed)
	}
}

// TestRateLimit_Reset 验证 Reset 清空 pausedUntil。
func TestRateLimit_Reset(t *testing.T) {
	rl := NewRateLimitHandler(10 * time.Millisecond)
	rl.Pause(time.Now().Add(200 * time.Millisecond))
	rl.Reset()

	// Reset 之后立即 Wait 应当几乎不 sleep
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		t.Errorf("after Reset, Wait should not sleep, got %v", elapsed)
	}
}

// TestRateLimit_ConcurrentSafety 验证并发安全:N 个 worker 并发 Wait 总时间 ≈ (N-1) * minInterval。
//
// 这是漏桶语义的体现：所有调用串行化排队，总耗时受 minInterval 限制。
func TestRateLimit_ConcurrentSafety(t *testing.T) {
	const workerCnt = 5
	const minInterval = 50 * time.Millisecond
	rl := NewRateLimitHandler(minInterval)

	var wg sync.WaitGroup
	wg.Add(workerCnt)

	start := time.Now()
	for i := 0; i < workerCnt; i++ {
		go func() {
			defer wg.Done()
			rl.Wait()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	// N 次 Wait 串行化,总耗时 ≈ (N-1) * minInterval
	expectedMin := time.Duration(workerCnt-1) * minInterval
	if elapsed < expectedMin-50*time.Millisecond {
		t.Errorf("concurrent Wait should take at least %v, got %v", expectedMin, elapsed)
	}
	// 上限放宽到 2x,因为 goroutine 调度有抖动
	if elapsed > 2*expectedMin+200*time.Millisecond {
		t.Errorf("concurrent Wait took too long: %v (expected ~%v)", elapsed, expectedMin)
	}
}
