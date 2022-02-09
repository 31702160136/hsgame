package account

import (
	"context"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"opt/base"
	"opt/config"
	account2 "opt/dao/account"
	"opt/gtime"
	"opt/log"
	"opt/service"
	t "opt/typedefine"
	"strings"
	"sync"
)

var (
	accounts      = make(map[string]*t.Account)
	accountsMux   = sync.RWMutex{}
	accountIds    = map[string]byte{}
	accountIdsMux = sync.RWMutex{}
)

func init() {
	service.RegPost("regAccount", onRegAccount)
	service.RegPost("accountLogin", onAccountLogin)
	service.RegSaveDataBase(saveData)
	service.RegLoadDataBase(loadData)
	service.RegGameClose(saveData)
}
func loadData() {
	accountBuff := account2.GetAllAccounts(context.TODO())
	accountIdsMux.Lock()
	for _, accStr := range accountBuff {
		accountIds[accStr] = 1
	}
	accountIdsMux.Unlock()
}
func saveData() {
	buff := map[string]*t.Account{}
	accountsMux.Lock()
	bt, _ := jsoniter.Marshal(accounts)
	_ = jsoniter.Unmarshal(bt, &buff)
	accounts = map[string]*t.Account{}
	accountsMux.Unlock()
	for account, info := range buff {
		err := account2.UpdateAccountInfo(context.TODO(), info)
		if err != nil {
			log.Error("save account error", err.Error())
			accountsMux.Lock()
			if _, ok := accounts[account]; !ok {
				accounts[account] = info
			}
			accountsMux.Unlock()
		}
	}
}

//注册账号
func onRegAccount(ctx *gin.Context) {
	var data struct {
		Account   string `json:"account"`
		Password  string `json:"password"`
		Signature string `json:"signature"`
	}
	if err := t.BindParam(ctx, &data); err != nil {
		t.Reply(ctx, nil, 1, "数据错误", err.Error())
		return
	}

	if !base.CheckSignature(base.StructToMap(data), config.Config.SignatureKey) {
		t.Reply(ctx, nil, 2, "签名错误")
		return
	}

	if strings.Trim(data.Account, " ") == "" ||
		strings.Trim(data.Password, " ") == "" {
		t.Reply(ctx, nil, 2, "缺少参数")
		return
	}
	accountIdsMux.Lock()
	_, ok := accountIds[data.Account]
	accountIdsMux.Unlock()
	if ok {
		t.Reply(ctx, nil, 3, "账号已存在")
		return
	}

	account := &t.Account{}
	account.Account = data.Account
	account.Password = base.MD5(data.Password)
	account.RegisterTime = gtime.Now().Unix()
	err := account2.Create(context.TODO(), account)
	if err != nil {
		t.Reply(ctx, nil, 4, "注册失败", err.Error())
		return
	}
	accountsMux.Lock()
	accounts[account.Account] = account
	accountsMux.Unlock()
	accountIdsMux.Lock()
	accountIds[data.Account] = 1
	accountIdsMux.Unlock()
	t.Reply(ctx, nil, 0, "success")
}

//玩家登录
func onAccountLogin(ctx *gin.Context) {
	data := struct {
		Account   string `json:"account"`
		Password  string `json:"password"`
		ServerId  int    `json:"server_id"`
		IP        string `json:"ip"`
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

	accountsMux.RLock()
	accountInfo, ok := accounts[data.Account]
	accountsMux.RUnlock()
	if !ok {
		accountInfo := account2.GetAccountInfoByAccount(context.TODO(), data.Account)
		if accountInfo == nil {
			t.Reply(ctx, nil, 3, "玩家不存在")
			return
		}
		if accountInfo.Password != base.MD5(data.Password) {
			t.Reply(ctx, nil, 5, "密码错误")
			return
		}
		accountsMux.Lock()
		accountInfo.LastLoginRecord = gtime.Now().Unix()
		if accountInfo.LoginServerRecord == nil {
			accountInfo.LoginServerRecord = map[int]int64{}
		}
		accountInfo.LoginServerRecord[data.ServerId] = gtime.Now().Unix()
		accountInfo.IP = data.IP
		if accountInfo.ServerIds == nil {
			accountInfo.ServerIds = map[int]byte{}
		}
		accountInfo.ServerIds[data.ServerId] = 1
		accounts[accountInfo.Account] = accountInfo
		accountsMux.Unlock()
	} else {
		if accountInfo.Password != base.MD5(data.Password) {
			t.Reply(ctx, nil, 5, "密码错误")
			return
		}
	}
	out := base.NewMap()
	out["data"] = accountInfo
	t.Reply(ctx, out, 0, "success")
}
