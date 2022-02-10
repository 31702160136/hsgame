package main

import (
	"fmt"
	"game/base"
	ihttp2 "game/common/ihttp"
	"game/config"
	"game/database"
	"game/dispatch"
	"game/gtime"
	"game/log"
	"game/pack"
	"game/service"
	account2 "game/service/account"
	_ "game/service/cross"
	"game/service/gm"
	_ "game/service/test"
	t "game/typedefine"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
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
	resultByte, err := ihttp2.Post(config.Config.OptAddress+"/getGameServerConfig", args, ihttp2.GetContentTypeJson())
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
	if err != nil {
		log.Fatal(err)
	}
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

	//注册游戏入口
	gameGateEngine := gin.New()
	gameGateEngine.GET("gate", gameGate)
	go func() {
		err := gameGateEngine.Run(":" + strconv.Itoa(config.ServerConfig.Port))
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
		"Game Start Success GamePort:[%d] HTTPPort[%d]",
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

//----------------------------------游戏入口-----------------------------------------------
var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func gameGate(ctx *gin.Context) {
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Fatal(err)
	}

	account := t.NewAccount(conn, ctx.ClientIP())

	defer func() {
		if !account.IsClose() {
			dispatch.PushSystemSyncMsg("logout", service.OnAccountLogout, account, account2.LogoutTag)
		}
	}()

	//发送消息
	go func() {
		for !config.IsGameClose() && !account.IsClose() {
			msgs := account.ReadMsg()
			for _, msg := range msgs {
				account.SyncReply(msg)
			}
		}
	}()

	//读取消息
	for !config.IsGameClose() && !account.IsClose() {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		reader := pack.NewReader(data)
		if reader.Len() < 4 {
			continue
		}

		var sys, cmd int16
		reader.Read(&sys, &cmd)
		dispatch.PushClientMessage(sys, cmd, account, reader)
	}
}
