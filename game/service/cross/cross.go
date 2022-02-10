package cross

import (
	"fmt"
	"game/config"
	"game/database/actordao"
	"game/dispatch"
	"game/helper"
	"game/log"
	"game/pack"
	"game/service"
	t "game/typedefine"
	"github.com/nats-io/nats.go"
	"proto"
	"runtime/debug"
	"sync"
	"time"
)

const (
	subBroadcast = "broadcast"
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
	nc       = &nats.Conn{}
	messages = make([]*message, 0)
	msgMux   = &sync.Mutex{}
	wait     = make(chan byte)
	isWait   = false
)

func init() {
	dispatch.ClientCrossActorMsgHandle = clientCrossActorMsgHandle
	service.RegGameStart(onGameStart)
	dispatch.RegCrossMsg(proto.CrossReplyActorMsg, onReplyActor)
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
	// 广播消息
	if _, err = nc.Subscribe(subBroadcast, recvMsg); err != nil {
		panic(err)
	}

	//跨服消息
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

func recvMsg(msg *nats.Msg) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("recvMsg err:%s\n%v", err, string(debug.Stack()))
		}
	}()
	reader := pack.NewReader(msg.Data)
	var serverId, msgId int
	reader.Read(&serverId, &msgId)
	dispatch.PushCrossMsg(serverId, msgId, reader)
}

//推送游戏服消息（可以根据serverId推送到任何服）
func PushGameServerMsg(serverId int, msgId int, data []byte) {
	writer := pack.NewWriter(config.ServerId, msgId, data)
	pushMsg(fmt.Sprintf(pubGame, serverId), writer.Bytes())
}

//推送跨服消息（推送到跨服端）
func PushCrossServerMsg(msgId int, data []byte) {
	writer := pack.NewWriter(msgId, config.ServerId, data)
	pushMsg(fmt.Sprintf(pubCross, config.ServerConfig.CrossServer), writer.Bytes())
}

//玩家消息转发到跨服 (客户端玩家消息转发到跨服端)
func clientCrossActorMsgHandle(sys, cmd int16, actor *t.Actor, reader *pack.Reader) {
	writer := pack.NewWriter(proto.CrossActorMsg, sys, cmd)
	writer.Writer(helper.PacketCrossActor(actor))
	var data = make([]byte, reader.Len())
	reader.Read(data)
	writer.Writer(data)
	pushMsg(fmt.Sprintf(pubCross, config.ServerConfig.CrossServer), writer.Bytes())
}

//推送global服消息
func PushGlobalServerMsg(msgId int, data []byte) {
	writer := pack.NewWriter(config.ServerId, msgId, data)
	pushMsg(pubGlobal, writer.Bytes())
}

//广播消息(所有游戏服都能收到)
func PushBroadcastMsg(msgId int, data []byte) {
	writer := pack.NewWriter(config.ServerId, msgId, data)
	pushMsg(pubBroadcast, writer.Bytes())
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

//回复玩家跨服消息
func onReplyActor(serverId int, reader *pack.Reader) {
	var (
		actorId int64
		sys     int16
		cmd     int16
	)

	reader.Read(&actorId)
	actor := actordao.GetOnlineActor(actorId)
	if actor == nil {
		return
	}
	reader.Read(&sys, &cmd)
	var data = make([]byte, reader.Len())
	reader.Read(data)
	actor.Reply(sys, cmd, data)
}
