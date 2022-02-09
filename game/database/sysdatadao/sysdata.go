package sysdatadao

import (
	"game/base"
	"game/database"
	"game/dispatch"
	"game/log"
	t "game/typedefine"
	"sync"
)

const sql = `
CREATE TABLE sysdata (
  id int NOT NULL COMMENT 'key',
  data longblob COMMENT '数据',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`

func init() {
	database.RegLoadData(loadData)
	database.RegInitDatabaseFinishCallBack(initDatabase)
	database.RegSaveData(saveData)
}

func initDatabase() {
	db := database.GetDB()
	//检测表是否存在
	if _, err := db.Exec("select id from sysdata limit 1"); err != nil {
		//创建表
		res, err := db.Exec(sql)
		if res == nil && err == nil || res != nil && err != nil {
			panic(err)
		}
	}
	for key, _ := range t.SysKeys {
		result, err := db.Query("select id from sysdata where id=?", key)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		isExist := false
		for result.Next() {
			isExist = true
		}
		if isExist {
			continue
		}
		_, err = db.Exec("INSERT INTO sysdata (`id`,`data`) VALUES (?,?)", key, base.Zip([]byte{}))
	}
}

func loadData() {
	db := database.GetDB()
	result, err := db.Query("select id,`data` from sysdata")
	if err != nil {
		panic(err)
	}
	for result.Next() {
		var id int
		data := make([]byte, 0)
		if err = result.Scan(&id, &data); err != nil {
			panic(err)
		}
		_, ok := t.SysKeys[id]
		if ok {
			data, err = base.Unzip(data)
			if err != nil {
				panic(err)
			}
			t.LoadSysData(id, data)
		}
	}
}

func saveData() {
	wg := &sync.WaitGroup{}
	for key, _ := range t.SysKeys {
		wg.Add(1)
		var data = make([]byte, 0)
		dispatch.PushSystemSyncMsg("copySysData", func() {
			data = t.MarshalData(key)
			wg.Done()
		})
		wg.Wait()
		go func(key int, data []byte) {
			updateSysData(key, data)
		}(key, data)
	}
}

func updateSysData(key int, data []byte) {
	db := database.GetDB()
	var err error
	data = base.Zip(data)
	_, err = db.Exec("UPDATE `sysdata` SET `data` = ? WHERE `id` = ?;", data, key)
	if err != nil {
		log.Error(err.Error())
	}
}
