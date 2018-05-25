package raft

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	workDir  = flag.String("workdir", "/opt/raft/", "raft server workdir")
	confPath = flag.String("confpath", "raft.cfg", "configuration filepath")
)

func Test_raftserver(t *testing.T) {
	flag.Parse()

	raftsvr, err := NewServer(*workDir, *confPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = raftsvr.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
