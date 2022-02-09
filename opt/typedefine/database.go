package typedefine

//游戏服
type GameServer struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name"`                 //名称
	IP          string `json:"ip" bson:"ip"`                     //ip
	Port        int    `json:"port" bson:"port"`                 //端口
	ServerId    int    `json:"server_id" bson:"server_id"`       //服务id
	CrossServer int    `json:"cross_server" bson:"cross_server"` //跨服id
	Status      byte   `json:"status" bson:"status"`             //服务状态
	MaxOnline   int    `json:"max_online" bson:"max_online"`     //最大在线人数
	CreateAt    int64  `json:"create_at" bson:"create_at"`
	UpdateAt    int64  `json:"update_at" bson:"update_at"`
}

//跨服
type CrossServer struct {
	Id       string `json:"id" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"name"`           //名称
	ServerId int    `json:"server_id" bson:"server_id"` //服务id
	IP       string `json:"ip" bson:"ip"`               //ip
	Port     int    `json:"port" bson:"port"`           //端口
	Nats     string `json:"nats" bson:"nats"`           //nats中间件链接地址
	CreateAt int64  `json:"create_at" bson:"create_at"`
	UpdateAt int64  `json:"update_at" bson:"update_at"`
}

//玩家账号
type Account struct {
	Id                string        `json:"id" bson:"_id,omitempty"`
	Account           string        `json:"account" bson:"account"`                         //账号
	Password          string        `json:"password" bson:"password"`                       //密码
	ServerIds         map[int]byte  `json:"server_ids" bson:"server_ids"`                   //服务id
	RegisterTime      int64         `json:"register_time" bson:"register_time"`             //注册时间
	LoginServerRecord map[int]int64 `json:"login_server_record" bson:"login_server_record"` //登录记录
	LastLoginRecord   int64         `json:"last_login_record" bson:"last_login_record"`     //最后一次登录记录
	IP                string        `json:"ip" bson:"ip"`                                   //最近登录的ip
}
