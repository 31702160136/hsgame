package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"opt/log"
	"time"
)

type Config struct {
	Host       string
	Port       int
	DBName     string
	AuthSource string
	UserName   string
	Password   string
}

/*
	新建数据库连接
	@param cnf Config 配置信息
*/
func New(cnf Config) *mongo.Database {
	uri := fmt.Sprintf("mongodb://%s:%d", cnf.Host, cnf.Port)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	opt := options.Client().ApplyURI(uri)
	if cnf.AuthSource != "" {
		opt.Auth = &options.Credential{
			AuthSource: cnf.AuthSource,
			Username:   cnf.UserName,
			Password:   cnf.Password,
		}
	} else {
		opt = nil
	}
	client, clientErr := mongo.Connect(ctx, opt)
	if clientErr != nil {
		panic(clientErr)
	}
	// Check the connection
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	db := client.Database(cnf.DBName)
	return db
}

func GetId(objectId interface{}) string {
	if id, ok := objectId.(primitive.ObjectID); ok {
		return id.Hex()
	} else {
		return ""
	}
}

func GetObjId(id string) primitive.ObjectID {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error("GetObjId error", err.Error())
		return [12]byte{}
	}
	return objId
}

func Get() {

}
