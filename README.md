## Run test
`$go test -v ./...`

## start server
`$go run main.go`

## connect redis-cli to redis-clone server
`$redis-cli -h localhost -p 7379`

## conigure git to use hooksPath for hooks
`$git config core.hooksPath .githooks`

## connect a locally running redis-exporter
`docker run -d -p 9121:9121 --name redis-exporter oliver006/redis_exporter --redis.addr=redis://host.docker.internal:7379`
`curl http://localhost:9121/metrics`

<!-- 
- refactor the code to separate circular dependency between eval and store
- wrap handling of eStatus in signal_handling and expose functions
- implement `sleep` command
- write tests
 -->