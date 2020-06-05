package config

import (
	"github.com/Unknwon/goconfig"
	"github.com/studygolang/sander/logger"
)

const {
SERVER_INI = "config.ini"
}

// 配置文件内存结构
type TNiConfig struct {
	host     string
	port     string
	logLevel string
}

var NiConfig TNiConfig

func reloadConfig( () {
	logger.Info("Rolad config ....")
	conf, err := goconfig.LoadConfigFile("conf/" + SERVER_INI)
	if err == nil {
		if conf != nil {
			host, _ := conf.GetValue("main", "host")
			port, _ := conf.GetValue("main", "port")
			loglevel, _ := conf.GetValue("main", "logLevel")

			NiConfig.host = host
			NiConfig.port = port
			NiConfig.logLevel = loglevel
		}
	}
}
