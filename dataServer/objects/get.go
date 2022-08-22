package objects

import (
	"net/http"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	//首先从URL中获取对象的散列值，然后以散列值为参数调getFile获得对象的文件名file
	file := getFile(strings.Split(r.URL.EscapedPath(), "/")[2])
	if file == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//将对象文件的内容输出到HTTP响应。
	sendFile(w, file)
}
