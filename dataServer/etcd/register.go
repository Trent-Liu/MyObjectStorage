package etcd

import (
	"log"
	"os"
	"strings"
)

var ser *ServiceRegister
var e error

func Register() {
	//etcd服务器地址
	var endpoints = []string{"localhost:2379"}
	tmp := strings.Split(os.Getenv("LISTEN_ADDRESS"), ".")[3]
	tmp = strings.Split(tmp, ":")[0]
	key := "/server/node" + tmp
	//本节点注册etcd，这里节点的地址为localhost:8000
	ser, e = NewServiceRegister(endpoints, key, os.Getenv("LISTEN_ADDRESS"), 6, 5)
	if e != nil {
		log.Fatalln(e)
	}
	////监听续租相应chan
	go ser.ListenLeaseRespChan()
	//
	//// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
	//c := make(chan os.Signal, 1)
	//go func() {
	//	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//}()
	//log.Printf("exit %s", <-c)

}

func CloseRegister() {
	ser.Close()
}
