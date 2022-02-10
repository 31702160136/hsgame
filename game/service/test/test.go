package test

import (
	"game/dispatch"
	"game/pack"
	t "game/typedefine"
	"proto"
)

/*
	功能示例
*/
func init() {
	dispatch.RegClientActorMsg(proto.Test, proto.TestCTest, onTest)
	dispatch.RegClientCrossActorMsg(proto.Test, proto.TestCCrossTest)
}

func onTest(actor *t.Actor, reader *pack.Reader) {
	actor.Reply(proto.Test, proto.TestSTest, "onTest")
}
