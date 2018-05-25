package raft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type ServerState struct {
	CommitIndex uint64 `json:"commitIndex"`
	Term        uint64 `json:"term"`
	VoteFor     string `json:"voteFor"`
}

// save data into file
func (s *server) FlushState() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state := &ServerState{
		CommitIndex: s.log.CommitIndex(),
		Term:        s.currentTerm,
	}
	d, err := json.Marshal(state)
	if err != nil {
		return err
	}

	logpath := path.Join(s.path, "internlog")
	fname := fmt.Sprintf("%s/state-%s", logpath, s.conf.Name)
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Truncate(0)

	w := bufio.NewWriter(file)
	w.Write([]byte(string(d) + "\n"))
	w.Flush()

	return nil
}

// load data from file
func (s *server) LoadState() error {
	logpath := path.Join(s.path, "internlog")
	fname := fmt.Sprintf("%s/state-%s", logpath, s.conf.Name)

	_, err := os.Stat(fname)
	if err != nil && os.IsNotExist(err) {
		return nil
	}

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	//s.srvstate = ServerState{}
	srvstate := &ServerState{}
	if err = json.Unmarshal(b, srvstate); err != nil {
		return err
	}
	fmt.Printf("state loaded: %+v\n", srvstate)
	s.log.UpdateCommitIndex(srvstate.CommitIndex)
	s.SetTerm(srvstate.Term)
	return nil
}
