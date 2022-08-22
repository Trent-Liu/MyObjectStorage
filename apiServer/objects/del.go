package objects

import (
	"awesomeProject/src/lib/mysql"
	"log"
	"net/http"
	"strings"
)

func del(w http.ResponseWriter, r *http.Request) {
	//获取名称
	tmpname := strings.Split(r.URL.EscapedPath(), "/")[2:]
	name := ""
	for _, tmp := range tmpname {
		name += "/" + tmp
	}
	name = strings.Split(name, ":")[0]
	user := strings.Split(r.URL.EscapedPath(), ":")[1]

	//获取最新版本
	version, e := mysql.SearchLatestVersion(name, user)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//插入一条新的元数据，最新版本号+1， size值为0， hash为空字符串，删除标记。
	e = mysql.PutMetadata(name, user, version.Version+1, 0, "")
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
