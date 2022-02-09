package database

import (
	"cross/config"
	"database/sql"
	"fmt"
)

var (
	db = &sql.DB{}
)

func GetDB() *sql.DB {
	return db
}

func OnInitDataBase() {
	cnf := Config{}
	cnf.Database = config.Config.Database
	db = New(cnf)
}

type Config struct {
	Database     string
	MaxIdleConns int //最大空闲链接
	MaxOpenConns int //最大打开连接数
}

//初始化配置
func New(c Config) *sql.DB {
	driveSource := fmt.Sprintf("%s?parseTime=True&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci", c.Database)
	var err error
	db, err := sql.Open("mysql", driveSource)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetMaxOpenConns(c.MaxOpenConns)
	if err != nil {
		panic("连接数据库失败；" + err.Error())
	}
	return db
}
