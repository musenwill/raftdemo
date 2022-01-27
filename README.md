# introduction
A demo implemention of raft distribution algorithm, demonstrate how election, append entries happens.

# build

1. just build
```
make
```

2. code check and run unit tests
```
make check
```

# quick start

## run server
Start a raft instance with 5 nodes, listen on port 8086 by default.
```
./rafts -node 5
```

## run client
Run raftc to connect localhost:8086 by default
```
./raftc
  app name: rafts
   version: v1.0
    branch: premaster
    commit: af444a7
build time: 2022-01-25T20:52:39+0800
   up time: 2022-01-25 21:06:07.0092706 +0800 CST
```

Get help manual.
```
> help
PING
HELP
EXIT or QUIT
SHOW NODES [WHERE NODE = <nodeID>]
SHOW LEADER
SET NODE <nodeID> STATE <state>
SHOW CONFIG
SET LOG LEVEL <level>
SHOW LOGS [WHERE NODE = <nodeID> AND INDEX = <index>]
WRITE LOG <data>
SHOW PIPES
SET PIPE FROM <nodeID> TO <nodeID> <ok/broken> [, FROM <nodeID> TO <nodeID> <ok/broken>]
```

Show all nodes status.
```
show nodes
 name: nodes
total: 5
  id   term  state   commitID lastAppliedID lastLogID lastLogTerm leader voteFor readable
------ ---- -------- -------- ------------- --------- ----------- ------ ------- --------
node01 1    follower 1        1             1         1           node02 node02  true
node02 1    leader   1        1             1         1                          true
node03 1    follower 1        1             1         1           node02 node02  true
node04 1    follower 1        1             1         1           node02 node02  true
node05 1    follower 1        1             1         1           node02 node02  true
```

Write data.
```
write log "the first log"
write log "the second log"
```

Show data.
```
show logs
 name: entries
total: 3
version id term type    payload
------- -- ---- ---- --------------
0       1  1    Nop
0       2  1    Data the first log
0       3  1    Data the second log
```

# features
1. Support start a raft instance with any number of nodes.
2. Support write data info raft instance.
3. Support change the state of any node by client.
4. Support modify network topology between raft nodes.

# License
See [LICENSE](LICENSE).
