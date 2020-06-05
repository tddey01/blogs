package config

const {
SERVER_INI = "config.ini"
}

// 配置文件内存结构
type TNiConfig struct{
	host string
	port string
	logLevel string

}

var NiConfig TNiConfig