## 依赖
```
go get github.com/facebookgo/pidfile
go get github.com/streadway/amqp
go get github.com/satori/go.uuid
go get gopkg.in/ini.v1
git clone https://github.com/golang/text.git $GOPATH/src/golang.org/x/text
git clone https://github.com/golang/net.git $GOPATH/src/golang.org/x/net
git clone https://github.com/golang/sys.git $GOPATH/src/golang.org/x/sys
git clone https://github.com/grpc/grpc-go.git $GOPATH/src/google.golang.org/grpc
git clone https://github.com/google/go-genproto.git $GOPATH/src/google.golang.org/genproto
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
./install.sh #安装
./start.sh #启动
./restart.sh #重启
./stop.sh #停止
```

# config 配置文件
## 系统配置
*详见config.ini*

## 队列配置文件
*样例见 queues.json, 主要是注明消息队列的配置以及回调地址和参数.*

