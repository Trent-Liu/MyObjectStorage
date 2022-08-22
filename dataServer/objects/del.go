package objects

import (
	"awesomeProject/dataServer/locate"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//
func del(w http.ResponseWriter, r *http.Request) {
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	//根据对象散列值搜索对象文件。
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + hash + ".*")
	if len(files) != 1 {
		return
	}
	//将该散列值移出对象定位缓存
	locate.Del(hash)
	//将对象文件移动到$%=STORAGE_ROOT/garbage/目录下
	os.Rename(files[0], os.Getenv("STORAGE_ROOT")+"/garbage/"+filepath.Base(files[0]))
}
