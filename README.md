## Run test
`$go test -v ./...`

## start server
`$go run main.go`

## connect redis-cli to redis-clone server
`$redis-cli -h localhost -p 7379`

## conigure git to use hooksPath for hooks
`$git config core.hooksPath .githooks`

<!-- 
implement expiry to expire 40% of keys when limit is hit
implement INFO command
write a script to PUT new key, value in redis db
use a grafana dashboard to monitor the key expiry

NEXT ACTION
- run the python script to insert 100 keys in store
- test INFO and FLUSHDB command from terminal
- connect redis exporter to dicedb
- view grafana dashboard

- refactor the code to separate circular dependency between eval and store
 -->