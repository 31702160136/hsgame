package cross

import (
	"cross/config"
	"cross/dispatch"
	"cross/log"
	"cross/pack"
	"cross/service"
	"fmt"
	"github.com/nats-io/nats.go"
	"proto"
	"runtime/debug"
	"sync"
	"time"
)

const (
	subCross = "cross_%d"

	pubGame      = "game_%d"
	pubGlobal    = "global"
	pubBroadcast = "broadcast"
)

type message struct {
	sub  string
	data []byte
}

var (
	nc       = &nats.Conn{}
	messages = make([]*message, 0)
	msgMux   = &sync.Mutex{}
	wait     = make(chan byte)
	isWait   = false
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
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			//断开回调
			log.Info("----------------Nats Discovered---------------")
			log.Info(err.Error())
			log.Error("----------------Nats Discovered---------------")
			log.Error(err.Error())
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			//重连回调
			log.Info("----------------Nats Reconnect---------------")
		}),
	)

	if err != nil {
		panic(err)
	}

	if _, err = nc.Subscribe(fmt.Sprintf(subCross, config.ServerConfig.ServerId), recvMsg); err != nil {
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
	msgs := messages
	messages = messages[len(messages):]
	if len(msgs) == 0 {
		isWait = true
	}
	msgMux.Unlock()
	if isWait {
		<-wait
	}
	return msgs
}

//派遣接收到的消息
func recvMsg(msg *nats.Msg) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("recvMsg err:%s\n%v", err, string(debug.Stack()))
		}
	}()
	reader := pack.NewReader(msg.Data)
	var msgId int
	reader.Read(&msgId)
	if msgId == proto.CrossActorMsg {
		var sys, cmd int16
		reader.Read(&sys, &cmd)
		dispatch.PushClientActorMessage(sys, cmd, reader)
	} else {
		var serverId int
		reader.Read(&serverId)
		dispatch.PushCrossMessage(msgId, serverId, reader)
	}
}

//回复玩家消息
func Reply(serverId int, actorId int64, data []byte) {
	PushGameServerMsg(serverId, proto.CrossReplyActorMsg, pack.NewWriter(actorId, data).Bytes())
}

//推送游戏服消息
func PushGameServerMsg(serverId int, msgId int, data []byte) {
	pushMsg(fmt.Sprintf(pubGame, serverId), pack.NewWriter(config.ServerId, msgId, data).Bytes())
}

//推送global服消息
func PushGlobalServerMsg(id int, data []byte) {
	pushMsg(pubGlobal, pack.NewWriter(id, data).Bytes())
}

//广播消息(所有游戏服都能收到)
func PushBroadcastMsg(id int, data []byte) {
	pushMsg(pubBroadcast, pack.NewWriter(id, data).Bytes())
}

//推送消息
func pushMsg(sub string, data []byte) {
	msgMux.Lock()
	msg := &message{
		sub:  sub,
		data: data,
	}
	messages = append(messages, msg)
	msgMux.Unlock()
	if isWait {
		isWait = false
		wait <- 1
	}
}
