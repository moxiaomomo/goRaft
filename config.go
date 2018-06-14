package raft

// Config configuration for a server node
type Config struct {
	LogPrefix string `json:"logprefix"`
	//	CommitIndex     uint64            `json:"commitIndex"`
	Peers           map[string]string `json:"peers"`
	Host            string            `json:"host"`
	Client          string            `json:"client"`
	Name            string            `json:"name"`
	BootstrapExpect int               `json:"bExpect"` // expect number of nodes for leader election
	JoinTarget      string            `json:"join"`
}
