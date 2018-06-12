package raft

// RunningStates if server is in these states, the server is considered as running
var RunningStates = map[string]bool{
	Leader:       true,
	Follower:     true,
	Candidate:    true,
	Snapshotting: true,
}

// DefaultConfig default configuration
var DefaultConfig = `{"logprefix":"raft-log-","commitIndex":0,
	"peerHosts":["127.0.0.1:3000","127.0.0.1:3001","127.0.0.1:3002"],
	"host":"127.0.0.1:3000","client":"127.0.0.1:4000","name":"server0"}`

// DefaultConfigContain default configuration in container
var DefaultConfigContain = `{"logprefix":"raft-log-","commitIndex":0,
	"peerHosts":["172.19.0.2:3000","172.19.0.3:3000","172.19.0.4:3000"],
	"host":"172.19.0.2:3000","client":"172.19.0.2:4000","name":"server0"}`
