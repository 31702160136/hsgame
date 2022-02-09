package main

import (
	"fmt"
	"opt/base"
)

var key = "fb45c06ff5aefa76c4aa254261ec085e"

func main() {
	data := base.NewMap()
	data["name"] = "测试服"
	data["server_id"] = 1
	data["ip"] = "127.0.0.1"
	data["port"] = 3000
	data["nats"] = "nat://127.0.0.1:4222"
	data["signature"] = base.Signature(data, key)
	result, err := Post("http://127.0.0.1:2000/createCrossServer", data, GetContentTypeJson())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))

}
