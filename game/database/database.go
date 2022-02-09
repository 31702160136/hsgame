package database

import (
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

var (
	initDatabaseFinishCallBack = make([]func(), 0)
	loadData                   = make([]func(), 0)
	saveData                   = make([]func(), 0)
)

//数据库初始化完成回调
func RegInitDatabaseFinishCallBack(fun func()) {
	initDatabaseFinishCallBack = append(initDatabaseFinishCallBack, fun)
}

//数据库初始化完成回调
func OnInitDatabaseFinishCallBack() {
	for _, fun := range initDatabaseFinishCallBack {
		fun()
	}
}

//保存数据
func RegSaveData(fun func()) {
	saveData = append(saveData, fun)
}

//保存数据
func OnSaveData() {
	wg := sync.WaitGroup{}
	wg.Add(len(saveData))
	for _, fun := range saveData {
		go func(fun func()) {
			fun()
			wg.Done()
		}(fun)
	}
	wg.Wait()
}

//加载数据
func RegLoadData(fun func()) {
	loadData = append(loadData, fun)
}

//加载数据
func OnLoadData() {
	wg := sync.WaitGroup{}
	wg.Add(len(loadData))
	for _, fun := range loadData {
		go func(fun func()) {
			fun()
			wg.Done()
		}(fun)
	}
	wg.Wait()
}
