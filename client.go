package raft

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moxiaomomo/goRaft/util/logger"
)

func redirect(w http.ResponseWriter, r *http.Request, s *server) {
	url := fmt.Sprintf("http://%s%s", s.currentLeaderExHost, r.RequestURI)
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	_, err = client.Do(req)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Member changed OK.\n"))
	}
}

// GetLeaderHandler handles the get-leader request from clients
func GetLeaderHandler(w http.ResponseWriter, r *http.Request, s *server) {
	ret := map[string]string{
		"leadername":   s.currentLeaderName,
		"leaderhost":   s.currentLeaderHost,
		"leaderexhost": s.currentLeaderExHost,
	}
	b, err := json.Marshal(ret)
	if err != nil {
		w.Write([]byte("Internel server error!\n"))
	} else {
		w.Write([]byte(b))
	}
}

// JoinHandler handles the join request from clients
func JoinHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.AddPeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// pass this request to the leader
		redirect(w, r, s)
	}
}

// LeaveHandler handles the leave request from clients
func LeaveHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.RemovePeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// pass this request to the leader
		redirect(w, r, s)
	}
}

func (s *server) RegisterHandler(urlpath string, fc HandleFuncType) {
	if urlpath == "" || fc == nil {
		return
	}
	s.handlefunc[urlpath] = fc
}

func (s *server) StartExternServe() {
	for url := range s.handlefunc {
		http.HandleFunc(url, s.handlefunc[url])
	}
	logger.Infof("extra handlefunc: %+v\n", s.handlefunc)

	http.HandleFunc("/intern/join", func(w http.ResponseWriter, r *http.Request) { JoinHandler(w, r, s) })
	http.HandleFunc("/intern/leave", func(w http.ResponseWriter, r *http.Request) { LeaveHandler(w, r, s) })
	http.HandleFunc("/intern/getleader", func(w http.ResponseWriter, r *http.Request) { GetLeaderHandler(w, r, s) })

	logger.Infof("listen client address: %s\n", s.conf.Client)
	http.ListenAndServe(s.conf.Client, nil)
}
