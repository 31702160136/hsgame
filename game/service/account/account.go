package account

import (
	sql2 "database/sql"
	"game/base"
	ihttp2 "game/common/ihttp"
	"game/config"
	"game/database"
	"game/database/accountdao"
	"game/database/actordao"
	"game/dispatch"
	"game/gtime"
	"game/log"
	"game/pack"
	"game/service"
	t "game/typedefine"
	jsoniter "github.com/json-iterator/go"
	"proto"
)

const (
	//账号登出
	LogoutTag        byte = 1 //正常退出
	LogoutTagReplace byte = 2 //被顶替
)

var (
	maxActorId int64
)

func init() {
	database.RegInitDatabaseFinishCallBack(LoadMaxActorId)

	dispatch.RegClientAccountMsg(proto.SystemCLogin, onAccountLogin)
	dispatch.RegClientAccountMsg(proto.SystemCGetActorList, onGetActorList)
	dispatch.RegClientAccountMsg(proto.SystemCCreateActor, onCreateActor)
	dispatch.RegClientAccountMsg(proto.SystemCEnterGame, onEnterGame)

	service.OnAccountLogout = onLogout
}

//账号登录
func onAccountLogin(account *t.Account, reader *pack.Reader) {
	var (
		accountId string
		password  string
	)
	reader.Read(&accountId, &password)

	dispatch.PushSystemAsyncMsg("login", func(account *t.Account, serverId int, accountId string, code int) {
		if code != 0 && code != 5 {
			return
		}
		writer := pack.NewPack(proto.System, proto.SystemSLogin)
		status := byte(0)
		if code == 5 {
			status = 1
			writer.Writer(status)
			account.ReplyWriter(writer)
			return
		}

		//顶下线
		oldAccount := accountdao.GetAccount(accountId)
		if oldAccount != nil {
			service.OnAccountLogout(oldAccount, LogoutTagReplace)
		}

		account.AccountId = accountId
		accountdao.AddAccount(account)
		writer.Writer(status)
		account.ReplyWriter(writer)
	}, func(account *t.Account, serverId int, accountId, password string) (*t.Account, int, string, int) {
		args := base.NewMap()
		args["server_id"] = serverId
		args["account"] = accountId
		args["password"] = password
		args["ip"] = account.IP
		args["signature"] = base.Signature(args, config.Config.SignatureKey)
		resultByte, err := ihttp2.Post(config.Config.OptAddress+"/accountLogin", args, ihttp2.GetContentTypeJson())
		if err != nil {
			return nil, 0, "", -1
		}
		result := t.HttpResult{}
		err = jsoniter.Unmarshal(resultByte, &result)
		if err != nil {
			return nil, 0, "", -1
		}

		return account, serverId, accountId, result.Code
	}, account, config.ServerId, accountId, password)
}

//账号登出
func accountLogout(account *t.Account, tag byte) {
	defer account.Close()
	writer := pack.NewPack(proto.System, proto.SystemSLogout, tag)
	account.SyncReply(writer.Bytes())
}

func onLogout(account *t.Account, tag byte) {
	defer func() {
		accountdao.DelAccount(account.AccountId)
		if account.Actor == nil {
			return
		}
		actordao.UnOnlineActor(account.Actor.ActorId)
		account.Actor = nil
	}()
	if account.Actor != nil {
		account.Actor.Account = nil
		account.Actor.LogoutTime = gtime.Now().Unix()
		log.Infof("actor(%d) logout", account.Actor.ActorId)
	}
	//异步
	dispatch.PushSystemAsyncMsg("logoutTag", nil, func(account2 *t.Account) {
		accountLogout(account2, tag)
	}, account)
}

//查询角色列表
func onGetActorList(account *t.Account, reader *pack.Reader) {
	if account.AccountId == "" || account.IsClose() {
		return
	}
	dispatch.PushSystemAsyncMsg("getActorList", func(actors []*t.Actor, err error) {
		if err != nil {
			log.Error(err.Error())
			return
		}
		writer := pack.NewPack(proto.System, proto.SystemSGetActorList, int16(len(actors)))
		for _, actor := range actors {
			baseData := actor.GetData()
			writer.Writer(actor.ActorId)
			writer.Writer(actor.Name)
			writer.Writer(baseData.Icon)
			writer.Writer(baseData.Level)
		}
		account.ReplyWriter(writer)
	}, actordao.GetActors, account.AccountId)
}

//创建角色
func onCreateActor(account *t.Account, reader *pack.Reader) {
	if account.AccountId == "" || account.IsClose() {
		return
	}
	var (
		name string
	)
	reader.Read(&name)
	actorNames := actordao.GetAllActorName()
	for _, v := range actorNames {
		if name == v {
			return
		}
	}
	actor := &t.Actor{
		ActorId:    newActorId(),
		Name:       name,
		AccountId:  account.AccountId,
		Data:       &t.Data{},
		CreateTime: gtime.Now().Unix(),
	}
	dispatch.PushSystemAsyncMsg("createActor", func(status bool) {
		if !status {
			return
		}
		actordao.AddActorName(actor.ActorId, name)
		account.Reply(proto.System, proto.SystemSCreateActor, actor.ActorId, actor.Name)
	}, actordao.InsertActor, actor)
}

//进入游戏
func onEnterGame(account *t.Account, reader *pack.Reader) {
	if account.AccountId == "" || account.IsClose() {
		return
	}
	var (
		actorId int64
	)
	reader.Read(&actorId)
	dispatch.PushSystemAsyncMsg("enterGame", func(actor *t.Actor) {
		if actor == nil || actor.AccountId != account.AccountId {
			return
		}
		account.Actor = actor
		actor.Account = account
		actor.LoginTime = gtime.Now().Unix()
		actordao.AddOnlineActor(actor)
		writer := pack.NewPack(proto.System, proto.SystemSEnterGame)
		baseData := actor.GetData()
		writer.Writer(
			config.ServerId,
			config.ServerConfig.Name,
			gtime.Now().Unix(),
			actor.Account.IP,
			actor.ActorId,
			actor.Name,
			baseData.Icon,
			baseData.Level,
		)
		actor.ReplyWriter(writer)
	}, actordao.GetActor, actorId)
}

//加载最大玩家id
func LoadMaxActorId() {
	var err error
	maxActorId, err = actordao.GetMaxActorId()
	if err != nil {
		if err != sql2.ErrNoRows {
			panic(err)
		}
	}
	serverId := int64(config.ServerId)
	if maxActorId == 0 {
		maxActorId = serverId << 32
	}
}

func newActorId() int64 {
	maxActorId++
	return maxActorId
}
