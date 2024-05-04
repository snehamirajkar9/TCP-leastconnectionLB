package ratelimiter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRatelimiter_Allow_Burst(t *testing.T) {

	r := NewRateLimiter(10, 1)

	now := time.Now()
	clock := &mockClock{now}

	r.clock = clock
	// Allow for new client should succeed
	assert.True(t, r.Allow("client1"))
	// wait to refill burst
	clock.TimeNow = now.Add(1 * time.Second)

	// Verify rate limiting
	for i := 0; i < 10; i++ {
		assert.True(t, r.Allow("client1"))
	}

	// Should be rate limited now
	assert.False(t, r.Allow("client1"))

	// Wait for bucket to refill -- 2 secs from beginning
	clock.TimeNow = now.Add(2 * time.Second)

	// Should allow again
	assert.True(t, r.Allow("client1"))

	// Different client should not be limited
	assert.True(t, r.Allow("client2"))
}

type mockClock struct {
	TimeNow time.Time
}

func (m *mockClock) Now() time.Time {
	return m.TimeNow
}

func TestRate(t *testing.T) {

	r := NewRateLimiter(10, 5)
	now := time.Now()
	clock := &mockClock{now}

	r.clock = clock

	// pass the burst
	for i := 0; i < 10; i++ {
		assert.True(t, r.Allow("client"))
	}
	// allowed rate
	for i := 1; i < 6; i++ {
		tick := time.Duration(i) * 200 * time.Millisecond
		clock.TimeNow = now.Add(tick)
		assert.True(t, r.Allow("client"))
		fmt.Println(i)
	}

	// Rate limit should kick in
	assert.False(t, r.Allow("client"))

	// Wait for full second
	clock.TimeNow = now.Add(2 * time.Second)

	// Should allow again
	assert.True(t, r.Allow("client"))

}
