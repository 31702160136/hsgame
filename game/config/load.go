package config

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
)

func loadConfig(name string, cnf interface{}) {
	bt, err := ioutil.ReadFile(name + "." + "json")
	if err != nil {
		panic(name + "," + err.Error())
	}
	err = jsoniter.Unmarshal(bt, &cnf)
	if err != nil {
		panic(name + "," + err.Error())
	}
}
