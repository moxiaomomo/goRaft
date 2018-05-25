package raft

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Log struct {
	ApplyFunc func(*LogEntry, Command) (interface{}, error)

	mutex     sync.RWMutex
	file      *os.File
	path      string
	entries   []*LogEntry
	toc_entry LogEntry

	startIndex  uint64
	commitIndex uint64
	initialized bool
}

func newLog() *Log {
	log := &Log{
		entries:     make([]*LogEntry, 0),
		toc_entry:   LogEntry{},
		commitIndex: 0,
		startIndex:  0,
	}
	return log
}

func (l *Log) LastCommitInfo() (index uint64, term uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	cmiLen := int(l.commitIndex - l.startIndex + 1)
	if cmiLen <= 0 || len(l.entries) == 0 || cmiLen > len(l.entries) {
		return 0, 0
	}

	last := l.entries[cmiLen-1]
	return last.Entry.Index, last.Entry.Term
}

func (l *Log) IsEmpty() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.entries) == 0
}

func (l *Log) PreLastLogInfo() (index uint64, term uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) <= 1 {
		return 0, 0
	}

	last := l.entries[len(l.entries)-2]
	return last.Entry.Index, last.Entry.Term
}

func (l *Log) FirstLogIndex() uint64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return 0
	}

	return l.entries[0].Entry.Index
}

func (l *Log) LastLogIndex() uint64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return 0
	}

	last := l.entries[len(l.entries)-1]
	return last.Entry.Index
}

func (l *Log) LastLogInfo() (index uint64, term uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return 0, 0
	}

	last := l.entries[len(l.entries)-1]
	return last.Entry.Index, last.Entry.Term
}

func (l *Log) GetLogEntries(s, e int) []*LogEntry {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return []*LogEntry{}
	}
	if s < 0 {
		s = 0
	}
	if e > len(l.entries) {
		e = len(l.entries)
	}
	return l.entries[s:e]
}

func (l *Log) LogInit(path string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var err error
	l.file, err = os.OpenFile(path, os.O_RDWR, 0600)
	l.path = path

	if err != nil {
		if os.IsNotExist(err) {
			l.file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
			if err == nil {
				l.initialized = true
			}
			return err
		}
		return err
	}

	startIndex := int64(0)
	for {
		lunit := &LogEntry{}
		startIndex, err = lunit.load(l.file, startIndex)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("Failed to recover raft.log: %v", err)
			}
		}
		l.entries = append(l.entries, lunit)
	}

	if len(l.entries) > 0 {
		l.startIndex = l.entries[0].Entry.GetIndex()
	}
	l.initialized = true
	return nil
}

func (l *Log) AppendEntry(lu *LogEntry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, err := lu.dump(l.file)
	if err != nil {
		return err
	}
	l.entries = append(l.entries, lu)
	return nil
}

func (l *Log) RefreshLog() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.file.Truncate(0) // clear file content
	for _, lu := range l.entries {
		_, err := lu.dump(l.file)
		if err != nil {
			return err
		}
	}
	if len(l.entries) > 0 {
		l.startIndex = l.entries[0].Entry.GetIndex()
	}
	return nil
}

func (l *Log) UpdateCommitIndex(index uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if index <= l.commitIndex {
		return
	}
	if len(l.entries) == 0 {
		fmt.Println("Local log is empty, not to update commitindex.")
		return
	}

	lastindex := l.entries[len(l.entries)-1].Entry.GetIndex()
	if index > lastindex {
		fmt.Printf("Local log is too old or index to commit is invalid,%d:%d\n", index, lastindex)
		return
	}

	l.commitIndex = index
}

func (l *Log) CommitIndex() uint64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.commitIndex
}

func (l *Log) LogEntriesToSync(fromterm uint64) []*LogEntry {
	lulist := []*LogEntry{}
	for _, l := range l.entries {
		if fromterm > 0 && l.Entry.Term <= fromterm {
			continue
		}
		lulist = append(lulist, l)
	}
	return lulist
}
