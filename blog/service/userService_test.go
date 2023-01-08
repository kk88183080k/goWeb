package service

import (
	"github.com/kk88183080k/goWeb/msgo/orm"
	"log"
	"testing"
)

func TestSaveUser(t *testing.T) {
	SaveUser()
}

func TestSaveUserBatch(t *testing.T) {
	SaveUserBatch()
}

func TestSaveUserBatch1(t *testing.T) {
	SaveUserBatch1()
}

func TestUpdateUser(t *testing.T) {
	UpdateUser()
}

func TestUpdateUser1(t *testing.T) {
	UpdateUser1()
}

func TestUpdateUserObject(t *testing.T) {
	UpdateUserObject()
}

func TestDeleteUserObject(t *testing.T) {
	DeleteUserObject()
}

func TestDeleteUserWhere(t *testing.T) {
	DeleteUserWhere()
}

func TestSelectUserWhere(t *testing.T) {
	SelectUserWhere()
}

func TestSelectUserListWhere(t *testing.T) {
	SelectListUserWhere()
}

func TestName(t *testing.T) {
	tableName := "msgoUser"
	log.Println(tableName, orm.Name(tableName))

	tableName = "MsgoUser"
	log.Println(tableName, orm.Name(tableName))

	tableName = "MsgoUseR"
	log.Println(tableName, orm.Name(tableName))

	tableName = "MsgoUseR"
	log.Println(tableName, orm.Name(tableName))
}
