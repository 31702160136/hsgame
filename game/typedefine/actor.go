package typedefine

import "game/pack"

type Actor struct {
	ActorId    int64  //玩家id
	AccountId  string //玩家账号
	Name       string //玩家名称
	Data       *Data  //玩家数据
	LoginTime  int64  //登录时间
	LogoutTime int64  //登出时间
	CreateTime int64  //创建时间
	ServerId   int    //服务id
	Account    *Account
}

type Data struct {
	Icon  int
	Level int
}

func (this *Actor) IsOnline() bool {
	return this.Account != nil
}

func (this *Actor) GetData() *Data {
	if this.Data == nil {
		this.Data = &Data{}
	}
	return this.Data
}

func (this *Actor) ReplyWriter(writer *pack.Writer) {
	if this.Account == nil {
		return
	}
	if this.Account.IsClose() {
		return
	}
	this.Account.WriterMsg(writer.Bytes())
}

func (this *Actor) Reply(sys, cmd int16, data ...interface{}) {
	if this.Account == nil {
		return
	}
	if this.Account.IsClose() {
		return
	}
	writer := pack.NewWriter(sys, cmd)
	writer.Writer(data...)
	this.Account.WriterMsg(writer.Bytes())
}
