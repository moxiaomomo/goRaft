package raft

import (
	"flag"
	"os"
	"strconv"
)

var (
	svrname = flag.String("name", "server0", "the server name")
	svrhost = flag.String("host", "0.0.0.0:3000", "the server internel address-binded")
	client  = flag.String("client", "0.0.0.0:4000", "the server externel address-binded")
	join    = flag.String("join", "", "the target address to join")
	bexpect = flag.Int("bexpect", 1, "the node number epxtected to start leader election")
)

type Params struct {
	SvrName         string
	SvrHost         string
	Client          string
	JoinTarget      string
	BootstrapExpect int
}

func ParseEnvParams() Params {
	flag.Parse()

	p := Params{}

	// get env-varaibles
	if srvaddr := os.Getenv(EnvClusterNodeSrvAddr); len(srvaddr) > 0 {
		p.SvrHost = srvaddr
	}
	if cliaddr := os.Getenv(EnvClusterNodeCliAddr); len(cliaddr) > 0 {
		p.Client = cliaddr
	}
	if srvname := os.Getenv(EnvClusterNodeName); len(srvname) > 0 {
		p.SvrName = srvname
	}
	if jointarget := os.Getenv(EnvJoinTarget); len(jointarget) > 0 {
		p.JoinTarget = jointarget
	}
	if bexpect := os.Getenv(EnvBootstrapExpect); len(bexpect) > 0 {
		p.BootstrapExpect, _ = strconv.Atoi(bexpect)
	}

	// get command variables
	if len(*svrname) > 0 {
		p.SvrName = *svrname
	}
	if len(*svrhost) > 0 {
		p.SvrHost = *svrhost
	}
	if len(*client) > 0 {
		p.Client = *client
	}
	if len(*join) > 0 {
		p.JoinTarget = *join
	}
	if *bexpect > 0 {
		p.BootstrapExpect = *bexpect
	}
	return p
}
