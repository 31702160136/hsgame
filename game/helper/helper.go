package helper

import (
	"game/config"
	t "game/typedefine"
)

func PacketCrossActor(actor *t.Actor) *t.CrossActor {
	return &t.CrossActor{
		ActorId:  actor.ActorId,
		ServerId: config.ServerId,
	}
}
