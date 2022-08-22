package objects

import (
	"compress/gzip"
	"io"
	"log"
	"os"
)

//用于写入对象数据的w和对象的文件名file
func sendFile(w io.Writer, file string) {
	//调用os.Open打开对象文件
	f, e := os.Open(file)
	if e != nil {
		log.Println(e)
		return
	}
	defer f.Close()
	gzipStream, e := gzip.NewReader(f)
	if e != nil {
		log.Println(e)
		return
	}
	//写入w
	io.Copy(w, gzipStream)
	gzipStream.Close()
}
