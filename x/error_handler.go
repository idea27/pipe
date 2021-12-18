package x

import (
	"sync"

	l "github.com/Reisender/pipe/line"
)

type ErrorHandler struct {
	TaskToTry l.InlineTfunc

	// NewErrorHandler  creates a new Tfunc to pair with
	// the new "errIn" channel created for each call to T here.
	// The Tfunc error handler is often and Embed of another
	// pipeline. Since we are creating a new "in" channel, we
	// need a new handler to range over that new channel.
	// Generally, this only happens when this error handler is
	// used inside of a line.Many, where T is called many times
	// and generates a new "in" channel for each call. If the
	// ErrorHandler is an Embed, it needs to be a new instance
	// of the line, otherwise it will cause a race condition
	// and one of the "in" channels might not get ranged over.
	NewErrorHandler func() l.Tfunc
}

// Process a message through the 'try' function.
// If there is an error, then pass it to the error handler. The error
// handler can determine what should happen, e.g. log in a special way
// and pass the error on, or ignore the error
func (eh ErrorHandler) T(in <-chan interface{}, out chan<- interface{}, errs chan<- error) {
	var wg sync.WaitGroup
	var errIn chan interface{}

	// Setup the error handler channel and goroutine if present
	errIn = make(chan interface{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		eh.NewErrorHandler()(errIn, out, errs)
	}()

	// For each message processed by the 'try' function
	for msg := range in {
		outMsg, err := eh.TaskToTry(msg)

		if err == nil { // No error, then pass the message on
			out <- outMsg
		} else {
			// Error, so pass it on to the error handler if it is present
			// TODO: what should we do with err?
			if outMsg == nil {
				errIn <- msg
			} else {
				errIn <- outMsg
			}
		}
	}

	close(errIn)
	wg.Wait()
}
