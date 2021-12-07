package x

import (
	"github.com/Reisender/pipe/line"
	"github.com/Reisender/pipe/message"
)

// Diff is a producer that compares two pipelines and emits the
// differences as message.Diff{} structs
func Diff(pipe1, pipe2 line.Pipeline, includeMatches bool) line.Pfunc {
	return func(out chan<- interface{}, errs chan<- error) {
		pipe1Out := make(chan interface{})
		pipe2Out := make(chan interface{})

		go func() {
			defer close(pipe1Out)
			pipe1.Add(Tap(pipe1Out)).Run()
		}()

		go func() {
			defer close(pipe2Out)
			pipe2.Add(Tap(pipe2Out)).Run()
		}()

		// process the first pipe and extract the keys
		data := map[string]interface{}{}
		for m := range pipe1Out {
			msg := m.(message.Keyer)
			data[msg.Key()] = m
		}

		// process the second pipe and compare the keys
		// this is where the diffs are produced
		for m := range pipe2Out {
			msg := m.(message.Keyer)

			if first, ok := data[msg.Key()]; ok {

				// diff the content and not just the Key() if we can
				if hasher1, ok1 := first.(message.Hasher); ok1 {
					if hasher2, ok2 := m.(message.Hasher); ok2 {
						if hasher1.Hash() != hasher2.Hash() {
							out <- message.Diff{Left: first, Right: m}
						} else if includeMatches {
							// since the hashes match, we can emit as a match
							out <- message.Diff{Left: first, Right: msg}
						}
					}
				} else if includeMatches {
					// since there is not hash to check, we have a match so emit
					out <- message.Diff{Left: first, Right: msg}
				}

				delete(data, msg.Key()) // remove it since it matches both
			} else {
				// not in first stream so emit a diff
				out <- message.Diff{Right: msg}
			}
		}

		// new emit anything left in the first stream's data
		for _, msg := range data {
			out <- message.Diff{Left: msg}
		}
	}
}
