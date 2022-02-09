package service

import (
	"cross/log"
	"runtime/debug"
)

type gmHandle = func(values map[string]string) (int, interface{})

var (
	gameStart        = make([]func(), 0)
	gameClose        = make([]func(), 0)
	configLoadFinish = make([]func(), 0)

	gmHandles = make(map[string]gmHandle)
)

func RegGameStart(fun func()) {
	gameStart = append(gameStart, fun)
}

func OnGameStart() {
	for _, fun := range gameStart {
		Try(func() {
			fun()
		})
	}
}

func RegGameClose(fun func()) {
	gameClose = append(gameClose, fun)
}

func OnGameClose() {
	for _, fun := range gameClose {
		Try(func() {
			fun()
		})
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

func RegGm(gm string, handle gmHandle) {
	gmHandles[gm] = handle
}

//获取gm方法异步调用
func GetGm(gm string) gmHandle {
	return gmHandles[gm]
}

func Try(fun func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err, string(debug.Stack()))
		}
	}()
	fun()
}
