package config

import (
	"gopkg.in/ini.v1"
)

var (
	ConfFile   = "config.ini"
	QueunsFile = "queues.json"
	LogDir     = "./logs/"
	AesToken   = "123456"
)

type SysConfig struct {
	AmqpHost                string `ini:"amqp_host"`
	ReceiveNum              int    `ini:"receive_num"`
	WorkNum                 int    `ini:"work_num"`
	AckNum                  int    `ini:"ack_num"`
	ResendNum               int    `ini:"resend_num"`
	HttpMaxIdleConns        int    `ini:"http_max_idle_conns"`
	HttpMaxIdleConnsPerHost int    `ini:"http_max_idle_conns_per_host"`
	HttpIdleConnTimeout     int    `ini:"http_idle_conn_timeout"`
	LogDir                  string `ini:"log_dir"`
	AesToken                string `ini:"aes_token"`
}

//加载系统配置文件
func GetConfig(configFileName string) (*SysConfig, error) {
	config := &SysConfig{}
	conf, err := ini.Load(configFileName) //加载配置文件
	if err != nil {
		return config, err
	}
	conf.BlockMode = false
	err = conf.MapTo(&config) //解析成结构体
	if err != nil {
		return config, err
	}
	return config, nil
}
