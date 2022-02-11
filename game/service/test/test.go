package test

import (
	"game/database/actordao"
	"game/dispatch"
	"game/pack"
	"game/service/cross"
	t "game/typedefine"
	"proto"
)

/*
	功能示例
*/
func init() {
	dispatch.RegClientActorMsg(proto.Test, proto.TestCTest, onTest)
	dispatch.RegClientActorMsg(proto.Test, proto.TestCCrossTest2, onCrossTest)
	dispatch.RegClientActorMsg(proto.Test, proto.TestCBroadcast, onBroadcast)
	dispatch.RegClientCrossActorMsg(proto.Test, proto.TestCCrossTest)

	//内部跨服
	dispatch.RegCrossMsg(proto.CrossTest, onTest2)
}

func onTest(actor *t.Actor, reader *pack.Reader) {
	actor.Reply(proto.Test, proto.TestSTest, "onTest")
}
func onCrossTest(actor *t.Actor, reader *pack.Reader) {
	cross.PushGameServerMsg(actor.ServerId, proto.CrossTest, pack.NewWriter(actor.ActorId).Bytes())
}

func onTest2(serverId int, reader *pack.Reader) {
	var actorId int64
	reader.Read(&actorId)
	actor := actordao.GetOnlineActor(actorId)
	if actor == nil {
		return
	}
	actor.Reply(proto.Test, proto.TestSCrossTest2, "success")
}

func onBroadcast(actor *t.Actor, reader *pack.Reader) {
	cross.PushBroadcastMsg(proto.CrossTest, pack.GetBytes(actor.ActorId))
}
