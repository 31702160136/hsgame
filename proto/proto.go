package proto

/*
	C为客户端调用协议
	S为服务端结果返回
*/

//系统内部跨服协议
const (
	CrossActorMsg      = 1 //玩家消息
	CrossReplyActorMsg = 2 //回复玩家消息
	CrossTest          = 3 //测试
)

//基础
const (
	System int16 = 1
	Test   int16 = 2 //测试
)

//System
const (
	SystemCLogin        int16 = 1 //账号登录
	SystemSLogin        int16 = 1 //账号登录结果
	SystemSLogout       int16 = 2 //下线通知
	SystemCGetActorList int16 = 3 //查询角色列表
	SystemSGetActorList int16 = 3 //查询角色列表
	SystemCCreateActor  int16 = 4 //创建角色
	SystemSCreateActor  int16 = 4 //创建角色
	SystemCEnterGame    int16 = 5 //进入游戏
	SystemSEnterGame    int16 = 5 //进入游戏
)

const (
	TestCTest       int16 = 1
	TestSTest       int16 = 1
	TestCCrossTest  int16 = 2
	TestSCrossTest  int16 = 2
	TestCCrossTest2 int16 = 3
	TestSCrossTest2 int16 = 3
)
