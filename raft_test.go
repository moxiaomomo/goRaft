package raft

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	confPath = flag.String("confpath", "raft.cfg", "configuration filepath")
)

func Test_raftserver(t *testing.T) {
	flag.Parse()

	raftsvr, err := NewServer("", *confPath)
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
