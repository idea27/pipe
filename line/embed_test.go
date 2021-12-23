package line_test

import (
	"fmt"
	"sync"
	"testing"

	l "github.com/Reisender/pipe/line"
)

func ExamplePipeline_embed() {
	// define a sub-pipeline to be embedded
	// this one just prints messages out to stdout
	subPipeline := l.New().Add(l.Stdout)

	// setup and run the main pipeline
	l.New().
		SetP(func(out chan<- interface{}, errs chan<- error) {
			out <- "foo from sub"
		}).
		Add(
			subPipeline.Embed, // embed it just like any other Tfunc
		).Run()
	// Output: foo from sub
}

func TestPipeline_Embed(t *testing.T) {
	in := make(chan interface{}, 2)
	out := make(chan interface{})
	errs := make(chan error)
	errCnt := 0
	var err error
	msgCnt := 0
	var msg string

	var wg sync.WaitGroup

	// start with two messages loaded in the buffer
	in <- "err"
	in <- "foo"
	close(in)

	// start the errs range
	wg.Add(1)
	go func() {
		defer wg.Done()
		for e := range errs {
			errCnt++
			err = e
		}
	}()

	// start the out range
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(errs)
		for m := range out {
			msgCnt++
			msg = m.(string)
		}
	}()

	l.New().
		Add(
			l.Inline(func(m interface{}) (interface{}, error) {
				if m.(string) == "err" {
					return nil, fmt.Errorf("foo error")
				} else {
					return m, nil
				}
			}),
		).
		Embed(in, out, errs)
	close(out)
	wg.Wait()

	if msgCnt != 1 {
		t.Errorf("message count: want 1 got %d", msgCnt)
	}

	if msg != "foo" {
		t.Errorf("message error: want foo got %s", msg)
	}

	if errCnt != 1 {
		t.Errorf("error count: want 1 got %d", errCnt)
	}

	if err.Error() != "foo error" {
		t.Errorf("error: want 'foo error' got '%s'", err.Error())
	}
}

func TestPipeline_EmbedMultiuseRace(t *testing.T) {
	in1 := make(chan interface{}, 2)
	in2 := make(chan interface{}, 2)
	out := make(chan interface{})
	errs := make(chan error)
	var wg sync.WaitGroup

	// collection counts to check later
	msgCnt := 0
	errCnt := 0

	in1 <- "foo"
	in1 <- "bar"
	close(in1)
	in2 <- "baz"
	in2 <- "quux"
	close(in2)

	// start the errs range
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range errs {
			errCnt++
		}
	}()

	// start the out range
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(errs)
		for range out {
			msgCnt++
		}
	}()

	// the Embed should error if it is used multiple times with different in channels
	pipe := l.New()

	pipe.Embed(in1, out, errs)

	// these two calls should produce errors
	pipe.Embed(in2, out, errs)
	pipe.Embed(in2, out, errs)

	close(out)
	wg.Wait()

	if errCnt != 2 {
		t.Errorf("Embed race condition err count: want 2 got %d", errCnt)
	}

	if msgCnt != 2 {
		t.Errorf("Embed race condition msg count: want 2 got %d", msgCnt)
	}
}
