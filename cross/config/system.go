package config

import t "cross/typedefine"

var (
	ServerId     int
	Config       = t.Config{}
	ServerConfig = t.ServerConfig{}
	gameStatus   bool
)

func IsGameClose() bool {
	return gameStatus
}
