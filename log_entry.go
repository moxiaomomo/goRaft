package raft

import (
	"bufio"
	"encoding/binary"
	//	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/moxiaomomo/goRaft/proto"
	"os"
)

type LogEntryMeta struct {
	DataLength uint32
}

type LogEntry struct {
	Entry       *pb.LogEntry
	LogPosition uint64 // position in logfile
}

func NewLogEntry(curterm, curindex uint64, commandname string, command []byte) *LogEntry {
	lu := &LogEntry{
		Entry: &pb.LogEntry{
			Index:       curindex,
			Term:        curterm,
			Commandname: commandname,
			Command:     command,
		},
	}
	return lu
}

// save data into file
func (l *LogEntry) dump(file *os.File) (indexend int64, err error) {
	n, _ := file.Seek(0, os.SEEK_END)

	w := bufio.NewWriter(file)

	d, err := proto.Marshal(l.Entry)
	if err != nil {
		return -1, err
	}
	data := []byte(d)
	meta := LogEntryMeta{
		DataLength: uint32(len(data)),
	}
	err = binary.Write(w, binary.BigEndian, &meta)
	if err != nil {
		return -1, err
	}

	w.Write(data)
	w.Flush()
	return n + int64(binary.Size(meta)) + int64(meta.DataLength), nil
}

// load data from file
func (l *LogEntry) load(file *os.File, startIndex int64) (indexend int64, err error) {
	n, _ := file.Seek(startIndex, 0)
	r := bufio.NewReader(file)

	meta := &LogEntryMeta{}
	err = binary.Read(r, binary.BigEndian, meta)
	if err != nil {
		return -1, err
	}

	data := make([]byte, meta.DataLength)
	_, err = r.Read(data)
	if err != nil {
		return -1, nil
	}

	l.Entry = &pb.LogEntry{}
	err = proto.Unmarshal(data, l.Entry)
	if err != nil {
		return -1, err
	}
	return n + int64(binary.Size(meta)) + int64(meta.DataLength), nil
}
