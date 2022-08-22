package temp

import (
	"awesomeProject/dataServer/locate"
	"awesomeProject/src/lib/calculatehash"
	"compress/gzip"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}

//调用os.Rename将临时对象的数据文件改名。
//重命名时，需要读取临时对象的数据并计算散列值
func commitTempObject(datFile string, tempinfo *tempInfo) {
	f, _ := os.Open(datFile)
	defer f.Close()
	d := url.PathEscape(calculatehash.CalculateHash(f))
	f.Seek(0, io.SeekStart)
	//创建正式文件w，以w为参数调用gzip.NewWriter创建w2
	w, _ := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + tempinfo.Name + "." + d)
	w2 := gzip.NewWriter(w)
	//将临时文件f中的数据复制进w2
	io.Copy(w2, f)
	w2.Close()
	//最后删除临时对象文件并添加对象定位缓存。
	os.Remove(datFile)
	//调用locate.Add将<hash>为键，分片的id为值，加入数据服务的对象定位缓存。
	locate.Add(tempinfo.hash(), tempinfo.id())
}
