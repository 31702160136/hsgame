package gameserver

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"opt/database"
	"opt/log"
	"opt/service"
	t "opt/typedefine"
)

const dbname = "game_server"

func this() *mongo.Collection {
	return database.GetDB().Collection(dbname)
}
func init() {
	indexs := map[string]int{
		"server_id": -1,
	}
	service.RegDatabaseInit(func() {
		_ = database.GetDB().CreateCollection(context.TODO(), dbname)
		database.CreateIndex(this(), indexs)
	})
}

func Create(ctx context.Context, obj *t.GameServer) error {
	result, err := this().InsertOne(ctx, obj)
	if err != nil {
		return err
	}
	obj.Id = database.GetId(result.InsertedID)
	return nil
}

func GetGameServer(ctx context.Context, filter bson.D, opt *options.FindOptions) []*t.GameServer {
	data := make([]*t.GameServer, 0)
	cur, err := this().Find(ctx, filter, opt)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	err = cur.All(ctx, &data)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return data
}
