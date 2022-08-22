package objects

import (
	"awesomeProject/apiServer/hashlocate"
	"awesomeProject/apiServer/locate"
	"awesomeProject/src/lib/calculatehash"
	"awesomeProject/src/lib/mysql"
	"awesomeProject/src/lib/rsconstruct"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

/*
1. 检查报文
2. 存储文件
3. 添加版本到数据库
*/
func put(w http.ResponseWriter, r *http.Request) {

	//1
	//从HTTP请求头部获取对象的唯一标识哈希值
	hash := getHash(r.Header)

	//客户端构建的报文错误
	if hash == "" {
		log.Println("no hash in header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//2
	//size参数，因为PUT需要在一开始就确定临时对象的大小。
	size := getSize(r.Header)

	//存
	c, e := storeAnObject(r.Body, hash, size)

	if e != nil {
		log.Println(e)
		w.WriteHeader(c)
		return
	}
	if c != http.StatusOK {
		w.WriteHeader(c)
		return
	}

	//3
	//获取文件路径以及用户版本等，添加进数据库
	tmpname := strings.Split(r.URL.EscapedPath(), "/")[2:]
	name := ""
	for _, tmp := range tmpname {
		name += "/" + tmp
	}
	name = strings.Split(name, ":")[0]
	user := strings.Split(r.URL.EscapedPath(), ":")[1]
	//为该数据添加版本，若以前没有则添加版本1
	e = mysql.AddVersion(name, user, hash, size)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getHash(h http.Header) string {
	//先获取digest头部
	digest := h.Get("digest")
	if len(digest) < 9 {
		return ""
	}
	//开头必须是SHA-256，之后才可以返回后面的散列值
	if digest[:8] != "SHA-256=" {
		return ""
	}
	return digest[8:]
}

func getSize(h http.Header) int64 {
	//获取content-length，并调用strconv.ParseInt将字符串转化为int64输出。
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

//将数据存储进数据节点
func storeAnObject(r io.Reader, hash string, size int64) (int, error) {

	//1
	//首先判断这个hash值表示的文件是否存在，提高性能
	if locate.Exist(url.PathEscape(hash)) {
		return http.StatusOK, nil
	}

	//url.PathEscape转义字符串，使其更加安全
	//为流获取六个数据节点地址
	stream, e := putStreamShard(url.PathEscape(hash), size)
	if e != nil {
		return http.StatusInternalServerError, e
	}

	//stream为RSPutStream，可以将数据编解码

	//reader被读取时，读取器r，写入器stream，返回的reader类型指的是从声明的r读取，写入给定的w
	reader := io.TeeReader(r, stream)
	//此时已经生成了存在数据节点中的临时文件，再校验哈希值
	d := calculatehash.CalculateHash(reader)

	//散列值进行比较
	if d != hash {
		//不一致则stream.Commit(false)删除临时对象
		stream.Commit(false)
		//并返回400 Bad Request
		return http.StatusBadRequest, fmt.Errorf("object hash mismatch, calculated=%s, requested=%s", d, hash)
	}
	//一致则调用stream.Commit(true)将临时对象转正并且返回200OK
	stream.Commit(true)
	return http.StatusOK, nil
}

//获取6个数据服务节点，生成数据流
func putStreamShard(hash string, size int64) (*rsconstruct.RSPutStream, error) {
	//获取6个随机数据服务节点
	servers := hashlocate.GetServerList(hash)
	if len(servers) != rsconstruct.ALL_SHARDS {
		return nil, fmt.Errorf("cannot find enough dataServer")
	}

	dataservers := make([]string, rsconstruct.ALL_SHARDS)
	for index, dataserver := range servers {
		dataservers[index] = dataserver
	}
	//调用NewRSPutStream生成一个数据流
	return rsconstruct.NewRSPutStream(dataservers, hash, size)
}
