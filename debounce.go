package debounce

import (
	"context"
	"sync/atomic"
	"time"
)

type Debouncer interface {
	Invoke()
	Cancel()
	Flush()
}

type debouncer struct {
	// user inputs
	f    func()
	wait time.Duration
	Options

	// wait trigger for Trailing use-case
	waitTicker *time.Ticker

	// atomic guard
	i uint32

	// support for Cancel() and maxWait via context.Timeout
	ctx        context.Context
	cancelFunc context.CancelFunc

	// handling for Leading + Trailing use-case
	intermediateCount int64
}

type Options struct {
	// Trigger on Leading edge
	Leading bool
	// Support for periodic trigger / Throttling
	MaxWait time.Duration
	// Trigger on Trailing edge
	Trailing bool
}

// Invoke conditionally calls f()
//
// If Leading is true and Invoke has not been called since wait, f() is called
// If Trailing is true wait is defered for options.Wait period
func (d *debouncer) Invoke() {
	// defer
	d.waitTicker.Reset(d.wait)

	// special handling for Leading + Trailing
	d.intermediateCount++

	// if ungaurded launch background listener
	if atomic.CompareAndSwapUint32(&d.i, 0, 1) {
		d.intermediateCount = 0
		// launching goroutine
		// goroutine exits when wait or maxwait or flush or cancel occur
		go func() {
			// unguard on exit
			defer atomic.StoreUint32(&d.i, 0)

			if d.Leading {
				d.f()
			}

			select {
			// timer tiggered
			case <-d.waitTicker.C:
				// Leading + Trailing requires intermediate Invoke()
				if d.Leading && d.Trailing {
					if d.intermediateCount > 0 {
						d.f()
					}
				} else if d.Trailing {
					d.f()
				}

			// maxwait expired or cancel
			//
			// for throttle use case, f() is called on subsequent Invoke()
			//
			// use of Trailing + maxWait is handled above in the waitTicker handler
			case <-d.ctx.Done():
			}
		}()
	}
}

// Will never call f()
func (d *debouncer) Cancel() {
	d.waitTicker.Stop()
	d.f = func() {} // load noop
	d.cancelFunc()
}

// Will call f() if an Invoke() is queued
func (d *debouncer) Flush() {
	d.waitTicker.Reset(0)
}

// Creates and returns a new Debouncer which will
// postpone its execution until after wait milliseconds have elapsed since the last time it
// was invoked. Useful for implementing behavior that should only happen after the input has
// stopped arriving.
func New(f func(), wait time.Duration, options Options) Debouncer {
	ctx, cancelFunc := context.WithCancel(context.Background())

	if options.MaxWait > 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, options.MaxWait)
	}

	waitTicker := time.NewTicker(wait)
	d := &debouncer{f, wait, options, waitTicker, 0, ctx, cancelFunc, 0}

	return d
}

// Creates and returns a new, throttled Debouncer, that, when invoked
// repeatedly, will only actually call the original function at most once per every wait period.
// Useful for rate-limiting events that occur faster than you can keep up with.
func Throttle(f func(), minInterval time.Duration) Debouncer {
	return New(f, 24*time.Hour, Options{Leading: true, Trailing: false, MaxWait: minInterval})
}
