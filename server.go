package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	pb "github.com/moxiaomomo/goRaft/proto"
	"github.com/moxiaomomo/goRaft/util"
	"github.com/moxiaomomo/goRaft/util/logger"
	"google.golang.org/grpc"
)

// HandleFuncType external handle function type
type HandleFuncType func(w http.ResponseWriter, r *http.Request)

type server struct {
	mutex sync.RWMutex

	path  string
	state string

	currentLeaderName   string
	currentLeaderHost   string
	currentLeaderExHost string
	currentTerm         uint64

	log      *Log
	conf     *Config
	confPath string

	lastHeartbeatTime int64
	lastSnapshotTime  int64
	heartbeatInterval uint64
	votedForTerm      uint64 // vote one peer as a leader in curterm

	stopped    chan bool
	ch         chan interface{}
	handlefunc map[string]HandleFuncType

	peers    map[string]*Peer
	syncpeer map[string]int
}

// Server current server operations
type Server interface {
	Start() error
	State() string
	IsRunning() bool
	CurLeaderExHost() string

	AddPeer(name string, host string) error
	RemovePeer(name string, host string) error

	RegisterCommand(cmd Command)
	RegisterHandler(urlpath string, fc HandleFuncType)
	OnAppendEntry(cmd Command, cmds []byte)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// NewServer creates a new server instance
func NewServer(workdir string, confPath string) (Server, error) {
	_ = os.Mkdir(workdir, 0700)

	if confPath == "" || !strings.Contains(confPath, "/") {
		srvaddr := os.Getenv(EnvClusterNodeSrvAddr)
		if len(srvaddr) <= 0 {
			return nil, errors.New("Invalid configuration")
		}
		confPath = fmt.Sprintf("/opt/raft/raft_%s.cfg", srvaddr)
	}
	confdir := confPath[0:strings.LastIndex(confPath, "/")]
	if exist, _ := pathExists(confdir); !exist {
		os.MkdirAll(confdir, 0700)
	}
	if exist, _ := pathExists(confPath); !exist {
		defaultCfg := `{"logprefix":"raft-log-","commitIndex":0,
			"peerHosts":["0.0.0.0:3000","0.0.0.0:3001","0.0.0.0:3002"],
			"host":"0.0.0.0:3000","client":"0.0.0.0:4000","name":"server0"}`
		ioutil.WriteFile(confPath, []byte(defaultCfg), 0644)
	}

	s := &server{
		path:              workdir,
		confPath:          confPath,
		state:             Stopped,
		log:               newLog(),
		heartbeatInterval: 300, // 300ms
		lastSnapshotTime:  util.GetTimestampInMilli(),
		syncpeer:          make(map[string]int),
		handlefunc:        make(map[string]HandleFuncType),
	}
	return s, nil
}

// SetTerm set current term
func (s *server) SetTerm(term uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentTerm = term
}

// SyncPeerStatusOrReset update or reset peers' response status
func (s *server) SyncPeerStatusOrReset() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sucCnt, failCnt := 0, 0
	for _, v := range s.syncpeer {
		if v == 1 {
			sucCnt++
		} else if v == 0 {
			failCnt++
		}
	}

	qsize := s.quorumSize()
	if sucCnt >= qsize {
		s.resetSyncPeer()
		return 1
	} else if failCnt >= qsize {
		s.resetSyncPeer()
		return 0
	}
	return -1
}

func (s *server) InitSyncPeer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.resetSyncPeer()
}

func (s *server) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := RunningStates[s.state]
	return ok
}

func (s *server) State() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.state
}

func (s *server) CurLeaderExHost() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.currentLeaderExHost
}

func (s *server) VotedForTerm() uint64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.votedForTerm
}

func (s *server) SetVotedForTerm(term uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.votedForTerm = term
}

func (s *server) Peers() map[string]*Peer {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.peers
}

func (s *server) IncrTermForvote() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentTerm++
}

func (s *server) SetState(state string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state = state
}

func (s *server) VoteForSelf() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.syncpeer[s.conf.Host] = 1
	s.peers[s.conf.Host].SetVoteRequestState(VoteGranted)
}

func (s *server) IsServerMember(host string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.peers[host]; ok {
		return true
	}
	return false
}

func (s *server) IsHeartbeatTimeout() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return util.GetTimestampInMilli()-s.lastHeartbeatTime > int64(s.heartbeatInterval*2)
}

func (s *server) RegisterCommand(cmd Command) {
	RegisterCommand(cmd)
}

// Init steps:
// check if running or initiated before
// load configuration file
// load raft log
// recover server persistent status
// set state = Initiated
func (s *server) Init() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%s", s.State())
	}

	if s.State() == Initiated {
		s.SetState(Initiated)
		return nil
	}

	err := os.MkdirAll(path.Join(s.path, "snapshot"), 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft initiation error: %s", err)
	}

	err = s.loadConf()
	if err != nil {
		return fmt.Errorf("raft load config error: %s", err)
	}
	logger.LogInfof("config: %+v\n", s.conf)

	logpath := path.Join(s.path, "internlog")
	err = os.MkdirAll(logpath, 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}
	logger.LogInfof("%+v\n", s.conf)
	if err = s.log.LogInit(fmt.Sprintf("%s/%s%s", logpath, s.conf.LogPrefix, s.conf.Name)); err != nil {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}

	err = s.LoadState()
	if err != nil {
		return fmt.Errorf("raft load srvstate error: %s", err)
	}

	// register raft commands
	s.RegisterCommand(&DefaultJoinCommand{})
	s.RegisterCommand(&DefaultLeaveCommand{})
	s.RegisterCommand(&NOPCommand{})

	s.SetState(Initiated)
	return nil
}

// start steps:
// comlete initiation
// set state = Follower
// new goroutine for tcp listening
// enter loop with a propriate state
func (s *server) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%s", s.State())
	}

	if err := s.Init(); err != nil {
		return err
	}

	s.stopped = make(chan bool)

	s.SetState(Follower)

	go s.StartInternServe()
	go s.StartExternServe()
	go s.TickTask()

	s.loop()
	return nil
}

func (s *server) StartInternServe() {
	server := grpc.NewServer()
	pb.RegisterRequestVoteServer(server, &RequestVoteImp{server: s})
	pb.RegisterAppendEntriesServer(server, &AppendEntriesImp{server: s})

	logger.LogInfof("listen internal rpc address: %s\n", s.conf.Host)
	address, err := net.Listen("tcp", s.conf.Host)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(address); err != nil {
		panic(err)
	}
}

func (s *server) TickTask() {
	t := time.NewTimer(300 * time.Millisecond)
	for {
		select {
		case <-t.C:
			s.FlushState()
			nowt := util.GetTimestampInMilli()
			// snapshotting every 5 second
			if nowt-s.lastSnapshotTime > 5000 {
				s.onSnapShotting()
				s.lastSnapshotTime = nowt
			}
			t.Reset(300 * time.Millisecond)
		}
	}
}

func (s *server) OnAppendEntry(cmd Command, cmds []byte) {
	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()

	entry := &pb.LogEntry{
		Index:       lindex + 1,
		Term:        s.currentTerm,
		Commandname: cmd.CommandName(),
		Command:     cmds,
	}
	s.log.AppendEntry(&LogEntry{Entry: entry})

	if s.State() == Leader {
		for idx := range s.peers {
			if s.conf.Host == s.peers[idx].Host {
				continue
			}
			go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{entry}, findex, lindex, lterm)
		}
	}
}

func (s *server) AddPeer(name string, host string) error {
	s.mutex.Lock()

	if s.peers[host] != nil {
		s.mutex.Unlock()
		return nil
	}

	if s.conf.Name != name {
		ti := time.Duration(s.heartbeatInterval) * time.Millisecond
		peer := NewPeer(s, name, host, ti)
		s.peers[host] = peer
	}

	// to flush configuration
	logger.LogInfo("To rewrite configuration to persistent storage.")
	_ = s.writeConf()

	s.mutex.Unlock()

	if s.State() == Leader {
		lindex, _ := s.log.LastLogInfo()
		cmdinfo := &DefaultJoinCommand{
			Name: name,
			Host: host,
		}
		cmdjson, _ := json.Marshal(cmdinfo)
		entry := &pb.LogEntry{
			Index:       lindex + 1,
			Term:        s.currentTerm,
			Commandname: cmdinfo.CommandName(),
			Command:     []byte(cmdjson),
		}
		s.onMemberChanged(entry)
	}
	return nil
}

func (s *server) RemovePeer(name string, host string) error {
	s.mutex.Lock()
	if s.peers[host] == nil || s.conf.Host == host {
		s.mutex.Unlock()
		return nil
	}

	delete(s.peers, host)

	// to flush configuration
	logger.LogInfo("To rewrite configuration to persistent storage.")
	_ = s.writeConf()
	s.mutex.Unlock()

	if s.State() == Leader {
		lindex, _ := s.log.LastLogInfo()
		cmdinfo := &DefaultLeaveCommand{
			Name: name,
			Host: host,
		}
		cmdjson, _ := json.Marshal(cmdinfo)
		entry := &pb.LogEntry{
			Index:       lindex + 1,
			Term:        s.currentTerm,
			Commandname: cmdinfo.CommandName(),
			Command:     []byte(cmdjson),
		}

		s.onMemberChanged(entry)
	}
	return nil
}

func (s *server) loop() {
	for s.State() != Stopped {
		logger.LogInfof("current state:%s, term:%d\n", s.State(), s.currentTerm)
		switch s.State() {
		case Follower:
			s.followerLoop()
		case Candidate:
			s.candidateLoop()
		case Leader:
			s.leaderLoop()
			//		case Snapshotting:
			//			s.snapshotLoop()
		case Stopped:
			// TODO: do something before server stop
			break
		}
	}
}

func (s *server) candidateLoop() {
	t := time.NewTimer(time.Duration(150+rand.Intn(150)) * time.Millisecond)
	for s.State() == Candidate {
		select {
		case c := <-s.ch:
			switch d := c.(type) {
			case *RequestVoteRespChan:
				if d.Failed == false {
					if d.Resp.VoteGranted {
						s.syncpeer[d.PeerHost] = 1
						s.peers[d.PeerHost].SetVoteRequestState(VoteGranted)
					} else {
						s.syncpeer[d.PeerHost] = 0
						s.peers[d.PeerHost].SetVoteRequestState(VoteRejected)
					}
				}
				respStatus := s.SyncPeerStatusOrReset()
				if respStatus == 1 {
					s.SetState(Leader)
					t.Stop()
					return
				}
			}
		case <-t.C:
			if s.State() != Candidate {
				return
			}
			s.IncrTermForvote()
			s.VoteForSelf()
			lindex, lterm := s.log.LastLogInfo()
			for idx := range s.peers {
				if s.conf.Host == s.peers[idx].Host {
					continue
				}
				go s.peers[idx].RequestVoteMe(lindex, lterm)
			}
			t.Reset(time.Duration(150+rand.Intn(150)) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) followerLoop() {
	t := time.NewTimer(time.Duration(s.heartbeatInterval) * time.Millisecond)
	for s.State() == Follower {
		select {
		case <-t.C:
			if s.State() != Follower {
				return
			}
			if s.IsHeartbeatTimeout() {
				s.SetState(Candidate)
			}
			t.Reset(time.Duration(s.heartbeatInterval) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) leaderLoop() {
	// to request append entry as a new leader is elected
	s.syncpeer[s.conf.Host] = 1
	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()
	entry := &pb.LogEntry{
		Index:       lindex + 1,
		Term:        s.currentTerm,
		Commandname: (NOPCommand{}).CommandName(),
		Command:     []byte(""),
	}
	s.log.AppendEntry(&LogEntry{Entry: entry})

	for idx := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			continue
		}
		go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{entry}, findex, lindex, lterm)
	}

	// send heartbeat as leader state
	s.lastHeartbeatTime = util.GetTimestampInMilli()
	t := time.NewTimer(time.Duration(s.heartbeatInterval) * time.Millisecond)
	for s.State() == Leader {
		select {
		case c := <-s.ch:
			switch d := c.(type) {
			case *AppendLogRespChan:
				if d.Failed == false {
					if d.Resp != nil && d.Resp.Success && d.Resp.Term == s.currentTerm {
						s.syncpeer[d.PeerHost] = 1
					} else {
						s.syncpeer[d.PeerHost] = 0
					}
					respStatus := s.SyncPeerStatusOrReset()
					if respStatus != -1 {
						s.lastHeartbeatTime = util.GetTimestampInMilli()
					}
					if respStatus == 1 {
						lcmiIndex, _ := s.log.LastCommitInfo()
						lindex, _ := s.log.LastLogInfo()
						if lindex > lcmiIndex {
							for _, entry := range s.log.entries {
								if entry.Entry.GetIndex() <= lcmiIndex {
									continue
								}
								cmd, _ := NewCommand(entry.Entry.Commandname, entry.Entry.Command)
								if cmdcopy, ok := cmd.(CommandApply); ok {
									cmdcopy.Apply(s)
								}
							}
							s.log.UpdateCommitIndex(lindex)
						}
					}
				}
				if s.IsHeartbeatTimeout() {
					s.SetState(Candidate)
					t.Stop()
					return
				}
			}
		case <-t.C:
			findex := s.log.FirstLogIndex()
			lindex, lterm := s.log.LastLogInfo()
			s.syncpeer[s.conf.Host] = 1
			for idx := range s.peers {
				if s.conf.Host == s.peers[idx].Host {
					continue
				}
				go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{}, findex, lindex, lterm)
			}
			t.Reset(time.Duration(s.heartbeatInterval) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) onMemberChanged(entry *pb.LogEntry) {
	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()
	for idx := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			s.log.AppendEntry(&LogEntry{Entry: entry})
			continue
		}
		go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{entry}, findex, lindex, lterm)
	}
}

func (s *server) onSnapShotting() {
	if s.State() != Leader {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	cmiIndex, _ := s.log.LastCommitInfo()
	backindex := len(s.log.entries) - 1
	for i := backindex; i >= 0; i-- {
		if s.log.entries[i].Entry.GetIndex() < cmiIndex {
			backindex = i
			break
		}
	}
	if backindex < 0 {
		return
	}
	s.log.entries = s.log.entries[backindex:len(s.log.entries)]
	s.log.RefreshLog()

	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()

	pbentries := [](*pb.LogEntry){}
	for _, entry := range s.log.entries {
		pbentries = append(pbentries, entry.Entry)
	}
	for idx := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			continue
		}
		go s.peers[idx].RequestAppendEntries(pbentries, findex, lindex, lterm)
	}
}

func (s *server) resetSyncPeer() {
	for k := range s.syncpeer {
		s.syncpeer[k] = -1
	}
}

func (s *server) quorumSize() int {
	return len(s.peers)/2 + 1
}

func (s *server) loadConf() error {
	cfg, err := ioutil.ReadFile(s.confPath)
	if err != nil {
		return err
	}

	conf := &Config{}
	if err = json.Unmarshal(cfg, conf); err != nil {
		return err
	}
	s.conf = conf

	// read env-variable
	if nodes := os.Getenv(EnvClusterNodeHosts); len(nodes) > 0 {
		s.conf.PeerHosts = strings.Split(nodes, ",")
	}
	if srvaddr := os.Getenv(EnvClusterNodeSrvAddr); len(srvaddr) > 0 {
		s.conf.Host = srvaddr
	}
	if cliaddr := os.Getenv(EnvClusterNodeCliAddr); len(cliaddr) > 0 {
		s.conf.Client = cliaddr
	}
	if srvname := os.Getenv(EnvClusterNodeName); len(srvname) > 0 {
		s.conf.Name = srvname
	}

	s.peers = make(map[string]*Peer)
	hostInPeers := false
	for _, c := range s.conf.PeerHosts {
		if c == s.conf.Host {
			hostInPeers = true
		}
		s.peers[c] = &Peer{
			Name:   c,
			Host:   c,
			server: s,
		}
	}
	if !hostInPeers {
		return errors.New("Current host not in the cluster nodes list")
	}

	s.ch = make(chan interface{}, len(s.peers)*2)
	return nil
}

func (s *server) writeConf() error {
	s.conf.PeerHosts = []string{}
	for _, p := range s.peers {
		s.conf.PeerHosts = append(s.conf.PeerHosts, p.Host)
	}

	f, err := os.OpenFile(s.confPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := json.Marshal(s.conf)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(d))
	return err
}
