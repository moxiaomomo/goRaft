package raft

const (
	// Initiated initiated status
	Initiated = "initiated"
	// Snapshotting snapshotting status
	Snapshotting = "snapshotting"
	// Stopped server is stopped
	Stopped = "stopped"

	// Leader node is a leader
	Leader = "leader"
	// Follower node is a follower
	Follower = "follower"
	// Candidate node is a candidate
	Candidate = "candidate"

	// NotYetVote node has not yet vote for a term
	NotYetVote = 0
	// VoteRejected node has rejected vote for a term
	VoteRejected = 1
	// VoteGranted node has voted for a term
	VoteGranted = 2

	// EnvClusterNodeHosts environment variable
	EnvClusterNodeHosts = "MEMBER_NODES"
	// EnvClusterNodeSrvAddr environment variable
	EnvClusterNodeSrvAddr = "MEMBER_NODE_SRV"
	// EnvClusterNodeCliAddr environment variable
	EnvClusterNodeCliAddr = "MEMBER_NODE_CLI"
	// EnvClusterNodeName environment variable
	EnvClusterNodeName = "MEMBER_NODE_Name"
)
