package main

import (
	"flag"
	"fmt"
	"gomh/registry/raft"
	"os"
)

var (
	confPath = flag.String("confpath", "raft.cfg", "configuration filepath")
)

func main() {
	flag.Parse()

	raftsvr, err := raft.NewServer("", *confPath)
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
