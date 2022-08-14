package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func NewEvent(name string, op fsnotify.Op) fsnotify.Event {
	return fsnotify.Event{
		Name: name,
		Op:   op,
	}
}

type results struct {
	ran     byte
	blocked byte
}

type AssertRunOnce struct {
	runs   int
	blocks int
	buffer string
}

// region TestRunOnce / TestRunOncePerPeriod
func TestRunOnceSync(t *testing.T) {

	got := results{
		ran:     0,
		blocked: 0,
	}
	want := results{
		ran:     1,
		blocked: 2,
	}

	lock := sync.Mutex{}

	for i := 0; i < 3; i++ {
		err := RunOnce(func() {}, &lock)
		if errors.Is(err, ErrBlocked) {
			got.blocked += 1
		} else {
			got.ran += 1
		}
	}

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestRunOnceConcurrent(t *testing.T) {

	res := make(chan results)
	got := results{
		ran:     0,
		blocked: 0,
	}
	want := results{
		ran:     1,
		blocked: 2,
	}

	lock := sync.Mutex{}

	for i := 0; i < 3; i++ {
		go func(i int, r chan results) {
			err := RunOnce(func() {}, &lock)
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
		}(i, res)
	}

	for i := 0; i < 3; i++ {
		r := <-res
		got.ran += r.ran
		got.blocked += r.blocked
	}

	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestRunOncePerPeriodSync(t *testing.T) {

	var tf = func(s string) string { return fmt.Sprintf("%v", s) }

	// Call OnConfigChange twice, expect only first to succeed
	cases := []struct {
		event fsnotify.Event
		after time.Duration
		ran   bool
	}{
		{
			event: NewEvent("first", fsnotify.Write),
			after: 0,
			ran:   true,
		},
		{
			event: NewEvent("second", fsnotify.Write),
			after: 0,
			ran:   false,
		},
		{
			event: NewEvent("third", fsnotify.Write),
			after: 2 * time.Second,
			ran:   true,
		},
	}

	// Spy on output
	spy := bytes.Buffer{}

	// Got/want structs
	got := AssertRunOnce{0, 0, ""}
	want := AssertRunOnce{0, 0, ""}

	lock := sync.Mutex{}
	timeout := 1 * time.Second

	for _, test := range cases {
		var err error
		if test.ran {
			want.runs += 1
		} else {
			want.blocks += 1
		}
		time.Sleep(test.after)
		err = RunOncePerPeriod(func() {
			spy.WriteString(tf(test.event.Name))
		}, &lock, timeout)
		if errors.Is(err, ErrBlocked) {
			got.blocks += 1
		} else {
			got.runs += 1
		}
	}

	if got.runs != want.runs {
		t.Errorf("number of runs: got %d, wanted %d", got.runs, want.runs)
	}

	if got.blocks != want.blocks {
		t.Errorf("number of blocks: got %d, wanted %d", got.blocks, want.blocks)
	}

	got.buffer = spy.String()
	want.buffer = "firstthird"

	if got.buffer != want.buffer {
		t.Errorf("spyBuffer: got %v, wanted %v", got.buffer, want.buffer)
	}

}

func TestRunOncePerPeriodConcurrent(t *testing.T) {

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

// endregion

// region TestUnmarshal
func TestUnmarshal(t *testing.T) {

	type nested struct {
		String string   `validate:"required"`
		Slice  []string `validate:"required"`
	}

	type cfg struct {
		String string   `validate:"required"`
		Slice  []string `validate:"required"`
		Nested nested   `validate:"required"`
	}

	viper.SetConfigType("yaml")

	cases := []struct {
		gotConfig  cfg
		gotError   error
		source     []byte
		wantConfig cfg
		wantError  error
	}{
		{
			gotConfig: cfg{},
			gotError:  nil,
			source: []byte(`
String: hello
Slice: [hello, world]
Nested:
  String: hello
  Slice: [hello, world]	
`),
			wantConfig: cfg{
				String: "hello",
				Slice:  []string{"hello", "world"},
				Nested: nested{
					String: "hello",
					Slice:  []string{"hello", "world"},
				},
			},
		},
		{
			gotConfig:  cfg{},
			gotError:   nil,
			source:     []byte(""),
			wantConfig: cfg{},
			wantError:  errors.New("validation error"),
		},
		{
			gotConfig: cfg{},
			gotError:  nil,
			source: []byte(`
		String: "hello"
		`),
			wantConfig: cfg{},
			wantError:  errors.New("validation error"),
		},
	}

	viper.SetConfigType("yaml")

	for i, test := range cases {
		viper.ReadConfig(bytes.NewBuffer(test.source))
		cases[i].gotError = Unmarshal(&cases[i].gotConfig)
	}

	for _, test := range cases {
		if !reflect.DeepEqual(test.gotConfig, test.wantConfig) {
			t.Errorf("unmarshalling: got %v, wanted %v", test.gotConfig, test.wantConfig)
		}
		if test.gotError != nil && test.wantError == nil {
			t.Errorf("errors: got %v, wanted %v", test.gotError, test.wantError)
		}
	}

}

// endregion

func TestDetermineChanges(t *testing.T) {

}
