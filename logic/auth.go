package main

import (
	"github.com/tidwall/gjson"
	"goim/libs/define"
	// "strings"
	//***abc自己的代码
	"errors"
	"fmt"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int32, err error)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string) (userId int64, roomId int32, err error) {
	fmt.Println("token = ", token)
	var whetherPutInRoom bool
	var userIdTemp int64

	whetherPutInRoom, userIdTemp, err = checkUserChatID(token)

	if err == nil {
		if whetherPutInRoom {
			userId = userIdTemp
			roomId = define.QXChatRoom
		} else {
			roomId = define.NoRoom
		}
	}

	return
}

//**检查是否有这个userid(token = chatID)
func checkUserChatID(authBody string) (whetherPutInRoom bool, userId int64, err error) {
	//1.  获得传过来的userid和password
	userName := gjson.Get(authBody, "UserName").String()
	password := gjson.Get(authBody, "Password").String()
	fmt.Println(userName, " - ", password)

	//2.  搜索数据库里的userid和password

	//3. session全部放在room里
	whetherPutInRoom = true

	//3.Compare
	// fmt.Println("userName = ", userName, "password = ", password)
	userId, result, err := authIMUser(userName, password)
	fmt.Println("userId = ", userId, "result = ", result, "err = ", err)

	if err != nil || result == false {
		userId = -1
		err = errors.New("user does not exist")
	}

	return
}

//***根据userName 查询chatId
func changeUserNameToUserId(userAccount string) (userId int64, err error) {

	// fmt.Println("userAccount = ", userAccount)
	userId, err = getUserChatId(userAccount)

	if err != nil {
		userId = -1
	}

	return

}
