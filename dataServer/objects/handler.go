package objects

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	//删除PUT方法，因为现在的数据服务的对象上传完全依靠temp接口的临时对象转正，所以不再需要object接口的PUT方法。
	if m == http.MethodGet {
		//读取对象时进行一次数据校验
		get(w, r)
		return
	}
	if m == http.MethodDelete {
		del(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
