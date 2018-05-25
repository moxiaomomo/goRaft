package raft

var RunningStates = map[string]bool{
	Leader:       true,
	Follower:     true,
	Candidate:    true,
	Snapshotting: true,
}
