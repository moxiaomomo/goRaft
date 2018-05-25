package raft

const (
	Initiated    = "initiated"
	Snapshotting = "snapshotting"
	Stopped      = "stopped"

	Leader    = "leader"
	Follower  = "follower"
	Candidate = "candidate"

	NotYetVote   = 0
	VoteRejected = 1
	VoteGranted  = 2
)
