package user

import (
	"awesomeProject/src/lib/mysql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

//实现GET＋user，即获取某用户存储的所有文件目录。
func Handler(w http.ResponseWriter, r *http.Request) {
	//首先检查HTTP方法是否为GET，不为GET则返回。
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tmpname := strings.Split(r.URL.EscapedPath(), "/")[2:]
	name := ""
	for _, tmp := range tmpname {
		name += "/" + tmp
	}
	name = strings.Split(name, ":")[0]
	user := strings.Split(r.URL.EscapedPath(), ":")[1]

	//无限循环中调用SearchAllVersions，会返回一个元数据的数组。
	metas, e := mysql.SearchUserAllObjects(user)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//遍历元数据数组，元数据一一写入HTTP响应正文。
	for i := range metas {
		b, _ := json.Marshal(metas[i])
		w.Write(b)
		w.Write([]byte("\n"))
	}

}
