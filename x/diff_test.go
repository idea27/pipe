package x_test

import (
	"fmt"

	l "github.com/Reisender/pipe/line"
	"github.com/Reisender/pipe/x"
)

type Data struct {
	ID      string
	Content string
}

type Keyed Data

func (d Keyed) Key() string {
	return d.ID
}

func ExampleDiff() {
	// first have two pipes you want to diff

	// pipe1 produces messages 0,1,2,3
	pipe1 := l.New().SetP(func(out chan<- interface{}, errs chan<- error) {
		for i := 0; i < 4; i++ {
			out <- Keyed(Data{
				ID:      fmt.Sprintf("%d", i),
				Content: fmt.Sprintf("foo%d", i),
			})
		}
	})

	// pipe2 produces messages 1,2,3,4
	pipe2 := l.New().SetP(func(out chan<- interface{}, errs chan<- error) {
		// don't have 0 and add 4
		for i := 1; i < 5; i++ {
			out <- Keyed(Data{
				ID:      fmt.Sprintf("%d", i),
				Content: fmt.Sprintf("foo%d", i),
			})
		}
	})

	l.New().SetP(x.Diff(
		pipe1,
		pipe2,
		false, // don't send matches downstream
	)).Add(
		l.Stdout,
	).Run()
	// Output:
	// {<nil> {4 foo4}}
	// {{0 foo0} <nil>}
}

type KeyedContent Data

func (kc KeyedContent) Key() string {
	return kc.ID
}

func (kc KeyedContent) Hash() string {
	return kc.Content
}

func ExampleDiff_hash() {
	// first have two pipes you want to diff

	// pipe1 produces messages 0,1,2,3
	pipe1 := l.New().SetP(func(out chan<- interface{}, errs chan<- error) {
		for i := 0; i < 4; i++ {
			out <- KeyedContent(Data{
				ID:      fmt.Sprintf("%d", i),
				Content: fmt.Sprintf("foo%d", i),
			})
		}
	})

	// pipe2 produces messages 1,2,3,4
	pipe2 := l.New().SetP(func(out chan<- interface{}, errs chan<- error) {
		// don't have 0 and add 4
		for i := 1; i < 5; i++ {
			out <- KeyedContent(Data{
				ID:      fmt.Sprintf("%d", i),
				Content: fmt.Sprintf("bar%d", i),
			})
		}
	})

	l.New().SetP(x.Diff(
		pipe1,
		pipe2,
		true, // send matches downstream (shouldn't be any in this case
	)).Add(
		l.Stdout,
	).Run()
	// Output:
	// {{1 foo1} {1 bar1}}
	// {{2 foo2} {2 bar2}}
	// {{3 foo3} {3 bar3}}
	// {<nil> {4 bar4}}
	// {{0 foo0} <nil>}
}
