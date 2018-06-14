package raft

import (
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
		for n, v := range res.Curnodes {
			if _, ok := s.peers[n]; !ok {
				s.peers[n] = &Peer{
					Host:   n,
					Name:   v,
					server: s,
				}
			}
		}
		s.conf.Peers = res.Curnodes
		s.conf.BootstrapExpect = int(res.Bootexpect)
	}
	//fmt.Printf("%+v\n", res.Curnodes)
}

// PreJoinResponse PreJoinResponse
func (s *server) PreJoin(ctx context.Context, req *pb.PreJoinRequest) (*pb.PreJoinResponse, error) {
	if s.conf.JoinTarget != s.conf.Host {
		resp := &pb.PreJoinResponse{
			Result:  -1,
			Message: "im not the boostrap server",
		}
		return resp, nil
	}
	resp := &pb.PreJoinResponse{
		Message: "join succeeded",
	}
	if _, ok := s.peers[req.Name]; ok {
		resp.Result = 1
	} else {
		s.peers[req.Name] = &Peer{
			Host:   req.Host,
			Name:   req.Name,
			server: s,
		}
		s.conf.Peers[req.Name] = req.Host
		resp.Result = 0
	}
	resp.Curnodes = s.conf.Peers
	resp.Bootexpect = int64(s.QuorumSize())
	//fmt.Printf("%+v\n", resp.Curnodes)
	return resp, nil
}
