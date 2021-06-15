package debounce_test

import (
	"testing"
	"time"

	"github.com/goodfoo/debounce"
)

func TestNotInkvokedDoesNothingEvenWithFlush(t *testing.T) {
	f := func() {
		t.Fail()
	}

	debounce := debounce.New(f, time.Millisecond, debounce.Options{Leading: true, Trailing: true})
	time.Sleep(5 * time.Millisecond)
	debounce.Flush()
}

func TestDebounceRunsOnce(t *testing.T) {
	i := 0

	f := func() {
		i++
	}

	debounce := debounce.New(f, time.Millisecond, debounce.Options{Trailing: true})

	debounce.Invoke()
	time.Sleep(5 * time.Millisecond)

	debounce.Flush()
	debounce.Cancel()

	if i != 1 {
		t.Errorf("Expected 1, got %d", i)
	}
}

func TestDebounceTrailingRunsOnceForWaitPeriod(t *testing.T) {
	i := 0

	f := func() {
		i++
	}

	debounce := debounce.New(f, 2*time.Millisecond, debounce.Options{Trailing: true})

	debounce.Invoke()
	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(5 * time.Millisecond)

	if i != 1 {
		t.Errorf("Expected 1, got %d", i)
	}
}

func TestDebounceCancelPreventsInvoke(t *testing.T) {
	f := func() {
		t.Fail()
	}

	debounce := debounce.New(f, time.Second, debounce.Options{Trailing: true})

	debounce.Invoke()
	debounce.Cancel()
}

func TestMaxWaitSansTrailing(t *testing.T) {
	i := 0
	f := func() {
		i++
	}

	debounce := debounce.New(f, time.Second, debounce.Options{Leading: true, Trailing: false, MaxWait: 2 * time.Millisecond})

	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(3 * time.Millisecond)
	debounce.Invoke()
	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(3 * time.Millisecond)
	debounce.Cancel()

	if i != 2 {
		t.Errorf("Expected 2, got %d", i)
	}
}

func TestMaxwaitWithTrailing(t *testing.T) {
	i := 0
	f := func() {
		i++
	}

	debounce := debounce.New(f, time.Second, debounce.Options{Leading: true, Trailing: true, MaxWait: 5 * time.Millisecond})

	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(2 * time.Millisecond)
	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(2 * time.Millisecond)
	debounce.Invoke()
	debounce.Invoke()
	debounce.Flush()
	time.Sleep(2 * time.Millisecond)

	if i != 2 {
		t.Errorf("Expected 2, got %d", i)
	}
}

func TestThrottle(t *testing.T) {
	i := 0
	f := func() {
		i++
	}

	debounce := debounce.Throttle(f, 2*time.Millisecond)

	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(3 * time.Millisecond)
	debounce.Invoke()
	debounce.Invoke()
	debounce.Invoke()
	time.Sleep(3 * time.Millisecond)

	if i != 2 {
		t.Errorf("Expected 2, got %d", i)
	}
}
