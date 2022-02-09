package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"opt/gtime"
)

var (
	apisGet          = map[string]func(ctx *gin.Context){}
	apisPost         = map[string]func(ctx *gin.Context){}
	gameStart        = make([]func(), 0)
	gameClose        = make([]func(), 0)
	configLoadFinish = make([]func(), 0)
	databaseInit     = make([]func(), 0)
	saveDataBase     = make([]func(), 0)
	loadDataBase     = make([]func(), 0)
)

func RegGet(path string, fun func(ctx *gin.Context)) {
	if _, ok := apisGet[path]; ok {
		panic(fmt.Sprintf("regGet path:%s exist", path))
	}
	apisGet[path] = fun
}

func RegPost(path string, fun func(ctx *gin.Context)) {
	if _, ok := apisPost[path]; ok {
		panic(fmt.Sprintf("regGet path:%s exist", path))
	}
	apisPost[path] = fun
}

func OnRegHttp(e *gin.Engine) {
	for path, fun := range apisGet {
		e.GET(path, fun)
	}
	for path, fun := range apisPost {
		e.POST(path, fun)
	}
}

func RegGameStart(fun func()) {
	gameStart = append(gameStart, fun)
}

func OnGameStart() {
	for _, fun := range gameStart {
		fun()
	}
}

func RegGameClose(fun func()) {
	gameClose = append(gameClose, fun)
}

func OnGameClose() {
	for _, fun := range gameClose {
		fun()
	}
}

func RegConfigLoadFinish(fun func()) {
	configLoadFinish = append(configLoadFinish, fun)
}
func OnConfigLoadFinish() {
	for _, finish := range configLoadFinish {
		finish()
	}
}

func RegDatabaseInit(fun func()) {
	databaseInit = append(databaseInit, fun)
}

func OnDatabaseInit() {
	for _, fun := range databaseInit {
		fun()
	}
}

func RegSaveDataBase(fun func()) {
	saveDataBase = append(saveDataBase, fun)
}

func OnSaveDataBase() {
	for _, fun := range saveDataBase {
		fun()
	}
	fmt.Println(gtime.Now().Format(gtime.DateTimeFormat)+": ", "save data success")
}

func RegLoadDataBase(fun func()) {
	loadDataBase = append(loadDataBase, fun)
}

func OnLoadDataBase() {
	for _, fun := range loadDataBase {
		fun()
	}
}
