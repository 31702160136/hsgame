package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"opt/config"
)

var (
	db *mongo.Database
)

func GetDB() *mongo.Database {
	return db
}

func InitDataBase() {
	cnf := Config{}
	cnf.Host = config.Config.DBHost
	cnf.Port = config.Config.DBPort
	cnf.DBName = config.Config.DBName
	cnf.AuthSource = config.Config.DBAuthSource
	cnf.UserName = config.Config.DBUserName
	cnf.Password = config.Config.DBPassword
	db = New(cnf)
}

func Transaction(fun func(session context.Context) error) error {
	return db.Client().UseSession(context.TODO(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}
		err = fun(sessionContext)
		if err != nil {
			_ = sessionContext.AbortTransaction(sessionContext)
			return err
		}
		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			return err
		}
		return nil
	})
}

func Find(ctx context.Context, this *mongo.Collection, filters bson.D, data interface{}, count *int64, opt ...*options.FindOptions) error {
	cur, err := this.Find(ctx, filters, opt...)
	if err != nil {
		return err
	}
	err = cur.All(ctx, data)
	if err != nil {
		return err
	}
	*count, err = this.CountDocuments(ctx, filters)
	if err != nil {
		return err
	}
	return err
}

func GetObjectIds(ids []string) []primitive.ObjectID {
	list := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil
		}
		list[i] = objId
	}
	return list
}

func CreateIndex(this *mongo.Collection, index map[string]int) {
	if len(index) == 0 {
		return
	}
	indexs := make([]mongo.IndexModel, 0)
	for k, i := range index {
		opt := &options.IndexOptions{}
		opt.SetUnique(true)
		indexs = append(indexs, mongo.IndexModel{Keys: bson.M{
			k: i,
		}, Options: opt})
	}
	_, err := this.Indexes().CreateMany(context.TODO(), indexs)
	if err != nil {
		panic(err)
	}
}
