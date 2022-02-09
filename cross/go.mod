module cross

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/nats-io/jwt v0.3.2 // indirect
	github.com/nats-io/nats.go v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20220208050332-20e1d8d225ab // indirect
	proto v0.0.0
)

replace (
	proto => ../proto
)