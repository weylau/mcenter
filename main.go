package main

import (
	"flag"
	"mcenter/app"
)
//加载配置文件
var configFile = flag.String("conf", "./etc/config.toml", "Server configuration file path")

func main() {
	flag.Parse()
	application := app.Default(*configFile)
	application.Run()
}
