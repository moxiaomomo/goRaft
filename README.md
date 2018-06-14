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
[INFO] [{node1 127.0.0.1:3333 127.0.0.1:3334  1}]
[INFO] &{LogPrefix:raft-log- Peers:map[node1:127.0.0.1:3333] Host:127.0.0.1:3333 Client:127.0.0.1:3334 Name:node1 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO] state loaded: &{CommitIndex:0 Term:1 VoteFor:}
[INFO] current state:bootstrapping, term:1
[INFO] listen internal rpc address: 127.0.0.1:3333
[INFO] extra handlefunc: map[]
[INFO] listen client address: 127.0.0.1:3334
[INFO] current state:follower, term:1
[INFO] current state:candidate, term:1
[INFO] current state:follower, term:1
```

```bash
$ go test -name=node2 -host=127.0.0.1:3335 -client=127.0.0.1:3336 -join=127.0.0.1:3333
[INFO] [{node2 127.0.0.1:3335 127.0.0.1:3336 127.0.0.1:3333 1}]
[INFO] &{LogPrefix:raft-log- Peers:map[node2:127.0.0.1:3335] Host:127.0.0.1:3335 Client:127.0.0.1:3336 Name:node2 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO] state loaded: &{CommitIndex:0 Term:0 VoteFor:}
[INFO] current state:bootstrapping, term:0
[INFO] listen internal rpc address: 127.0.0.1:3335
[INFO] extra handlefunc: map[]
[INFO] listen client address: 127.0.0.1:3336
[INFO] prejoin result:bootexpect:2 message:"join succeeded" jointarget:"127.0.0.1:3333" curnodes:<key:"node1" value:"127.0.0.1:3333" > curnodes:<key:"node2" value:"127.0.0.1:3335" > 
[INFO] current state:follower, term:0
[INFO] current state:candidate, term:0
[INFO] current state:leader, term:1
[INFO] append entries suc: 127.0.0.1:3333 Term:1 
[INFO] append entries suc: 127.0.0.1:3333 Success:true Term:1 Index:1 
```

```bash
$ go test -name=node3 -host=127.0.0.1:3337 -client=127.0.0.1:3338 -join=127.0.0.1:3333
[INFO] [{node3 127.0.0.1:3337 127.0.0.1:3338 127.0.0.1:3333 1}]
[INFO] &{LogPrefix:raft-log- Peers:map[node3:127.0.0.1:3337] Host:127.0.0.1:3337 Client:127.0.0.1:3338 Name:node3 BootstrapExpect:2 JoinTarget:127.0.0.1:3333}
[INFO] current state:bootstrapping, term:0
[INFO] extra handlefunc: map[]
[INFO] listen client address: 127.0.0.1:3338
[INFO] listen internal rpc address: 127.0.0.1:3337
[INFO] prejoin result:result:-1 bootexpect:2 message:"you should to join the boostrap server or leader" jointarget:"127.0.0.1:3335" curnodes:<key:"node1" value:"127.0.0.1:3333" > curnodes:<key:"node2" value:"127.0.0.1:3335" > 
[INFO] prejoin result:bootexpect:2 message:"join succeeded" jointarget:"127.0.0.1:3335" curnodes:<key:"node1" value:"127.0.0.1:3333" > curnodes:<key:"node2" value:"127.0.0.1:3335" > curnodes:<key:"node3" value:"127.0.0.1:3337" > 
[INFO] current state:follower, term:0
```

#### 测试成员变更

启动要加入集群的新节点后，可以通过外部随机向已组成集群的节点发送请求，比如：

```bash
添加节点：(一般节点启动时会首先通过 -join=xxx 方式来加入集群)
curl "http://localhost:3334/intern/join?name=server3&host=127.0.0.1:3003"

剔除节点：
curl "http://localhost:3334/intern/leave?name=server3&host=127.0.0.1:3003"
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
