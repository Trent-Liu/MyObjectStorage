package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Metadata struct {
	Name    string
	User    string
	Version int
	Size    int64
	Hash    string
}

//数据库指针
var db *sqlx.DB

//初始化数据库连接，init()方法系统会在动在main方法之前执行。
func InitDB() {
	database, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/mystorage")
	if err != nil {
		fmt.Println("open mysql failed,", err)
	}
	err = database.Ping()
	if err != nil {
		fmt.Println("open mysql failed,", err)
	}
	db = database
}

//用于根据对象的名字和版本号来获取对象的元数据
func getMetadata(name string, user string, versionId int) (meta Metadata, e error) {
	sql := "select Name, User, Version, Size, Hash from files where Name = ? and Version = ? and User = ?"
	err := db.QueryRow(sql, name, versionId, user).Scan(&meta.Name, &meta.User, &meta.Version, &meta.Size, &meta.Hash)
	if err != nil {
		return
	}
	return meta, err

}

//先获取最新版本，然后再版本号上+1调用PutMetadata
func AddVersion(name, user, hash string, size int64) error {
	version, e := SearchLatestVersion(name, user)
	if e != nil {
		return e
	}
	return PutMetadata(name, user, version.Version+1, size, hash)
}

//以对象的名字为参数，调用ES搜索API
func SearchLatestVersion(name, user string) (meta Metadata, e error) {
	//版本号以降序排列只返回第一个结果，为最新版本
	sql := "select Name, User, Version, Size, Hash from files where Name = ? and User = ? order by Version desc limit 1"
	err := db.QueryRow(sql, name, user).Scan(&meta.Name, &meta.User, &meta.Version, &meta.Size, &meta.Hash)
	if err != nil {
		return
	}

	return
}

//封装getMetadata，区别只有当version为0时，会调用SearchLatestVersion获取当前最新版本
func GetMetadata(name, user string, version int) (metadata Metadata, e error) {
	if version == 0 {
		return SearchLatestVersion(name, user)
	}
	return getMetadata(name, user, version)
}

//向ES服务上传一个新的元数据，四个参数对应数据的四个属性
func PutMetadata(name string, user string, version int, size int64, hash string) error {
	sql := "insert into files(Name, User, Version, Size, Hash) values (?, ?, ?, ?, ?)"
	result, err := db.Exec(sql, name, user, version, size, hash)
	if err != nil {
		fmt.Printf("insert data failed, err:%v\n", err)
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		fmt.Printf("get insert lastInsertId failed, err:%v\n", err)
		return err
	}
	return err
}

//搜索某个对象或者所有对象的全部版本
//输入参数name表示对象的名字，不为空搜索该对象，为空搜索所有对象
func SearchAllVersions(name string, user string) (metas []Metadata, e error) {
	sql := "select Name, User, Version, Size, Hash from files where Name = ? and User = ?"
	rows, err := db.Query(sql, name, user)
	if err != nil {
		fmt.Printf("query data failed，err:%s\n", err)
		return
	}

	for rows.Next() {
		meta := Metadata{}
		err := rows.Scan(&meta.Name, &meta.User, &meta.Version, &meta.Size, &meta.Hash)
		if err != nil {
			fmt.Printf("scan data failed, err:%v\n", err)
			return
		}
		metas = append(metas, meta)
	}
	return

}

func SearchUserAllObjects(user string) (metas []Metadata, e error) {
	sql := "select Name, User, Version, Size, Hash from files where User = ? order by Name, Version"
	rows, err := db.Query(sql, user)
	if err != nil {
		fmt.Printf("query data failed，err:%s\n", err)
		return
	}

	for rows.Next() {
		meta := Metadata{}
		err := rows.Scan(&meta.Name, &meta.User, &meta.Version, &meta.Size, &meta.Hash)
		if err != nil {
			fmt.Printf("scan data failed, err:%v\n", err)
			return
		}
		metas = append(metas, meta)
	}
	return

}

func DelMetadata(name string, user string, version int) {
	sql := "insert into files(Name, User, Version) values (?, ?, ?)"
	result, err := db.Exec(sql, name, user, version)
	if err != nil {
		fmt.Printf("delete data failed, err:%v\n", err)
		return
	}
	_, err = result.LastInsertId()
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	return

}

//Bucket结构体
type Bucket struct {
	Key         string   //字符串Key，表示对象的名字
	User        string   //对象用户
	Doc_count   int      //整型Doc_count，表示该对象目前的版本数量
	Min_version struct { //当前对象的最小版本号
		Value float32
	}
}

//把版本数量大于等于min_doc_count的对象都搜索出来保存在Bucket结构体的数组里
func SearchVersionStatus(min_doc_count int) (buckets []Bucket, e error) {
	sql := "select Name, User, count(Version) from files group by Name, User having count(Version) > ?"
	rows, err := db.Query(sql, min_doc_count)
	if err != nil {
		fmt.Printf("query data failed，err:%s\n", err)
		return
	}

	for rows.Next() {
		bucket := Bucket{}
		err := rows.Scan(&bucket.Key, &bucket.User, &bucket.Doc_count)
		if err != nil {
			fmt.Printf("scan data failed, err:%v\n", err)
			return
		}
		buckets = append(buckets, bucket)
	}

	for index, bucket := range buckets {
		sql := "select Version from files where Name = ? and User = ? order by Version limit 1"
		err := db.QueryRow(sql, bucket.Key, bucket.User).Scan(&buckets[index].Min_version.Value)
		if err != nil {
			return
		}
	}

	return
}

//搜索所有对象元数据中hash属性等于散列值的文档，如果满足条件的文档数量不为0，说明还存在对该散列值的引用，返回true，否则返回false
func HasHash(hash string) (bool, error) {
	sql := "select Name from files where Hash = ?"
	rows, err := db.Query(sql, hash)
	if err != nil {
		return false, err
	}
	i := 0
	for rows.Next() {
		i++
	}
	return i != 0, nil
}

func SearchHashSize(hash string) (size int64, e error) {
	sql := "select Size from files where Hash = ? limit 1"
	err := db.QueryRow(sql, hash).Scan(&size)
	if err != nil {
		return
	}
	return
}
