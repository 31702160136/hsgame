package account

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"opt/database"
	"opt/log"
	"opt/service"
	t "opt/typedefine"
)

const dbname = "account"

func this() *mongo.Collection {
	return database.GetDB().Collection(dbname)
}
func init() {
	indexs := map[string]int{
		"account": -1,
	}
	service.RegDatabaseInit(func() {
		_ = database.GetDB().CreateCollection(context.TODO(), dbname)
		database.CreateIndex(this(), indexs)
	})
	service.RegGet("accountTest", func(ctx *gin.Context) {
		//GetAllAccounts(context.TODO())
	})
}

func Create(ctx context.Context, obj *t.Account) error {
	result, err := this().InsertOne(ctx, obj)
	if err != nil {
		return err
	}
	obj.Id = database.GetId(result.InsertedID)
	return nil
}

//更新账号信息
func UpdateAccountInfo(ctx context.Context, info *t.Account) error {
	objId := database.GetObjId(info.Id)
	info.Id = ""
	_, err := this().UpdateOne(ctx, bson.M{"_id": objId}, bson.M{"$set": info})
	return err
}

func GetAccountInfoByAccount(ctx context.Context, account string) *t.Account {
	data := &t.Account{}
	res := this().FindOne(ctx, bson.M{"account": account})
	err := res.Decode(data)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return data
}

//返回所有账号
func GetAllAccounts(ctx context.Context) []string {
	key := "account"
	opt := options.FindOptions{}
	opt.SetProjection(bson.M{key: 1})
	cur, err := this().Find(ctx, bson.M{}, &opt)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	data := make([]string, cur.RemainingBatchLength())
	index := 0
	for cur.Next(ctx) {
		data[index] = cur.Current.Lookup(key).StringValue()
		index++
	}
	return data
}
