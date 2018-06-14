package raft

import (
	"time"

	pb "github.com/moxiaomomo/goRaft/proto"
	"github.com/moxiaomomo/goRaft/util/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// PreJoinRequest PreJoinRequest
func (s *server) PreJoinRequest() {
	conn, err := grpc.Dial(s.conf.JoinTarget, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("dail rpc failed, err: %s\n", err)
		if conn != nil {
			conn.Close()
		}
		return
	}
	defer conn.Close()

	client := pb.NewDiscoveryAsBootClient(conn)
	req := &pb.PreJoinRequest{
		Name:   s.conf.Name,
		Host:   s.conf.Host,
		Client: s.conf.Client,
	}
	res, err := client.PreJoin(context.Background(), req)
	if err != nil {
		logger.Errorf("call rpc failed, err: %s\n", err)
		return
	}
	logger.Infof("prejoin result:%+v\n", res)

	if res.Result != -1 && res.Jointarget == s.conf.JoinTarget {
		for name, host := range res.Curnodes {
			if _, ok := s.peers[name]; !ok {
				ti := time.Duration(s.heartbeatInterval) * time.Millisecond
				s.peers[name] = NewPeer(s, name, host, ti)
			}
		}
		s.conf.Peers = res.Curnodes
		s.conf.BootstrapExpect = int(res.Bootexpect)
	} else {
		s.conf.JoinTarget = res.Jointarget
		s.PreJoinRequest()
	}
}

// PreJoinResponse PreJoinResponse
func (s *server) PreJoin(ctx context.Context, req *pb.PreJoinRequest) (*pb.PreJoinResponse, error) {
	resp := &pb.PreJoinResponse{
		Message:    "join succeeded",
		Bootexpect: int64(s.QuorumSize()),
		Jointarget: s.conf.JoinTarget,
		Curnodes:   s.conf.Peers,
	}
	if s.currentLeaderHost != "" {
		resp.Jointarget = s.currentLeaderHost
	}

	if (s.currentLeaderName != "" && s.State() != Leader) ||
		(s.currentLeaderName == "" && s.conf.JoinTarget != s.conf.Host) {
		resp.Result = -1
		resp.Message = "You should join the boostrap server or leader"
		return resp, nil
	}

	if _, ok := s.peers[req.Name]; ok {
		resp.Result = 1
	} else {
		defer func() { s.AddPeer(req.Name, req.Host) }()
		resp.Result = 0
	}

	// fmt.Printf("%+v\n", resp.Curnodes)
	return resp, nil
}
