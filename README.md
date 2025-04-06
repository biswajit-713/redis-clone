## Run test
`$go test -v ./...`

## start server
`$go run main.go`

## connect redis-cli to redis-clone server
`$redis-cli -h localhost -p 7379`

## conigure git to use hooksPath for hooks
`$git config core.hooksPath .githooks`