## 依赖
```
go get github.com/facebookgo/pidfile
go get github.com/garyburd/redigo/redis
go get github.com/satori/go.uuid
go get gopkg.in/ini.v1
```
## 编译安装
```
cd ./src/mcenter
go install
```
## 启动
```
cd ../bin
cp example.config.ini config.ini
cp example.queues.json queues.json
./start.sh #启动
./restart.sh #重启
./stop.sh #停止
```

# config 配置文件
## 系统配置
*详见config.ini*

## 队列配置文件
*样例见 queues.json, 主要是注明消息队列的配置以及回调地址和参数.*

