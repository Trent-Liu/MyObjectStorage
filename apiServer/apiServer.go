package main

import (
	"awesomeProject/apiServer/etcd"
	"awesomeProject/apiServer/objects"
	"awesomeProject/apiServer/user"
	"awesomeProject/apiServer/versions"
	"awesomeProject/src/lib/mysql"
	"log"
	"net/http"
	"os"
)

func main() {
	//启动线程执行ListenHeartBeat
	go etcd.Watcher()
	//连接数据库
	mysql.InitDB()
	//处理正常的对象请求
	http.HandleFunc("/objects/", objects.Handler)
	//处理版本
	http.HandleFunc("/versions/", versions.Handler)
	//处理版本
	http.HandleFunc("/user/", user.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
