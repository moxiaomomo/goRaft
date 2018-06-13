package raft

import (
	"sync"

	pb "github.com/moxiaomomo/goRaft/proto"
	"golang.org/x/net/context"
)

// RequestVoteImp a vote-request implementation
type RequestVoteImp struct {
	mutex  sync.RWMutex
	server *server
}

// RequestVoteMe handles the vote request from others
func (e *RequestVoteImp) RequestVoteMe(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	voteGranted := false

	if e.server.IsServerMember(req.GetCandidateName()) {
		lastindex, _ := e.server.log.LastLogInfo()
		if e.server.State() == Candidate && req.Term > e.server.VotedForTerm() && req.LastLogIndex >= lastindex {
			voteGranted = true
			// vote only once for one term
			e.server.SetVotedForTerm(req.Term)
		}
	}

	pb := &pb.VoteResponse{
		Term:        req.Term,
		VoteGranted: voteGranted,
	}
	return pb, nil
}
