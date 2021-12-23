package line

import (
	"errors"
)

// Embed runs the whole pipeline as a transformer of a parent pipeline.
func (l *Line) Embed(parentIn <-chan interface{}, parentOut chan<- interface{}, parentErrs chan<- error) {
	l.embedInMut.Lock()

	// if not embedded yet
	if l.embedIn == nil {
		l.embedIn = parentIn
		l.p = func(out chan<- interface{}, errs chan<- error) {
			for msg := range parentIn {
				out <- msg
			}
		}
		l.c = func(in <-chan interface{}, errs chan<- error) {
			for msg := range in {
				parentOut <- msg
			}
		}
		l.SetErrs(parentErrs)
		l.embedInMut.Unlock()
	} else if l.embedIn != parentIn {
		l.embedInMut.Unlock()
		parentErrs <- errors.New("multiple input channels passed to same Embed() causing a race condition! Create a new instance of this pipeline instead of calling this Embed multiple times with different 'in' channels")
		return
	}

	err := l.Run()
	if err != nil {
		parentErrs <- err
	}
}
