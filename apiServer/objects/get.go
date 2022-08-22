package objects

import (
	"awesomeProject/apiServer/hashlocate"
	"awesomeProject/src/lib/mysql"
	"awesomeProject/src/lib/rsconstruct"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func get(w http.ResponseWriter, r *http.Request) {
	//文件名，因为报文中格式为 XX.XX.XX.XX:12345/objects/file，按照/分割后下标为2的位置为文件名
	tmpname := strings.Split(r.URL.EscapedPath(), "/")[2:]
	name := ""
	for _, tmp := range tmpname {
		name += "/" + tmp
	}
	name = strings.Split(name, ":")[0]
	user := strings.Split(r.URL.EscapedPath(), ":")[1]

	//查询出version的值
	versionId := r.URL.Query()["version"]
	version := 0
	var e error
	if len(versionId) != 0 {
		version, e = strconv.Atoi(versionId[0])
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	//以对象的名字和版本号为参数调用 mysql.GetMetadata
	//meta.Hash为对象的散列值
	meta, e := mysql.GetMetadata(name, user, version)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//散列值为空表示该版本是一个删除标记
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash := url.PathEscape(meta.Hash)

	//RS码的实现要求每一个数据片的长度完全一样，在编码时如果对象长度不能被4整除，函数就会对最后一个数据片进行填充。
	//因此解码时必须提供对象的准确长度，防止填充数据被当成原始对象数据返回
	stream, e := GetStream(hash, meta.Size)

	//此时得到的stream类型是一个指向rs.RSGetStream的结构体指针，GET对象时对缺失的分片进行修复。
	//修复的过程也使用数据服务的temp接口

	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//get函数多了一个对于Accept-Encoding请求头部的检查，如果该头部中含有gzip，则说明客户端可以接收gzip压缩数据。
	acceptGzip := false
	encoding := r.Header["Accept-Encoding"]
	for i := range encoding {
		if encoding[i] == "gzip" {
			acceptGzip = true
			break
		}
	}

	//可以接收gzip数据压缩，设置Content-Encoding响应头部为gzip
	if acceptGzip {
		w.Header().Set("content-encoding", "gzip")
		//以w为参数调用gzip.NewWriter创建一个指向gzip结构体的指针w2.
		w2 := gzip.NewWriter(w)
		//对象数据流stream内容用io.Copy写入w2，数据自动压缩。
		io.Copy(w2, stream)
		w2.Close()
	} else {
		io.Copy(w, stream)
	}

	//在流关闭时将临时对象转正。
	stream.Close()
}

func GetStream(hash string, size int64) (*rsconstruct.RSGetStream, error) {
	//首先根据对象散列值hash定位对象
	locateInfo := hashlocate.GetServerList(hash)
	//如果反馈的定位结果locateInfo数组长度小于4则返回错误

	//此处出错
	if len(locateInfo) < rsconstruct.DATA_SHARDS {
		return nil, fmt.Errorf("object %s locate fail, result %v", hash, locateInfo)
	}

	//创建rs.RSGetStream
	return rsconstruct.NewRSGetStream(locateInfo, hash, size)
}
