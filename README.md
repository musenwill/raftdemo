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
