package raft

type Context interface {
	Server() Server
	CurrentTerm() uint64
	CurrentIndex() uint64
	//	CommitIndex() uint64
}
