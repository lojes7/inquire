package model

const (
	DEFAULT uint = iota
	RECALLED
	SYSTEM
)

type Message struct {
	MyModel
}

type Conversation struct {
	MyModel
}

type MessageUser struct {
	MyModel
}
