package main

import (
	"cross/base"
	"cross/common/ihttp"
	"cross/config"
	"cross/database"
	_ "cross/database/sysdatadao"
	"cross/dispatch"
	"cross/gtime"
	"cross/log"
	"cross/service"
	"cross/service/gm"
	t "cross/typedefine"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	//加载基础配置文件
	configName := "config.json"
	bt, err := ioutil.ReadFile(configName)
	if err != nil {
		panic(configName + "," + err.Error())
	}

	//加载配置信息
	cnfs := make(map[int]map[string]interface{})
	err = jsoniter.Unmarshal(bt, &cnfs)
	if err != nil {
		panic(configName + "," + err.Error())
	}
	serverId := 0
	cnf := base.NewMap()
	if len(os.Args) > 1 {
		serverId, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
		if _, ok := cnfs[serverId]; !ok {
			panic(fmt.Sprintf("服务 %d 配置不存在", serverId))
		}
		cnf = cnfs[serverId]
	} else {
		if len(cnfs) == 1 {
			for cnfServerId, v := range cnfs {
				serverId = cnfServerId
				cnf = v
			}
		} else {
			panic("no serverId")
		}
	}

	config.ServerId = serverId
	bt, err = jsoniter.Marshal(cnf)
	if err != nil {
		panic(err)
	}
	err = jsoniter.Unmarshal(bt, &config.Config)
	if err != nil {
		panic(err)
	}

	//初始化日志
	log.InitLog(config.Config.Log)
	defer log.Close()

	//注册服务
	args := base.NewMap()
	args["server_id"] = config.ServerId
	args["signature"] = base.Signature(args, config.Config.SignatureKey)
	resultByte, err := ihttp.Post(config.Config.OptAddress+"/getCrossServerConfig", args, ihttp.GetContentTypeJson())
	if err != nil {
		log.Fatal("中心服系统错误", err)
	}
	result := t.HttpResult{}
	err = jsoniter.Unmarshal(resultByte, &result)
	if err != nil {
		log.Fatal(err)
	}

	if result.Code > 0 {
		log.Fatalf("code=%d msg=%s", result.Code, result.Msg)
	}

	bt, err = jsoniter.Marshal(result.Data)
	err = jsoniter.Unmarshal(bt, &config.ServerConfig)

	log.Info("load database...")
	//初始化数据库
	database.OnInitDataBase()
	//初始化数据库完成回调
	database.OnInitDatabaseFinishCallBack()
	//加载数据
	database.OnLoadData()
	log.Info("load config...")
	//加载配置
	config.LoadConfig()
	service.OnConfigLoadFinish()
	log.Info("create http engine")
	gin.SetMode(gin.ReleaseMode)
	//注册http引擎
	httpEngine := gin.New()
	httpEngine.POST("/gm/:name", gm.GmHandle)
	go func() {
		err := httpEngine.Run(":" + strconv.Itoa(config.ServerConfig.Port+config.ServerConfig.HttpSpanPort))
		if err != nil {
			log.Fatal(err)
		}
	}()

	//自动保存数据
	closeAutoSaveData := make(chan byte, 0)
	go func() {
		isClose := false
		for {
			sec := 60 * 30
			c := time.NewTimer(gtime.Duration(sec))
			select {
			case <-c.C:
				database.OnSaveData()
				log.Info("saveData")
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
	dispatch.OnRunGame()
	service.OnGameStart()
	signal.Notify(sysClose, os.Interrupt, os.Kill)
	log.Infof(
		"Cross Start Success Cross-Port:[%d] HTTP-Port[%d]",
		config.ServerConfig.Port,
		config.ServerConfig.Port+config.ServerConfig.HttpSpanPort,
	)
	<-sysClose
	closeAutoSaveData <- 1
	<-closeAutoSaveData
	service.OnGameClose()
	database.OnSaveData()
	log.Info("game close")
}
