package main

//mysql -h 121.42.237.244 -u root -p
import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

//******* 私信处理
func authIMUser(userName string, password string) (userId int64, result bool, err error) {

	result = true
	url := "http://airchat.ideas-lab.cn/api/system/v2/authUser?account=" + userName + "&password=" + password

	resp, err := http.Get(url)
	if err != nil {
		result = false
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result = false
		return
	}

	code := gjson.Get(string(body), "code").Int()
	if code == 1 {
		dataString := gjson.Get(string(body), "data").String()
		userInfoString := gjson.Get(string(dataString), "userInfo").String()
		
		id := gjson.Get(userInfoString, "id").Int()
		token := gjson.Get(userInfoString, "token").String()
		voiceSettings := gjson.Get(userInfoString, "voice_settings").String()
		os_type := gjson.Get(userInfoString, "os_type").String()
		userId = id

		insertUserInfo(userName, token, userId, voiceSettings, os_type)
	} else {
		result = false
	}

	return
}

func getUserChatId(userAccount string) (userChatID int64, err error) {
	userInfo, err := getUserId(userAccount)
	userChatID = userInfo.UserId
	// fmt.Println("userChatID", userChatID, " err = ", err)
	return
}

//********* 推送的处理
func getUserDeviceToken(userAccount string) (deviceToken, voiceSetting, osType string, err error) {

	userInfo, err := getUserDeviceTokenMoethod(userAccount)
	deviceToken = userInfo.DeviceToken
	voiceSetting = userInfo.VoiceSettings
	osType = userInfo.OsType

	return
}
