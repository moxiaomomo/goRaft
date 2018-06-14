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
	} else {
		for name, host := range res.Curnodes {
			if _, ok := s.peers[name]; !ok {
				ti := time.Duration(s.heartbeatInterval) * time.Millisecond
				s.peers[name] = NewPeer(s, name, host, ti)
			}
		}
		s.conf.Peers = res.Curnodes
		s.conf.BootstrapExpect = int(res.Bootexpect)
		s.currentLeaderName = res.Leadername
	}
	logger.Infof("prejoin result:%+v\n", res)
}

// PreJoinResponse PreJoinResponse
func (s *server) PreJoin(ctx context.Context, req *pb.PreJoinRequest) (*pb.PreJoinResponse, error) {
	if s.conf.JoinTarget != s.conf.Host && s.State() != Leader {
		resp := &pb.PreJoinResponse{
			Result:  -1,
			Message: "im not the boostrap server or leader",
		}
		return resp, nil
	}
	resp := &pb.PreJoinResponse{
		Message: "join succeeded",
	}
	if _, ok := s.peers[req.Name]; ok {
		resp.Result = 1
	} else {
		defer func() { s.AddPeer(req.Name, req.Host) }()
		resp.Result = 0
	}
	resp.Curnodes = s.conf.Peers
	resp.Bootexpect = int64(s.QuorumSize())
	resp.Leadername = s.currentLeaderName

	// fmt.Printf("%+v\n", resp.Curnodes)
	// fmt.Printf("%+v %+v\n", s.peers, s.conf.Peers)
	return resp, nil
}
