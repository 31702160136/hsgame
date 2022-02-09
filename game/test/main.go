package main

import (
	"fmt"
	"runtime/debug"
)

func main() {
	//values := base.NewMap()
	////values["actor"] = "1735166787799"
	////values["inviteAccountId"] = "1848658242"
	////values["inviteActorId"] = "1735166787588"
	////values["inviteServerId"] = "404"
	////values["beInviteAccountId"] = "2195779861"
	////values["beInviteActorId"] = "1735166787799"
	//pKeys := make([]string, 0)
	//for key, _ := range values {
	//	pKeys = append(pKeys, key)
	//}
	//
	//sort.Slice(pKeys, func(i, j int) bool {
	//	return pKeys[i] < pKeys[j]
	//})
	//content := make([]string, len(pKeys))
	//for i, key := range pKeys {
	//	content[i] = fmt.Sprintf("%s=%v", key, values[key])
	//}
	//fmt.Println(content)
	//key := "cmuokg05128anADspc295982@05APMby"
	//fmt.Println(base.MD5(fmt.Sprintf("%s%s", key, strings.Join(content, ""))))
	//bt, _ := json.Marshal(values)
	//fmt.Println(string(bt))
	Try(func() {
		panic("sfaslkjfsalkfj")
	})
	fmt.Println(1113)
}

func Try(fun func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err,string(debug.Stack()))
		}
	}()
	fun()
}
