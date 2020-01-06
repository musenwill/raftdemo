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
