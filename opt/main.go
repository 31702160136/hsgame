package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"opt/config"
	"opt/database"
	"opt/gtime"
	"opt/log"
	"opt/service"
	_ "opt/service/account"
	_ "opt/service/server"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func init() {
	service.RegGet("test", func(ctx *gin.Context) {

	})
}

func main() {
	//加载基础配置
	configName := "config.json"
	bt, err := ioutil.ReadFile(configName)
	if err != nil {
		panic(configName + "," + err.Error())
	}
	err = jsoniter.Unmarshal(bt, &config.Config)
	if err != nil {
		panic(configName + "," + err.Error())
	}
	log.InitLog(config.Config.Log)
	log.Info("load database...")
	//初始化数据库
	database.InitDataBase()
	//初始化数据库完成回调
	service.OnDatabaseInit()
	//加载数据回调
	service.OnLoadDataBase()
	log.Info("load config...")
	//加载配置
	config.LoadConfig()
	service.OnConfigLoadFinish()
	log.Info("create http engine")
	gin.SetMode(gin.ReleaseMode)
	//注册http引擎
	engine := gin.New()
	engine.Use(cors())
	service.OnRegHttp(engine)
	go func() {
		err := engine.Run(":" + strconv.Itoa(config.Config.Port))
		if err != nil {
			panic(err)
		}
	}()
	//自动保存数据
	closeAutoSaveData := make(chan byte, 0)
	go func() {
		isClose := false
		for {
			sec := 60 * 60
			c := time.NewTimer(gtime.Duration(sec))
			select {
			case <-c.C:
				service.OnSaveDataBase()
			case <-closeAutoSaveData:
				isClose = true
				break
			}
			if isClose {
				break
			}
		}
		closeAutoSaveData <- 1
	}()
	sysClose := make(chan os.Signal, 1)
	go func() {
		for {
			str := ""
			fmt.Scanln(&str)
			switch str {
			case "exit":
				sysClose <- os.Interrupt
			}
		}
	}()
	service.OnGameStart()
	signal.Notify(sysClose, os.Interrupt, os.Kill)
	log.Info("Opt Start Success")
	<-sysClose
	closeAutoSaveData <- 1
	<-closeAutoSaveData
	service.OnGameClose()
	service.OnSaveDataBase()
	log.Info("Opt Close")
	log.Close()
}

// 跨域
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Access-Token,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Set("content-type", "application/json")
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}
