package actordao

import (
	sql2 "database/sql"
	"fmt"
	"game/base"
	"game/config"
	"game/database"
	"game/dispatch"
	"game/log"
	t "game/typedefine"
	jsoniter "github.com/json-iterator/go"
	"strings"
	"sync"
	"time"
)

type actorBuffer struct {
	ActorId    int64  //玩家id
	AccountId  string //玩家账号
	Name       string //玩家名称
	Data       []byte //玩家数据
	LoginTime  int64  //登录时间
	LogoutTime int64  //登出时间
	CreateTime int64  //创建时间
}

const sql = `
	CREATE TABLE actor (
	  actorId bigint NOT NULL,
	  accountId varchar(64) DEFAULT NULL,
	  name varchar(64) DEFAULT NULL,
	  data longblob,
	  loginTime int DEFAULT NULL,
	  logoutTime int DEFAULT NULL,
	  createTime int DEFAULT NULL,
	  serverId int DEFAULT NULL,
	  PRIMARY KEY (actorId),
	  KEY accountId (accountId) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`

var (
	actorOnline    = make(map[int64]*t.Actor)
	actorOnlineMux = sync.Mutex{}
	actorName      = make(map[int64]string)
	actorCache     = make(map[int64]*t.Actor)
	actorCacheMux  = sync.Mutex{}
)

func init() {
	database.RegLoadData(loadData)
	database.RegInitDatabaseFinishCallBack(initDatabase)
	database.RegSaveData(saveData)
}

func initDatabase() {
	db := database.GetDB()
	//检测表是否存在
	if _, err := db.Exec("select actorId from actor limit 1"); err != nil {
		//创建表
		res, err := db.Exec(sql)
		if res == nil && err == nil || res != nil && err != nil {
			panic(err)
		}
	}
}

func saveData() {
	buffs := make(map[int64]*actorBuffer)
	wg := &sync.WaitGroup{}
	actors := map[int64]byte{}
	wg.Add(1)
	//到逻辑线程获取所有玩家id
	dispatch.PushSystemSyncMsg("copyActorIds", func() {
		names := GetAllActorName()
		for actorId, _ := range names {
			actors[actorId] = 1
		}
		wg.Done()
	})
	wg.Wait()

	//到逻辑线程获取玩家数据
	for actorId, _ := range actors {
		wg.Add(1)
		dispatch.PushSystemSyncMsg("saveActor", func(actorId int64) {
			actor := getActorCache(actorId)
			if actor != nil {
				data, _ := jsoniter.Marshal(actor.Data)
				buffs[actorId] = &actorBuffer{
					ActorId:    actorId,
					AccountId:  actor.AccountId,
					Name:       actor.Name,
					Data:       data,
					LoginTime:  actor.LoginTime,
					LogoutTime: actor.LogoutTime,
					CreateTime: actor.CreateTime,
				}
				if !actor.IsOnline() {
					delActorCache(actorId)
				}
			}
			wg.Done()
		}, actorId)
		wg.Wait()
		time.Sleep(time.Millisecond * 10)
	}
	wg.Add(len(buffs))
	//保存玩家数据
	for _, buff := range buffs {
		go func(buf *actorBuffer) {
			saveActor(buf)
			wg.Done()
		}(buff)
	}
	wg.Wait()
}

func loadData() {
	loadAllActorName()
}

//获取玩家信息
func GetActor(actorId int64) *t.Actor {
	actorOnlineMux.Lock()
	actor, ok := actorOnline[actorId]
	actorOnlineMux.Unlock()
	if ok {
		return actor
	}

	actor = getActorCache(actorId)
	if actor != nil {
		return actor
	}

	db := database.GetDB()
	result := db.QueryRow("select accountId, `name`, `data`, loginTime, logoutTime, createTime, serverId from actor where actorId=? limit 1", actorId)
	var (
		accountId  = ""
		name       = ""
		data       = make([]byte, 0)
		loginTime  int64
		logoutTime int64
		createTime int64
		serverId   int
	)

	err := result.Scan(&accountId, &name, &data, &loginTime, &logoutTime, &createTime, &serverId)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	actor = &t.Actor{
		ActorId:    actorId,
		AccountId:  accountId,
		Name:       name,
		LoginTime:  loginTime,
		LogoutTime: logoutTime,
		CreateTime: createTime,
		ServerId:   serverId,
	}
	dataStr, _ := base.Unzip(data)
	_ = jsoniter.Unmarshal([]byte(dataStr), &actor.Data)
	saveActorCache(actor)
	return actor
}

func AddOnlineActor(actor *t.Actor) {
	actorOnlineMux.Lock()
	actorOnline[actor.ActorId] = actor
	actorOnlineMux.Unlock()
}

func GetOnlineActor(actorId int64) *t.Actor {
	actorOnlineMux.Lock()
	defer actorOnlineMux.Unlock()
	return actorOnline[actorId]
}

func DelOnlineActor(actorId int64) {
	actorOnlineMux.Lock()
	defer actorOnlineMux.Unlock()
	delete(actorOnline, actorId)
}

//批量插入玩家
func InsertActors(actors map[int64]*t.Actor) bool {
	db := database.GetDB()
	ages := make([]string, 0)
	for _, actor := range actors {
		value := make([]string, 0)
		value = append(value, fmt.Sprintf("%d", actor.ActorId))
		value = append(value, fmt.Sprintf("'%s'", actor.AccountId))
		value = append(value, fmt.Sprintf("'%s'", actor.Name))
		data, _ := jsoniter.Marshal(actor.Data)
		value = append(value, fmt.Sprintf("'%s'", string(base.Zip(data))))
		value = append(value, fmt.Sprintf("%d", actor.LoginTime))
		value = append(value, fmt.Sprintf("%d", actor.LogoutTime))
		value = append(value, fmt.Sprintf("%d", actor.CreateTime))
		str := "(" + strings.Join(value, ",") + ")"
		ages = append(ages, str)
	}
	fmt.Println(strings.Join(ages, ","))
	_, err := db.Exec("INSERT INTO actor (`actorId`, `accountId`, `name`, `data`, `loginTime`, `logoutTime`, `createTime`) VALUES " + strings.Join(ages, ","))
	if err != nil {
		log.Error(err.Error())
		return false
	}
	return true
}

//插入玩家
func InsertActor(actor *t.Actor) bool {
	actors := map[int64]*t.Actor{}
	actors[actor.ActorId] = actor
	return InsertActors(actors)
}

//更新玩家信息
func saveActor(actor *actorBuffer) bool {
	db := database.GetDB()
	args := []interface{}{actor.AccountId, actor.Name, actor.Data, actor.LoginTime, actor.LogoutTime, actor.CreateTime, actor.ActorId}
	_, err := db.Exec("UPDATE `actor` SET `accountId` = ?, `name` = ?, `data` = ?, `loginTime` = ?, `logoutTime` = ?, `createTime` = ? WHERE `actorId` = ?;", args...)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	log.Infof("save actor:%d", actor.ActorId)
	return true
}

//查询全服玩家名称
func loadAllActorName() {
	db := database.GetDB()
	result, err := db.Query("select actorId,`name` from actor")
	if err != nil {
		panic(err)
	}
	for result.Next() {
		actorId := int64(0)
		name := ""
		if err = result.Scan(&actorId, &name); err != nil {
			panic(err)
		}
		actorName[actorId] = name
	}
}

//获取所有用户名字
func GetAllActorName() map[int64]string {
	return actorName
}

//增加玩家名称
func AddActorName(actorId int64, name string) {
	actorName[actorId] = name
}

//获取玩家名称
func GetActorName(actorId int64) string {
	return actorName[actorId]
}

//是否在线
func IsOnline(actorId int64) bool {
	actorOnlineMux.Lock()
	_, ok := actorOnline[actorId]
	actorOnlineMux.Unlock()
	return ok
}

func GetActors(accountId string) ([]*t.Actor, error) {
	db := database.GetDB()
	rows, err := db.Query("select `actorId`, `name`, `data`, loginTime, logoutTime, createTime from actor where accountId=?", accountId)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	actors := make([]*t.Actor, 0)
	for rows.Next() {
		actorId := int64(0)
		name := ""
		data := make([]byte, 0)
		loginTime := int64(0)
		logoutTime := int64(0)
		createTime := int64(0)
		err := rows.Scan(&actorId, &name, &data, &loginTime, &logoutTime, &createTime)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		actor := &t.Actor{
			ActorId:    actorId,
			AccountId:  accountId,
			Name:       name,
			LoginTime:  loginTime,
			LogoutTime: logoutTime,
			CreateTime: createTime,
		}
		dataStr, _ := base.Unzip(data)
		_ = jsoniter.Unmarshal([]byte(dataStr), &actor.Data)
		actors = append(actors, actor)
	}
	return actors, nil
}

//获取最大玩家id
func GetMaxActorId() (int64, error) {
	var maxId sql2.NullInt64
	db := database.GetDB()
	err := db.QueryRow("select max(actorId) from actor where (actorid>>32)=?", config.ServerId).Scan(&maxId)
	if err != nil {
		return 0, err
	}
	return maxId.Int64, nil
}

//获取玩家缓存
func getActorCache(actorId int64) *t.Actor {
	actorCacheMux.Lock()
	actor := actorCache[actorId]
	actorCacheMux.Unlock()
	return actor
}

//保存玩家缓存
func saveActorCache(actor *t.Actor) {
	actorCacheMux.Lock()
	actorCache[actor.ActorId] = actor
	actorCacheMux.Unlock()
}

//删除玩家缓存
func delActorCache(actorId int64) {
	actorCacheMux.Lock()
	delete(actorCache, actorId)
	actorCacheMux.Unlock()
}

//下线玩家
func UnOnlineActor(actorId int64) {
	DelOnlineActor(actorId)
}
