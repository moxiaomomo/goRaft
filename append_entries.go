package raft

import (
	//	"encoding/json"
	//	"fmt"
	pb "github.com/moxiaomomo/goRaft/proto"
	"github.com/moxiaomomo/goRaft/util"
	"golang.org/x/net/context"
	"sync"
)

type AppendEntriesImp struct {
	server *server
	mutex  sync.Mutex
}

// handle appendentries request
func (e *AppendEntriesImp) AppendEntries(ctx context.Context, req *pb.AppendEntriesReuqest) (*pb.AppendEntriesResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	reqentries := req.GetEntries()
	isFullLog := false
	if len(reqentries) > 0 {
		isFullLog = req.GetFirstLogIndex() == reqentries[0].GetIndex()
	}
	lindex, _ := e.server.log.LastLogInfo()

	pb := &pb.AppendEntriesResponse{
		Success: false,
		Index:   lindex,
		Term:    req.GetTerm(),
	}

	// peer host should be in the configuration
	if e.server.IsServerMember(req.LeaderHost) {
		// update current server state
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeaderName = req.GetLeaderName()
		e.server.currentLeaderHost = req.GetLeaderHost()
		e.server.currentLeaderExHost = req.GetLeaderExHost()
		e.server.lastHeartbeatTime = util.GetTimestampInMilli()
		e.server.heartbeatInterval = req.GetHeartbeatInterval()

		// 1.if isfulllog, just overwrite log file
		if isFullLog {
			e.server.log.entries = []*LogEntry{}
			for _, entry := range reqentries {
				e.server.log.entries = append(e.server.log.entries, &LogEntry{Entry: entry})
			}
			e.server.log.RefreshLog()
			pb.Success = true
			// 2.if req-entries's startindex is later than current server's lastlogindex,
			//   or the current server's log is empty,
			//   to request re-send fulllog
		} else if req.GetFirstLogIndex() > lindex || e.server.log.IsEmpty() {
			pb.Success = false
			pb.Index = 0
			// 3.if req-entries's prelogindex is later than current server's lastlogindex,
			//   to request re-send logs from the position of lindex
		} else if req.GetPreLogIndex() > lindex {
			pb.Success = false
			// 4. if req-entries's prelogindex is earlier than current server's lastlogindex,
			//    truncate reqentries and append into local logfile
		} else if req.GetPreLogIndex() < lindex {
			backindex := len(e.server.log.entries) - 1
			for i := backindex; i >= 0; i-- {
				if e.server.log.entries[i].Entry.GetIndex() == req.GetPreLogIndex() {
					backindex = i
					break
				}
			}
			// 5.if backindex<0, something imnormal...just to request re-send fulllog
			if backindex < 0 {
				pb.Success = false
				pb.Index = 0
			} else {
				e.server.log.entries = e.server.log.entries[0 : backindex+1]
				e.server.log.RefreshLog()
				for _, entry := range reqentries {
					e.server.log.AppendEntry(&LogEntry{Entry: entry})
				}
				pb.Success = true
			}
			// 6.if req-entries's prelogindex is equal to current server's lastlogindex,
			//   just to append reqentries into logfile
		} else if req.GetPreLogIndex() == lindex {
			for _, entry := range reqentries {
				e.server.log.AppendEntry(&LogEntry{Entry: entry})
			}
			pb.Success = true
		}
	}

	// if appendentries succeeded, apply commands and update commited index
	if pb.Success {
		cmiindex, _ := e.server.log.LastCommitInfo()
		// apply the command
		if req.GetCommitIndex() > cmiindex {
			for _, entry := range e.server.log.entries {
				if entry.Entry.GetIndex() <= cmiindex || entry.Entry.GetIndex() > req.GetCommitIndex() {
					continue
				}
				cmd, _ := NewCommand(entry.Entry.Commandname, entry.Entry.Command)
				if cmdcopy, ok := cmd.(CommandApply); ok {
					cmdcopy.Apply(e.server)
				}
			}
		}

		// update commited index
		lindex := e.server.log.LastLogIndex()
		if lindex > req.GetCommitIndex() {
			e.server.log.UpdateCommitIndex(req.GetCommitIndex())
		} else {
			e.server.log.UpdateCommitIndex(lindex)
		}
		pb.Index = lindex
	}

	return pb, nil
}
