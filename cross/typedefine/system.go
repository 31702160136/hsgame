package typedefine

import (
	"cross/log"
	jsoniter "github.com/json-iterator/go"
)

//该数据为系统数据，系统启动后就会加载完毕
type SystemData struct {
	Common *Common //公共数据
}

/*
	新增数据步骤
	1、新增系统数据对象key
	2、把新增的key注册在SysKeys中
	3、在LoadSysData方法中注册解析
*/

//系统数据对象的key
const (
	KeyCommon = 1
)

var (
	SysKeys = map[int]bool{
		KeyCommon: true,
	}
	systemData = &SystemData{
		Common: &Common{},
	}
)

func LoadSysData(key int, data []byte) {
	switch key {
	case KeyCommon:
		_ = jsoniter.Unmarshal(data, &systemData.Common)
	}
}

func MarshalData(key int) []byte {
	var data = make([]byte, 0)
	var err error
	switch key {
	case KeyCommon:
		data, err = jsoniter.Marshal(systemData.Common)
	}
	if err != nil {
		log.Error(err.Error())
	}
	return data
}

type Common struct {
}

func GetCommonData() *Common {
	return systemData.Common
}
