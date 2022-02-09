package cross

import (
	"fmt"
	"game/config"
	"game/log"
	"game/service"
	"github.com/nats-io/nats.go"
	"sync"
	"time"
)

const (
	subBroadcast = "broadcast"
	subCross     = "cross_%d"
	subGame      = "game_%d"

	pubCross     = "cross_%d"
	pubGame      = "game_%d"
	pubGlobal    = "global"
	pubBroadcast = "broadcast"
)

type message struct {
	sub  string
	data []byte
}

var (
	nc     = &nats.Conn{}
	msgs   = make([]*message, 0)
	msgMux = &sync.Mutex{}
	wait   = make(chan byte)
	isWait = false
)

func init() {
	service.RegGameStart(onGameStart)
}

func onGameStart() {
	// 连接Nats服务器
	var err error
	nc, err = nats.Connect(
		config.ServerConfig.Nats,
		nats.ReconnectWait(time.Second), //重连等待
		nats.MaxReconnects(-1),
		nats.DiscoveredServersHandler(func(conn *nats.Conn) {
			//断开回调
			log.Info("----------------Nats Discovered---------------")
			log.Error("----------------Nats Discovered---------------")
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			//重连回调
			log.Info("----------------Nats Reconnect---------------")
		}),
	)

	if err != nil {
		panic(err)
	}
	// 广播消息
	if _, err = nc.Subscribe(subBroadcast, recvMsg); err != nil {
		panic(err)
	}

	// 跨服消息
	if _, err = nc.Subscribe(fmt.Sprintf(subCross, config.ServerConfig.CrossGroup), recvMsg); err != nil {
		panic(err)
	}

	// 游戏服消息
	if _, err = nc.Subscribe(fmt.Sprintf(subGame, config.ServerId), recvMsg); err != nil {
		panic(err)
	}

	onRun()
}

func onRun() {
	go func() {
		for {
			msgs := readMsgs()
			for _, msg := range msgs {
				nc.Publish(msg.sub, msg.data)
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

func readMsgs() []*message {
	msgMux.Lock()
	msgs := msgs
	msgs = msgs[len(msgs):]
	if len(msgs) == 0 {
		isWait = true
	}
	msgMux.Unlock()
	if isWait {
		<-wait
	}
	return msgs
}

func recvMsg(msg *nats.Msg) {

}

//推送游戏服消息
func PushGameServerMsg(serverId int, id int, data []byte) {
	pushMsg(fmt.Sprintf(pubGame, serverId), id, data)
}

//推送跨服消息
func PushCrossServerMsg(id int, data []byte) {
	pushMsg(fmt.Sprintf(pubCross, config.ServerConfig.CrossGroup), id, data)
}

//推送global服消息
func PushGlobalServerMsg(id int, data []byte) {
	pushMsg(pubGlobal, id, data)
}

//广播消息(所有游戏服都能收到)
func PushBroadcastMsg(id int, data []byte) {
	pushMsg(pubBroadcast, id, data)
}

//推送消息
func pushMsg(sub string, id int, data []byte) {
	msgMux.Lock()
	msg := &message{
		sub:  sub,
		data: data,
	}
	msgs = append(msgs, msg)
	msgMux.Unlock()
	if isWait {
		isWait = false
		wait <- 1
	}
}
