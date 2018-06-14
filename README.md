## goRaft 简介

[![Build Status](https://travis-ci.org/moxiaomomo/goRaft.svg?branch=master)](https://travis-ci.org/moxiaomomo/goRaft)
[![Go Report Card](https://goreportcard.com/badge/github.com/moxiaomomo/goraft)](https://goreportcard.com/report/github.com/moxiaomomo/goraft)

#### raft协议的go版本，实现功能包括：

- 选主投票
- 节点心跳
- 日志同步
- 成员变更
- 日志压缩

#### 存在的问题或需要解决的问题：

- 完善snapshot快照同步逻辑
- 成员变更，每次只允许变动一个节点
- 非leader节点转发请求，考虑改用rpc方式
- 改善状态机。。。。。。

#### 测试启动：

在主目录下，启动多个测试节点，比如(测试配置文件中定义了三个节点):

```bash
$ go test -bexpect=2 -name=node1 -host=127.0.0.1:3333 -client=127.0.0.1:3334
{SvrName:node1 SvrHost:127.0.0.1:3333 Client:127.0.0.1:3334 JoinTarget: BootstrapExpect:2}
[INFO]&{LogPrefix:raft-log- CommitIndex:0 Peers:map[node1:127.0.0.1:3333] Host:127.0.0.1:3333 Client:127.0.0.1:3334 Name:node1 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO]state loaded: &{CommitIndex:8 Term:20 VoteFor:}
[INFO]current state:bootstrapping, term:20
[INFO][INFO]extra handlefunc: map[]
[INFO]listen client address: 127.0.0.1:3334
listen internal rpc address: 127.0.0.1:3333
[INFO]current state:candidate, term:20
[INFO]current state:leader, term:21
[INFO]append entries suc: 127.0.0.1:3337 Term:21 Index:8
```

```bash
$ go test -name=node2 -host=127.0.0.1:3335 -client=127.0.0.1:3336 -join=127.0.0.1:3333
{SvrName:node2 SvrHost:127.0.0.1:3335 Client:127.0.0.1:3336 JoinTarget:127.0.0.1:3333 BootstrapExpect:2}
[INFO]&{LogPrefix:raft-log- CommitIndex:0 Peers:map[node2:127.0.0.1:3335] Host:127.0.0.1:3335 Client:127.0.0.1:3336 Name:node2 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO]state loaded: &{CommitIndex:8 Term:20 VoteFor:}
[INFO]current state:bootstrapping, term:20
[INFO]extra handlefunc: map[]
[INFO]listen client address: 127.0.0.1:3336
[INFO]listen internal rpc address: 127.0.0.1:3335
[INFO]current state:candidate, term:20
[INFO]current state:follower, term:21
```

```bash
$ go test -name=node3 -host=127.0.0.1:3337 -client=127.0.0.1:3338 -join=127.0.0.1:3333
{SvrName:node3 SvrHost:127.0.0.1:3337 Client:127.0.0.1:3338 JoinTarget:127.0.0.1:3333 BootstrapExpect:2}
[INFO]&{LogPrefix:raft-log- CommitIndex:0 Peers:map[node3:127.0.0.1:3337] Host:127.0.0.1:3337 Client:127.0.0.1:3338 Name:node3 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO]state loaded: &{CommitIndex:8 Term:21 VoteFor:}
[INFO]current state:bootstrapping, term:21
[INFO]extra handlefunc: map[]
[INFO]listen client address: 127.0.0.1:3338
[INFO]listen internal rpc address: 127.0.0.1:3337
[INFO]current state:candidate, term:21
[INFO]current state:follower, term:21
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
