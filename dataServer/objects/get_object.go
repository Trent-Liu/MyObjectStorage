package objects

import (
	"awesomeProject/dataServer/locate"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func getFile(name string) string {
	//根据对象的散列值<hash>找到$STORAGE_ROOT/objects/<hash>.X对象文件
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + name + ".*")
	if len(files) != 1 {
		return ""
	}
	file := files[0]

	h := sha256.New()
	sendFile(h, file)

	//对这个对象的内容计算SHA-256散列值，使用url.PathEscape转义，最后得到的就是可以用于URL的散列值字符串
	d := url.PathEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	hash := strings.Split(file, ".")[2]
	if d != hash {
		//如果不一致则打印错误日志，并从缓存和磁盘上删除对象，返回空字符串，如果一致则返回对象的文件名
		//这里进行校验删除是用于防止存储系统的数据降价，哪怕是上传正确的数据也可能随时间的流逝而损坏。
		log.Println("object hash mismatch, remove", file)
		locate.Del(hash)
		os.Remove(file)
		return ""
	}
	return file
}
