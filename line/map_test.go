package line_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Reisender/pipe/line"
)

type foo struct {
	ID   int
	Name string
}

func TestMap(t *testing.T) {
	ctx := context.Background()

	run := func(in <-chan interface{}, fn interface{}) (chan interface{}, chan error) {
		out := make(chan interface{}, 10)
		errs := make(chan error, 10)
		defer close(out)
		defer close(errs)

		line.Map(fn)(ctx, in, out, errs)

		return out, errs
	}

	check := func(fn, want interface{}) {
		in := make(chan interface{}, 1)
		in <- want
		close(in)

		out, errs := run(in, fn)
		got := <-out

		if got != want {
			t.Errorf("want %v got %v", want, got)
		}
		if len(errs) > 0 {
			t.Errorf("got errs %v", <-errs)
		}
	}

	checkForErr := func(fn interface{}) {
		in := make(chan interface{}, 1)
		in <- struct{}{}
		close(in)

		out, errs := run(in, fn)

		if len(out) != 0 {
			t.Errorf("want 0 got %v", len(out))
		}
		if len(errs) != 1 {
			t.Errorf("want 1 got %v", len(errs))
		}
	}

	checkForNone := func(fn interface{}) {
		in := make(chan interface{}, 1)
		in <- struct{}{}
		close(in)

		out, errs := run(in, fn)

		if len(out) != 0 {
			t.Errorf("want 0 got %v", len(out))
		}
		if len(errs) != 0 {
			t.Errorf("want 0 got %v", len(errs))
		}
	}

	t.Run("InlineTfunc", func(t *testing.T) { // make sure it is backwards compatible with the InlineTfunc
		check(func(msg interface{}) (interface{}, error) {
			return msg, nil
		}, "foo")
	})

	t.Run("string arg", func(t *testing.T) {
		check(func(msg string) (string, error) {
			return msg, nil
		}, "foo")
	})

	t.Run("int arg", func(t *testing.T) {
		check(func(msg int) (int, error) {
			return msg, nil
		}, 42)
	})

	t.Run("struct arg", func(t *testing.T) {
		check(func(msg foo) (foo, error) {
			return msg, nil
		}, foo{42, "bar"})
	})

	t.Run("pointer arg", func(t *testing.T) {
		check(func(msg *foo) (*foo, error) {
			return msg, nil
		}, &foo{42, "bar"})
	})

	t.Run("pass err on", func(t *testing.T) {
		checkForErr(func(msg interface{}) (interface{}, error) {
			return nil, errors.New("foo")
		})
	})

	t.Run("filter out nils", func(t *testing.T) {
		checkForNone(func(msg interface{}) (interface{}, error) {
			return nil, nil
		})

		checkForNone(func(msg interface{}) interface{} {
			return nil
		})
	})

	t.Run("with context", func(t *testing.T) {
		check(func(localctx context.Context, msg *foo) (*foo, error) {
			if ctx != localctx {
				t.Fail()
			}
			return msg, nil
		}, &foo{42, "bar"})
	})

	t.Run("type mismatch", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("want panic but didn't happen")
			}
		}()

		// mismatched type between the want message and the msg arg string -> int
		fn := func(msg int) (int, error) {
			return msg, nil
		}
		want := "foo"

		in := make(chan interface{}, 1)
		out := make(chan interface{}, 1)
		errs := make(chan error, 1)
		defer close(in)
		defer close(out)
		defer close(errs)

		in <- want
		line.Map(fn)(ctx, in, out, errs)
	})

	t.Run("wrong shape", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("want panic but didn't happen")
			}
		}()

		line.Map(func(msg, msg2 int) (int, error) {
			return msg, nil
		})
	})

	t.Run("too many args", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("want panic but didn't happen")
			}
		}()

		line.Map(func(lctx context.Context, msg, msg2 int) (int, error) {
			return msg, nil
		})
	})
}
