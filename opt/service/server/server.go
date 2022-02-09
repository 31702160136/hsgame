package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"opt/base"
	"opt/config"
	"opt/dao/crossserver"
	"opt/dao/gameserver"
	"opt/gtime"
	"opt/service"
	t "opt/typedefine"
	"strings"
	"sync"
)

var (
	gameServers     = map[int]*t.GameServer{}
	gameServersMux  = sync.Mutex{}
	crossServers    = map[int]*t.CrossServer{}
	crossServersMux = sync.Mutex{}
)

func init() {
	service.RegPost("createGameServer", onCreateGameServer)
	service.RegPost("createCrossServer", onCreateCrossServer)
	service.RegPost("getGameServerConfig", onGetGameServerConfig)
	service.RegPost("getCrossServerConfig", onGetCrossServerConfig)

	service.RegLoadDataBase(onLoadDataBase)
}
func onLoadDataBase() {
	servers := gameserver.GetGameServer(context.TODO(), bson.D{}, nil)
	gameServersMux.Lock()
	defer gameServersMux.Unlock()
	for _, server := range servers {
		gameServers[server.ServerId] = server
	}

	crossies := crossserver.GetCrossServer(context.TODO(), bson.D{}, nil)
	crossServersMux.Lock()
	defer crossServersMux.Unlock()
	for _, server := range crossies {
		crossServers[server.ServerId] = server
	}
}

//创建游戏服
func onCreateGameServer(ctx *gin.Context) {
	param := base.NewMap()
	if err := t.BindParam(ctx, &param); err != nil {
		t.Reply(ctx, nil, 1, "数据错误", err.Error())
		return
	}

	if !base.CheckSignature(param, config.Config.SignatureKey) {
		t.Reply(ctx, nil, 2, "签名错误")
		return
	}
	data := t.GameServer{}
	_ = base.Transfer(param, &data)
	if data.Port == 0 ||
		data.CrossServer == 0 ||
		data.MaxOnline == 0 ||
		strings.Trim(data.Name, " ") == "" ||
		strings.Trim(data.IP, " ") == "" ||
		data.ServerId == 0 {
		t.Reply(ctx, nil, 3, "缺少参数")
		return
	}
	if (data.ServerId << 32) <= 0 {
		t.Reply(ctx, nil, 4, "服务id过大或不正确")
		return
	}
	gameServersMux.Lock()
	_, ok := gameServers[data.ServerId]
	gameServersMux.Unlock()
	if ok {
		t.Reply(ctx, nil, 5, "此服已存在")
		return
	}

	data.CreateAt = gtime.Now().Unix()
	data.UpdateAt = gtime.Now().Unix()
	data.Status = 0
	data.Id = ""
	err := gameserver.Create(context.TODO(), &data)
	if err != nil {
		t.Reply(ctx, nil, 6, "系统错误", err.Error())
		return
	}
	gameServersMux.Lock()
	gameServers[data.ServerId] = &data
	gameServersMux.Unlock()
	t.Reply(ctx, nil, 0, "success")

}

//创建跨服
func onCreateCrossServer(ctx *gin.Context) {
	param := base.NewMap()
	if err := t.BindParam(ctx, &param); err != nil {
		t.Reply(ctx, nil, 1, "数据错误", err.Error())
		return
	}

	if !base.CheckSignature(param, config.Config.SignatureKey) {
		t.Reply(ctx, nil, 2, "签名错误")
		return
	}
	data := t.CrossServer{}
	_ = base.Transfer(param, &data)
	if data.Port == 0 ||
		strings.Trim(data.Name, " ") == "" ||
		strings.Trim(data.IP, " ") == "" ||
		data.ServerId == 0 {
		t.Reply(ctx, nil, 3, "缺少参数")
		return
	}
	if (data.ServerId << 32) <= 0 {
		t.Reply(ctx, nil, 4, "服务id过大或不正确")
		return
	}
	crossServersMux.Lock()
	_, ok := crossServers[data.ServerId]
	crossServersMux.Unlock()
	if ok {
		t.Reply(ctx, nil, 5, "此服已存在")
		return
	}

	data.CreateAt = gtime.Now().Unix()
	data.UpdateAt = gtime.Now().Unix()
	data.Id = ""
	err := crossserver.Create(context.TODO(), &data)
	if err != nil {
		t.Reply(ctx, nil, 6, "系统错误", err.Error())
		return
	}
	crossServersMux.Lock()
	crossServers[data.ServerId] = &data
	crossServersMux.Unlock()
	t.Reply(ctx, nil, 0, "success")
}

//获得服务配置
func onGetGameServerConfig(ctx *gin.Context) {
	data := struct {
		ServerId  int    `json:"server_id"`
		Signature string `json:"signature"`
	}{}
	if err := t.BindParam(ctx, &data); err != nil {
		t.Reply(ctx, nil, 1, "数据错误", err.Error())
		return
	}

	if !base.CheckSignature(base.StructToMap(data), config.Config.SignatureKey) {
		t.Reply(ctx, nil, 2, "签名错误")
		return
	}
	serverInfo, ok := gameServers[data.ServerId]
	if !ok {
		t.Reply(ctx, nil, 2, fmt.Sprintf("服务 %d 不存在", data.ServerId))
		return
	}

	crossInfo := &t.CrossServer{}
	if serverInfo.CrossServer > 0 {
		crossInfo, ok = crossServers[serverInfo.CrossServer]
		if !ok {
			t.Reply(ctx, nil, 3, fmt.Sprintf("跨服%d不存在", serverInfo.CrossServer))
			return
		}
	}

	out := base.NewMap()
	out["server_id"] = serverInfo.ServerId
	out["port"] = serverInfo.Port
	out["name"] = serverInfo.Name
	out["http_span_port"] = config.Config.HttpSpanPort
	out["nats"] = crossInfo.Nats
	out["cross_server"] = serverInfo.CrossServer
	out["max_online"] = serverInfo.MaxOnline
	t.Reply(ctx, out, 0, "success")
}

//获得服务配置
func onGetCrossServerConfig(ctx *gin.Context) {
	data := struct {
		ServerId  int    `json:"server_id"`
		Signature string `json:"signature"`
	}{}
	if err := t.BindParam(ctx, &data); err != nil {
		t.Reply(ctx, nil, 1, "数据错误", err.Error())
		return
	}

	if !base.CheckSignature(base.StructToMap(data), config.Config.SignatureKey) {
		t.Reply(ctx, nil, 2, "签名错误")
		return
	}

	crossInfo := &t.CrossServer{}
	var ok bool
	crossInfo, ok = crossServers[data.ServerId]
	if !ok {
		t.Reply(ctx, nil, 3, fmt.Sprintf("跨服%d不存在", data.ServerId))
		return
	}

	out := base.NewMap()
	out["server_id"] = crossInfo.ServerId
	out["port"] = crossInfo.Port
	out["name"] = crossInfo.Name
	out["http_span_port"] = config.Config.HttpSpanPort
	out["nats"] = crossInfo.Nats
	t.Reply(ctx, out, 0, "success")
}
