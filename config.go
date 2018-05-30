package raft

// Config configuration for a server node
type Config struct {
	LogPrefix   string   `json:"logprefix"`
	CommitIndex uint64   `json:"commitIndex"`
	PeerHosts   []string `json:"peerHosts"`
	Host        string   `json:"host"`
	Client      string   `json:"client"`
	Name        string   `json:"name"`
}
