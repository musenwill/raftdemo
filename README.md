# 介绍
raftdeomo 是用 go 语言写的 raft 分布式共识算法，在单进程中使用 goroutine 模拟网络节点，用 channel 模拟网络通信，实现了 n 节点选举，n 节点间日志同步的功能，并可以在前端页面上进行演示。该项目主要用于自己理解 raft 算法思想以及 go 语言编码练手。当然，后面可以提供实际的网络通信代替 channel 通信实现，也可以容易的扩展为在实际网络环境下的 raft 算法实现。

## build frontend

1. run node.js docker
```
docker run -it --name node -p 9526:9526 -v $PWD:/home node bash
```

2. read frontend/README.md and follow its steps


## build backend

1. run golang docker
```
docker run -it --name golang -v $PWD:/home golang bash
```

## distribution

```
make
cd dist
./raftdemo -h
```

## pprof

```
pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=60
```
