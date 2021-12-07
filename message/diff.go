package message

// Keyer defines a message that has a Key() func
// that can be used as the "key" for this message.
type Keyer interface {
	Key() string
}

// Hasher defines a message that has a Hash() func.
// The Hash() func is mean to return a hash of the content in the message.
// Then the hash value of the message can easily be compaired
// with the content of other messages.
type Hasher interface {
	Hash() string
}

// Diff is a message that holds two messages that differ
// but that match by Key()
type Diff struct {
	Left  interface{} `json:"mg"`
	Right interface{} `json:"pg"`
}
