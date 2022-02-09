package dispatch

import (
	"fmt"
	"game/base"
	"game/log"
	"game/pack"
	t "game/typedefine"
	"proto"
	"runtime/debug"
	"sync"
	"time"
)

type message struct {
	Name string
	Type byte
	Func interface{}
	Args []interface{}
}

const (
	TypeClientAccountMsg byte = byte(proto.System) //账号消息
	TypeClientActorMsg   byte = 2                  //角色消息
	TypeSystemMsg        byte = 3                  //系统消息
	TypeSystemGoMsg      byte = 4                  //系统线程消息
	TypeCrossMsg         byte = 5                  //跨服消息
)

var (
	dispatchMsg    = make([]*message, 0)
	dispatchMsgMux = sync.Mutex{}
	wait           = make(chan byte)
	isWait         = false

	clientActorMsg   = make(map[int16]map[int16]func(actor *t.Actor, reader *pack.Reader))
	clientAccountMsg = make(map[int16]func(account *t.Account, reader *pack.Reader))
)

func OnRunGame() {
	go func() {
		for {
			msgs := readMsgs()
			for _, msg := range msgs {
				dispatch(msg)
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

//push消息
func PushClientMessage(sys, cmd int16, args ...interface{}) {
	var cbFunc interface{}
	msgType := byte(0)
	if sys == int16(TypeClientAccountMsg) {
		cbFun, ok := clientAccountMsg[cmd]
		if !ok {
			log.Fatalf("%d-%d not found\n%s", sys, cmd, string(debug.Stack()))
			return
		}
		cbFunc = cbFun
		msgType = TypeClientAccountMsg
	} else {
		cbFun, ok := clientActorMsg[sys][cmd]
		if !ok {
			log.Fatalf("%d-%d not found\n%s", sys, cmd, string(debug.Stack()))
			return
		}
		cbFunc = cbFun
		msgType = TypeClientActorMsg
	}

	pushMessage(fmt.Sprintf("proto %d-%d", sys, cmd), msgType, cbFunc, args...)
}

//push消息
func pushMessage(name string, msgType byte, cbFun interface{}, args ...interface{}) {
	msg := &message{}
	msg.Name = name
	msg.Type = msgType
	msg.Func = cbFun
	msg.Args = args
	dispatchMsgMux.Lock()
	dispatchMsg = append(dispatchMsg, msg)
	dispatchMsgMux.Unlock()
	if isWait {
		isWait = false
		wait <- 1
	}
}

//派遣处理消息
func dispatch(msg *message) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("%s %v\n%s", msg.Name, err, string(debug.Stack()))
		}
	}()
	switch msg.Type {
	case TypeClientAccountMsg:
		base.CallReflectFunc(msg.Func, msg.Args...)
	case TypeClientActorMsg:
		base.CallReflectFunc(msg.Func, msg.Args...)
	case TypeSystemMsg:
		base.CallReflectFunc(msg.Func, msg.Args...)
	case TypeSystemGoMsg:
		go base.CallReflectFunc(msg.Func, msg.Args...)
	}
}

//读取消息
func readMsgs() []*message {
	dispatchMsgMux.Lock()
	msgs := dispatchMsg
	dispatchMsg = dispatchMsg[len(dispatchMsg):]
	if len(msgs) == 0 {
		isWait = true
	}
	dispatchMsgMux.Unlock()
	if isWait {
		<-wait
	}
	return msgs
}

//注册客户端角色消息
func RegClientActorMsg(sys, cmd int16, fun func(actor *t.Actor, reader *pack.Reader)) {
	if clientActorMsg[sys] == nil {
		clientActorMsg[sys] = map[int16]func(actor *t.Actor, reader *pack.Reader){}
	}
	clientActorMsg[sys][cmd] = fun
}

//注册客户端账号消息
func RegClientAccountMsg(cmd int16, fun func(account *t.Account, reader *pack.Reader)) {
	clientAccountMsg[cmd] = fun
}

//推送系统异步消息
/*
	param
	1.消息名称
	2.回调方法
	3.异步方法
	4.参数
*/
func PushSystemAsyncMsg(msgName string, cbFunc interface{}, asyncFunc interface{}, args ...interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("pushSystemAsyncMsg: %s %v\n%s", msgName, err, string(debug.Stack()))
			}
		}()
		vals := base.CallReflectFunc(asyncFunc, args...)
		newArgs := make([]interface{}, len(vals))
		for i, _ := range newArgs {
			newArgs[i] = vals[i].Interface()
		}
		if cbFunc != nil {
			pushMessage("pushSystemAsyncMsg: "+msgName, TypeSystemMsg, cbFunc, newArgs...)
		}
	}()
}

//推送系统同步消息
func PushSystemSyncMsg(msgName string, cbFunc interface{}, args ...interface{}) {
	pushMessage("pushSystemSyncMsg: "+msgName, TypeSystemMsg, cbFunc, args...)
}

//推送系统线程消息
func PushSystemGoMsg(msgName string, cbFunc interface{}, args ...interface{}) {
	pushMessage("pushSystemGoMsg: "+msgName, TypeSystemGoMsg, cbFunc, args...)
}
