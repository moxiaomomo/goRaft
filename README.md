## goRaft 简介

[![Build Status](https://travis-ci.org/moxiaomomo/goRaft.svg?branch=master)](https://travis-ci.org/moxiaomomo/goRaft)
[![Go Report Card](https://goreportcard.com/badge/github.com/moxiaomomo/goraft)](https://goreportcard.com/report/github.com/moxiaomomo/goraft)

#### raft协议的go版本，实现功能包括：

-  选主投票
- 节点心跳
- 日志同步
- 成员变更
- 日志压缩

#### 存在的问题或需要解决的问题：

- 完善snapshot快照同步逻辑
- 成员变更，每次只允许变动一个节点
- 成员变更日志，leader节点在追加持久化后即可apply
- 非leader节点转发请求，考虑改用rpc方式
- 改善和统一运行日志
- 改善状态机。。。。。。

#### 测试启动：

在主目录下，启动多个测试节点，比如(测试配置文件中定义了三个节点):
```bash
$ go test -confpath /opt/raft/raft_test0.cfg
config: &{LogPrefix:raft-log- CommitIndex:0 PeerHosts:[127.0.0.1:3001 127.0.0.1:3002 127.0.0.1:3000] Host:127.0.0.1:3000 Client:127.0.0.1:4000 Name:server0}
state loaded: &{CommitIndex:2 Term:87 VoteFor:}
current state:follower, term:87
listen internal rpc address: 127.0.0.1:3000
extra handlefunc: map[]
listen client address: 127.0.0.1:4000
current state:candidate, term:87
current state:leader, term:95
To rewrite configuration to persistent storage.
Appendentries succeeded: 127.0.0.1:3003 Success:true Term:95 Index:4 
```
```bash
$ go test -confpath /opt/raft/raft_test1.cfg
config: &{LogPrefix:raft-log- CommitIndex:0 PeerHosts:[127.0.0.1:3002 127.0.0.1:3000 127.0.0.1:3001] Host:127.0.0.1:3001 Client:127.0.0.1:4001 Name:server1}
state loaded: &{CommitIndex:2 Term:87 VoteFor:}
current state:follower, term:87
extra handlefunc: map[]
listen client address: 127.0.0.1:4001
listen internal rpc address: 127.0.0.1:3001
current state:candidate, term:87
current state:follower, term:95
```
```bash
$ go test -confpath /opt/raft/raft_test2.cfg
config: &{LogPrefix:raft-log- CommitIndex:0 PeerHosts:[127.0.0.1:3002 127.0.0.1:3000 127.0.0.1:3001] Host:127.0.0.1:3002 Client:127.0.0.1:4002 Name:server2}
state loaded: &{CommitIndex:2 Term:87 VoteFor:}
current state:follower, term:87
listen internal rpc address: 127.0.0.1:3002
extra handlefunc: map[]
listen client address: 127.0.0.1:4002
```
#### 测试成员变更

启动要加入集群的新节点后，随机向已组成集群的节点发送请求，比如：

```bash
添加节点：
curl "http://localhost:4001/internal/join?name=server3&host=127.0.0.1:3003"
剔除节点：
curl "http://localhost:4002/internal/leave?name=server3&host=127.0.0.1:3003"
```
#### 日志同步状态检查

在raft集群运行过程中，想简单检查日志log/节点状态是否一致，可直接尝试：

```bash
$ sha1sum internlog/raft-log-server*
ae58578bc6513c96eb79714c45cddadbbc2d7eb9  internlog/raft-log-server0
ae58578bc6513c96eb79714c45cddadbbc2d7eb9  internlog/raft-log-server1
ae58578bc6513c96eb79714c45cddadbbc2d7eb9  internlog/raft-log-server2
ae58578bc6513c96eb79714c45cddadbbc2d7eb9  internlog/raft-log-server3

$ sha1sum internlog/state-server*
c77e5ac4aefd215f212e083e32ada9087bd3a7d5  internlog/state-server0
c77e5ac4aefd215f212e083e32ada9087bd3a7d5  internlog/state-server1
c77e5ac4aefd215f212e083e32ada9087bd3a7d5  internlog/state-server2
c77e5ac4aefd215f212e083e32ada9087bd3a7d5  internlog/state-server3
```
