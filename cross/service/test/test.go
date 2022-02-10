package test

import (
	"cross/dispatch"
	"cross/pack"
	"cross/service/cross"
	t "cross/typedefine"
	"proto"
)

func init() {
	dispatch.RegClientActorMsg(proto.Test, proto.TestCCrossTest, onCrossTest)
}

func onCrossTest(actor *t.CrossActor, reader *pack.Reader) {
	writer := pack.NewWriter()
	writer.Writer(proto.Test, proto.TestSCrossTest, "onCrossTest")
	cross.Reply(actor.ServerId, actor.ActorId, writer.Bytes())
}
