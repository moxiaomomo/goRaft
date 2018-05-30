package raft

import (
	"io"
)

// JoinCommand join command interface
type JoinCommand interface {
	Command
	NodeName() string
}

// DefaultJoinCommand default join command implementation
type DefaultJoinCommand struct {
	Name string `json:"name"`
	Host string `json:"connectioninfo"`
}

// LeaveCommand leave command interface
type LeaveCommand interface {
	Command
	NodeName() string
}

// DefaultLeaveCommand default leave command implementation
type DefaultLeaveCommand struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// NOPCommand no-operation command struct
type NOPCommand struct {
}

// CommandName returns DefaultJoinCommand's name
func (c *DefaultJoinCommand) CommandName() string {
	return "raft:join"
}

// Apply applys DefaultJoinCommand
func (c *DefaultJoinCommand) Apply(server Server) (interface{}, error) {
	err := server.AddPeer(c.Name, c.Host)
	return []byte("join"), err
}

// NodeName return the name of the node to join
func (c *DefaultJoinCommand) NodeName() string {
	return c.Name
}

// CommandName returns DefaultLeaveCommand's name
func (c *DefaultLeaveCommand) CommandName() string {
	return "raft:leave"
}

// Apply applys DefaultLeaveCommand
func (c *DefaultLeaveCommand) Apply(server Server) (interface{}, error) {
	err := server.RemovePeer(c.Name, c.Host)
	return []byte("leave"), err
}

// NodeName return the name of the node to leave
func (c *DefaultLeaveCommand) NodeName() string {
	return c.Name
}

// CommandName returns NOPCommand's name
func (c NOPCommand) CommandName() string {
	return "raft:nop"
}

// Apply just do nothing
func (c NOPCommand) Apply(server Server) (interface{}, error) {
	return nil, nil
}

// Encode encode nop-command data
func (c NOPCommand) Encode(w *io.Writer) error {
	return nil
}

// Decode decode nop-command data
func (c NOPCommand) Decode(r io.Reader) error {
	return nil
}
