package service

import (
	"fmt"
	"github.com/kk88183080k/goWeb/msgo/orm"
	"log"
	"net/url"
)

// User 无tag场景
type User struct {
	Id       int64
	UserName string
	Password string
	Age      int
}

// User 无tag场景
type UserTag struct {
	Id       int64  `json:"id" msgo:"id"`
	UserName string `json:"userName" msgo:"user_name"`
	Password string `json:"password" msgo:"password"`
	Age      int    `json:"age" msgo:"age"`
}

func SaveUser() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       0,
		UserName: "ljw",
		Password: "ljw",
		Age:      18,
	}
	id, rows, err := db.NewSessionByTableName("user").Insert(user)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(id, rows)
	db.Close()
}

func SaveUserBatch() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	var users []any
	user := &User{
		UserName: "ljw",
		Password: "ljw",
		Age:      18,
	}
	users = append(users, user)
	user1 := &User{
		UserName: "张三",
		Password: "123456",
		Age:      18,
	}
	users = append(users, user1)

	user2 := &User{
		UserName: "李四",
		Password: "123456",
		Age:      30,
	}
	users = append(users, user2)

	user3 := &User{
		UserName: "王五",
		Password: "123456",
		Age:      30,
	}
	users = append(users, user3)
	id, rows, err := db.NewSessionByTableName("user").InsertBatch(users)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(id, rows)
	db.Close()
}

func SaveUserBatch1() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	var users []any
	user := &UserTag{
		UserName: "ljw",
		Password: "ljw",
		Age:      18,
	}
	users = append(users, user)
	user1 := &UserTag{
		UserName: "张三",
		Password: "123456",
		Age:      18,
	}
	users = append(users, user1)

	user2 := &UserTag{
		UserName: "李四",
		Password: "123456",
		Age:      30,
	}
	users = append(users, user2)

	user3 := &UserTag{
		UserName: "王五",
		Password: "123456",
		Age:      30,
	}
	users = append(users, user3)
	id, rows, err := db.NewSessionByTableName("user").InsertBatch(users)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(id, rows)
	db.Close()
}

func UpdateUser() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       1,
		UserName: "ljw",
		Password: "ljw",
		Age:      18,
	}
	rows, err := db.New(user).UpdateField("age", 10).UpdateField("password", "1234").Where("id", user.Id).And().Where("user_name", user.UserName).Update()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(rows)
	db.Close()
}

func UpdateUser1() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       1,
		UserName: "ljw",
		Password: "ljw",
		Age:      18,
	}
	rows, err := db.New(user).UpdateField("age", 10).UpdateField("password", "1234").Where("id", user.Id).Or().Where("user_name", user.UserName).Update()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(rows)
	db.Close()
}

func UpdateUserObject() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       1,
		UserName: "ljw",
		Password: "pskdfsdf",
		Age:      35,
	}
	rows, err := db.New(user).UpdateObject(user)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(rows)
	db.Close()
}

func DeleteUserObject() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       1,
		UserName: "ljw",
		Password: "pskdfsdf",
		Age:      35,
	}
	rows, err := db.New(user).DeleteObject(user)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(rows)
	db.Close()
}

func DeleteUserWhere() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{
		Id:       2,
		UserName: "ljw",
		Password: "pskdfsdf",
		Age:      35,
	}
	rows, err := db.New(user).Where("id", user.Id).Delete()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(rows)
	db.Close()
}

func SelectUserWhere() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{}
	err := db.New(user).Where("id", 4).SelectOne(user)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(user)
	db.Close()
}

func SelectListUserWhere() {
	dbUrl := fmt.Sprintf("root:123456@tcp(localhost:3306)/msgo?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	db := orm.Open("mysql", dbUrl)
	user := &User{}
	rs, err := db.New(user).SelectList(user)
	if err != nil {
		log.Println(err)
		return
	}
	for i, r := range rs {
		fmt.Println(i, r)
	}
	db.Close()
}
