package pigeon

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func NewEvent(name string, op fsnotify.Op) fsnotify.Event {
	return fsnotify.Event{
		Name: name,
		Op:   op,
	}
}

func TestRunOncePerPeriod(t *testing.T) {

	type results struct {
		ran     byte
		blocked byte
	}

	res := make(chan results)
	got := []results{{}, {}, {}}
	want := []results{{1, 2}, {0, 3}, {1, 2}}

	lock := sync.Mutex{}

	// 	 j
	// i 0 1 2
	// 0 r b b
	// 1 b b b
	// 2 r b b
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		for j := 0; j < 3; j++ {
			go func(r chan results) {
				err := RunOncePerPeriod(func() {}, &lock, 150*time.Millisecond)
				if errors.Is(err, ErrBlocked) {
					res <- results{
						ran:     0,
						blocked: 1,
					}
				} else {
					res <- results{
						ran:     1,
						blocked: 0,
					}
				}
			}(res)
		}
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			r := <-res
			got[i].ran += r.ran
			got[i].blocked += r.blocked
		}
	}

	for i := 0; i < 3; i++ {
		if got[i] != want[i] {
			t.Errorf("batch %d: got %v want %v", i, got[i], want[i])
			t.Log(got[i], want[i])
		}
	}
}
