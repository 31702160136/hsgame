module game

go 1.16

require (
	github.com/gin-gonic/gin v1.7.7
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gorilla/websocket v1.4.2
	github.com/json-iterator/go v1.1.12
	github.com/nats-io/nats-server/v2 v2.7.2 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	google.golang.org/protobuf v1.27.1 // indirect
	proto v0.0.0
)

replace proto => ../proto
