package etcd

import (
	"log"
)

var Ser *ServiceDiscovery
var ServiceLength int

func Watcher() {
	var endpoints = []string{"localhost:2379"}
	Ser = NewServiceDiscovery(endpoints)
	defer Ser.Close()

	err := Ser.WatchService("/server/")
	if err != nil {
		log.Fatal(err)
	}
	ServiceLength = len(Ser.serverList)

	//// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
	//c := make(chan os.Signal, 1)
	//go func() {
	//	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//}()
	//
	//for {
	//	select {
	//	case <-time.Tick(10 * time.Second):
	//		log.Println(ser.GetServices())
	//	case <-c:
	//		log.Println("server discovery exit")
	//		return
	//	}
	//}
}
