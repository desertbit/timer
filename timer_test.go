package timer

import (
	"sync"
	"testing"
	"time"
)

func TestNullTimout(t *testing.T) {
	// Timeout for 0 seconds.
	start := time.Now()
	timer := NewTimer(0)
	<-timer.C
	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestNegativeTimout(t *testing.T) {
	// Timeout for -1 seconds.
	start := time.Now()
	timer := NewTimer(-1)
	<-timer.C
	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}

	// Timeout for -100 seconds.
	start = time.Now()
	timer = NewTimer(-100 * time.Second)
	<-timer.C
	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestOverflowTimout(t *testing.T) {
	timer := NewTimer(1<<63 - 1)
	defer timer.Stop()
	if timer.when != 1<<63-1 {
		t.Errorf("overflow: invalid max value")
	}
}

func TestTimeValue(t *testing.T) {
	// Timeout for 0 seconds.
	start := time.Now()
	timer := NewTimer(time.Second)
	v := <-timer.C
	if diff := v.Sub(start).Seconds(); int(diff) != 1 {
		t.Errorf("invalid time value: %v", int(diff))
	}
}

func TestSingleTimout(t *testing.T) {
	// Timeout for 1 second and wait.
	start := time.Now()
	timer := NewTimer(time.Second)
	<-timer.C
	if int(time.Since(start).Seconds()) != 1 {
		t.Errorf("took ~%v seconds, should be ~1 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestMultipleTimouts(t *testing.T) {
	start := time.Now()
	var timers []*Timer

	for i := 0; i < 1000; i++ {
		timers = append(timers, NewTimer(time.Second))
	}

	// Wait for them all to expire.
	for _, timer := range timers {
		<-timer.C
	}

	if int(time.Since(start).Seconds()) != 1 {
		t.Errorf("took ~%v seconds, should be ~1 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestMultipleDifferentTimouts(t *testing.T) {
	start := time.Now()
	var timers []*Timer

	for i := 0; i < 1000; i++ {
		timers = append(timers, NewTimer(time.Duration(i%4)*time.Second))
	}

	// Wait for them all to expire.
	for _, timer := range timers {
		<-timer.C
	}

	if int(time.Since(start).Seconds()) != 3 {
		t.Errorf("took ~%v seconds, should be ~3 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestStoppedTimer(t *testing.T) {
	timer := NewStoppedTimer()
	if timer.when != 0 {
		t.Errorf("invalid stopped timer when value")
	}

	start := time.Now()
	wasActive := timer.Reset(time.Second)
	if wasActive {
		t.Errorf("stopped timer: was active is true")
	}

	<-timer.C
	if int(time.Since(start).Seconds()) != 1 {
		t.Errorf("took ~%v seconds, should be ~1 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestStop(t *testing.T) {
	timer := NewTimer(time.Second)
	wasActive := timer.Stop()
	if !wasActive {
		t.Errorf("stop timer: was active is false")
	}

	select {
	case <-timer.C:
		t.Errorf("failed to stop timer")
	case <-time.After(2 * time.Second):
	}

	wasActive = timer.Stop()
	if wasActive {
		t.Errorf("stop timer: was active is true")
	}
}

func TestStopPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || r.(string) != "timer: Stop called on uninitialized Timer" {
			t.Errorf("stop timer: invalid stop panic")
		}
	}()

	timer := &Timer{}
	timer.Stop()
}

func TestMultipleStop(t *testing.T) {
	var timers []*Timer

	for i := 0; i < 1000; i++ {
		timer := NewTimer(time.Second)
		wasActive := timer.Stop()
		if !wasActive {
			t.Errorf("stop timer: was active is false")
		}

		timers = append(timers, timer)
	}

	time.Sleep(2 * time.Second)

	// All channels must block.
	for _, timer := range timers {
		select {
		case <-timer.C:
			t.Errorf("failed to stop timer")
		default:
		}
	}

	for _, timer := range timers {
		wasActive := timer.Stop()
		if wasActive {
			t.Errorf("stop timer: was active is true")
		}
	}
}

func TestReset(t *testing.T) {
	start := time.Now()
	timer := NewTimer(time.Second)
	wasActive := timer.Reset(2 * time.Second)
	if !wasActive {
		t.Errorf("reset timer: was active is false")
	}

	<-timer.C

	if int(time.Since(start).Seconds()) != 2 {
		t.Errorf("took ~%v seconds, should be ~2 seconds\n", int(time.Since(start).Seconds()))
	}

	start = time.Now()
	wasActive = timer.Reset(time.Second)
	if wasActive {
		t.Errorf("reset timer: was active is true")
	}

	<-timer.C

	if int(time.Since(start).Seconds()) != 1 {
		t.Errorf("took ~%v seconds, should be ~1 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestNegativeReset(t *testing.T) {
	// Timeout for -1 seconds.
	start := time.Now()
	timer := NewTimer(time.Second)
	timer.Reset(-1)
	<-timer.C
	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}

	// Timeout for -100 seconds.
	start = time.Now()
	timer = NewTimer(time.Second)
	timer.Reset(-100 * time.Second)
	<-timer.C
	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestOverflowReset(t *testing.T) {
	timer := NewTimer(time.Second)
	timer.Reset(1<<63 - 1)
	defer timer.Stop()
	if timer.when != 1<<63-1 {
		t.Errorf("overflow: invalid max value")
	}
}

func TestMultipleResets(t *testing.T) {
	start := time.Now()
	var timers []*Timer

	for i := 0; i < 1000; i++ {
		timer := NewTimer(time.Second)
		timers = append(timers, timer)
		timer.Reset(2 * time.Second)
	}

	// Wait for them all to expire.
	for _, timer := range timers {
		<-timer.C
	}

	if int(time.Since(start).Seconds()) != 2 {
		t.Errorf("took ~%v seconds, should be ~2 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestMultipleZeroResets(t *testing.T) {
	start := time.Now()
	var timers []*Timer

	for i := 0; i < 1000; i++ {
		timer := NewTimer(time.Second)
		timers = append(timers, timer)
		timer.Reset(0)
	}

	// Wait for them all to expire.
	for _, timer := range timers {
		<-timer.C
	}

	if int(time.Since(start).Seconds()) != 0 {
		t.Errorf("took ~%v seconds, should be ~0 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestResetChannelClear(t *testing.T) {
	timer := NewTimer(0)
	time.Sleep(time.Second)

	if len(timer.C) != 1 {
		t.Errorf("reset timer: channel should be filled")
	}

	wasActive := timer.Reset(2 * time.Second)
	if wasActive {
		t.Errorf("reset timer: was active is true")
	}

	if len(timer.C) != 0 {
		t.Errorf("reset timer: channel should be empty")
	}

	start := time.Now()
	<-timer.C

	if int(time.Since(start).Seconds()) != 2 {
		t.Errorf("took ~%v seconds, should be ~2 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestResetPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil || r.(string) != "timer: Reset called on uninitialized Timer" {
			t.Errorf("reset timer: invalid reset panic")
		}
	}()

	timer := &Timer{}
	timer.Reset(0)
}

func TestResetBehavior(t *testing.T) {
	start := time.Now()

	// Start a new timer with a timeout of 1 second.
	timer := NewTimer(1 * time.Second)

	// Wait for 2 seconds.
	// Meanwhile the timer fired filled the channel.
	time.Sleep(2 * time.Second)

	// Reset the timer. This should act exactly as creating a new timer.
	timer.Reset(1 * time.Second)

	// However this will fire immediately, because the channel was not drained.
	// See issue: https://github.com/golang/go/issues/11513
	<-timer.C

	if int(time.Since(start).Seconds()) != 3 {
		t.Errorf("took ~%v seconds, should be ~3 seconds\n", int(time.Since(start).Seconds()))
	}
}

func TestMultipleTimersForValidTimeouts(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		dur := time.Duration(i%11) * time.Second
		start := time.Now()
		timer := NewTimer(dur)
		wg.Add(1)
		go func() {
			dur /= time.Second
			<-timer.C
			if int(time.Since(start).Seconds()) != int(dur) {
				t.Errorf("took ~%v seconds, should be ~%v seconds\n", int(time.Since(start).Seconds()), int(dur))
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMultipleTimersConcurrentAddRemove(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 100000; i++ {
		timer := NewTimer(time.Nanosecond)
		wg.Add(1)
		go func() {
			<-timer.C
			wg.Done()
		}()
	}

	wg.Wait()
}
