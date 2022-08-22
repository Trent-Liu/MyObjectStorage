package main

import (
	"awesomeProject/dataServer/etcd"
	"awesomeProject/dataServer/locate"
	"awesomeProject/dataServer/objects"
	"awesomeProject/dataServer/temp"
	"log"
	"net/http"
	"os"
)

func main() {
	//启动线程执行Locate，数据服务的locate包是用来对节点本地磁盘上的对象进行定位的
	//减少对磁盘访问的次数，提高磁盘的性能，程序启动时扫描一遍本地磁盘，将磁盘中所有对象的散列值读入内存。
	//之后只需要搜索内存即可
	locate.CollectObjects()
	go locate.StartLocate()

	etcd.Register()

	//HTTP处理函数
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))

	etcd.CloseRegister()
}
