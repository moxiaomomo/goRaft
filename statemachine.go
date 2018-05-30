package raft

// StateMachine statemachine interface
type StateMachine interface {
	Save() ([]byte, error)
	Recovery([]byte)
}
