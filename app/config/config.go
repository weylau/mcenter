package config

import (
	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
	"io/ioutil"
	"mcenter/app/queue"
)

var Conf *SysConfig

type SysConfig struct {
	Debug                   bool        `toml:"debug"`
	LogLevel                string      `toml:"log_level"`
	AmqpHost                string      `toml:"amqp_host"`
	ReceiveNum              int         `toml:"receive_num"`
	WorkNum                 int         `toml:"work_num"`
	AckNum                  int         `toml:"ack_num"`
	ResendNum               int         `toml:"resend_num"`
	HttpMaxIdleConns        int         `toml:"http_max_idle_conns"`
	HttpMaxIdleConnsPerHost int         `toml:"http_max_idle_conns_per_host"`
	HttpIdleConnTimeout     int         `toml:"http_idle_conn_timeout"`
	LogDir                  string      `toml:"log_dir"`
	AesToken                string      `toml:"aes_token"`
	Projects                []*queue.Projects `toml:"projects"`
}

//加载系统配置文件
func Default(configFile string) ( error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return errors.Trace(err)
	}
	_, err = toml.Decode(string(data), &Conf)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
