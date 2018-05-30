package raft

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	exitTime = flag.Int("exittime", 0, "the time (second) for running server")
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

	if *exitTime > 0 {
		t := time.NewTimer(time.Duration(*exitTime) * time.Second)
		go func() {
			select {
			case <-t.C:
				os.Exit(0)
			}
		}()
	}

	err = raftsvr.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
