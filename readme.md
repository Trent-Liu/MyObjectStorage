#### 系统与go版本

Ubuntu18.04LTS

go version go1.18.4 linux/amd64







#### 服务器网卡注册多个虚拟地址

```shell
sudo ifconfig eth0:1 10.29.1.1/16
sudo ifconfig eth0:2 10.29.1.2/16
sudo ifconfig eth0:3 10.29.1.3/16
sudo ifconfig eth0:4 10.29.1.4/16
sudo ifconfig eth0:5 10.29.1.5/16
sudo ifconfig eth0:6 10.29.1.6/16

sudo ifconfig eth0:7 10.29.2.1/16
sudo ifconfig eth0:8 10.29.2.2/16

sudo ifconfig eth0:9 10.29.1.7/16
sudo ifconfig eth0:10 10.29.1.8/16
```



创建相应的$STORAGE_ROOT目录以及子目录objects

```shell
for i in `seq 1 6`
do
    mkdir -p /tmp/$i/objects
    mkdir -p /tmp/$i/temp
    mkdir -p /tmp/$i/garbage
done
```



#### 启动

同时启动8个dataServer

```shell
export RABBITMQ_SERVER=amqp://test:test@localhost:5672
LISTEN_ADDRESS=10.29.1.1:12345 STORAGE_ROOT=/tmp/1 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.2:12345 STORAGE_ROOT=/tmp/2 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.3:12345 STORAGE_ROOT=/tmp/3 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.4:12345 STORAGE_ROOT=/tmp/4 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.5:12345 STORAGE_ROOT=/tmp/5 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.6:12345 STORAGE_ROOT=/tmp/6 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.7:12345 STORAGE_ROOT=/tmp/7 go run ./dataServer/dataServer.go &
LISTEN_ADDRESS=10.29.1.8:12345 STORAGE_ROOT=/tmp/8 go run ./dataServer/dataServer.go &

```

启动2个apiServer

```shell
export RABBITMQ_SERVER=amqp://test:test@localhost:5672
LISTEN_ADDRESS=10.29.2.1:12345 go run ./apiServer/apiServer.go &
LISTEN_ADDRESS=10.29.2.2:12345 go run ./apiServer/apiServer.go &

```



#### 计算散列值

```shell
echo -n "this is object test8 version 1" | openssl dgst -sha256 -binary | base64
```





dd if=/dev/urandom of=./UploadFiles/testfile1 bs=100000 count=10000





#### 测试PUT与GET

##### PUT数据

```shell
curl -v 10.29.2.1:12345/objects/dir1/test8:Trent -XPUT -d "this is object test8 version 1" -H "Digest: SHA-256=2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE="
```

##### get

默认最新版本

```shell
curl -v 10.29.2.1:12345/objects/dir1/test8:Trent
```

指定版本

```shell
curl -v 10.29.2.1:12345/objects/dir1/test8:Trent?version=3
```



##### put

```shell
curl -v 10.29.2.1:12345/objects/dir1/test1:Trent -XPUT -d "MyTest" -H "Digest: SHA-256=4S6el8A92j8xe21KhJj+4MLrnOSkPyrXXUd3sr9iMdw="
```

##### get

```shell
curl -v 10.29.2.1:12345/objects/dir1/test1:Trent
```





curl -v 10.29.2.1:12345/objects/TestDir/1.log:Wang5

```
openssl dgst -sha256 -binary /tmp/file |base64
curl -v 10.28.2.1:12345/objects/failfile:Trent -XPUT --data-binary ./9.log -H "Digest: SHA-256=F4VvXomxMWKqH4oVZms/bf9cooF3oJfoAaH3Aub4UCo="
```



#### 获取某一用户的全部文件

```shell
curl -v 10.29.2.1:12345/user/:Trent
```

#### 获取指定版本，GET方法添加一个版本名称

```shell
curl -v 10.29.2.1:12345/objects/test3:Trent?version=1
```

#### 删除文件命令，本质上只是添加一个hash为null的版本

```shell
curl -v 10.29.2.1:12345/objects/test3:Trent -XDELETE
```

#### 展示某一文件的所有版本

```shell
curl -v 10.29.2.1:12345/versions/dir1/test1:Trent
```





#### 结束

```shell
killall apiServer
```

```shell
killall dataServer
```



#### 维护工具

##### 删除旧版本

```shell
go run ../deleteOldMetadata/deleteOldMetadata.go
```

##### 六个数据服务节点上运行deleteOrphanObject，删除没有任何元数据引用的散列值，移到了garbage垃圾站内

```shell
STORAGE_ROOT=/tmp/1 LISTEN_ADDRESS=10.29.1.1:12345 go run ../deleteOrphanObject/deleteOrphanObject.go
```

##### 恢复降解的数据，在某个数据节点上运行

```shell
STORAGE_ROOT=/tmp/2 go run ../objectScanner/objectScanner.go
```





#### mysql安装和go包的安装

安装musql  sudo apt-get install mysql-server，并配置用户密码

go get github.com/go-sql-driver/mysql

go get github.com/jmoiron/sqlx









#### 创建的mysql数据库表结构

```
CREATE TABLE IF NOT EXISTS files(
   Name varchar(20),
   User VARCHAR(10),
   Version int default 1,
   Size bigint default 0,
   Hash varchar(60) default null,
   PRIMARY KEY (Name, User, Version)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;
```









#### 测试etcd，8个节点注册以及1个watcher

register

```shell
LISTEN_ADDRESS=10.29.1.1:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.2:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.3:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.4:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.5:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.6:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.7:12345 go run ./register/register.go &

LISTEN_ADDRESS=10.29.1.8:12345 go run ./register/register.go &

```

watcher

```shell
LISTEN_ADDRESS=10.29.2.1:12345 go run ./watcher/watcher.go
```







#### 重启后如何打开etcd

无法连接到etcd server，输入以下两个命令

```shell
# systemctl enable etcd
# systemctl restart etcd
```





#### 测试生成大量相同小文件

```shell
mkdir UploadFiles
for ((i = 1; i <= $1; i++)); do
    dd if=/dev/zero of=./UploadFiles/${i}.bin bs=${2}k count=1 &>/dev/null
done
```

#### 测试生成大量随机小文件

```shell
mkdir UploadFiles
for ((i = 1; i <= $1; i++)); do
    dd if=/dev/urandom of=./UploadFiles/${i}.log bs=`shuf -n 1 -i 0-16`k count=1 &>/dev/null
done
```











#### 关于curl报文，本项目必须的几个header

 curl -v **10.29.2.1:12345/objects/dir1/test1:Trent** -XPUT -d "MyTest" -H "Digest: SHA-256=4S6el8A92j8xe21KhJj+4MLrnOSkPyrXXUd3sr9iMdw="
> PUT /objects/dir1/test1:Trent HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.58.0
> Accept: */*
> Digest: SHA-256=4S6el8A92j8xe21KhJj+4MLrnOSkPyrXXUd3sr9iMdw=
> Content-Length: 6
> Content-Type: application/x-www-form-urlencoded





 curl -v **10.29.2.1:12345/objects/dir1/test1:Trent**
> GET /objects/dir1/test1:Trent HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.58.0
> Accept: */*





 curl -v **10.29.2.1:12345/versions/dir1/test1:Trent**
> GET /versions/dir1/test1:Trent HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.58.0
> Accept: */*





 curl -v **10.29.2.1:12345/user:Trent**
> GET /user:Trent HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.58.0
> Accept: */*










