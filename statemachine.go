package raft

type StateMachine interface {
	Save() ([]byte, error)
	Recovery([]byte)
}
