package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	delayTime = 5
	address   = "http://118.24.158.102:30010/api/push"
)

func ProcessAfterDelay(messageReceiver string, messageIdServer string) {
	go func() {
		time.AfterFunc(delayTime*time.Second, func() {
			PushMethod(messageReceiver, messageIdServer)
		})
	}()
}

func ResetPushNumber(messageCreatorBody string) {
	messageCreator := gjson.Get(messageCreatorBody, "messageCreator").String()
	fmt.Println("Push messageCreator", messageCreatorBody)
	resetNumberOfPushes(messageCreator)
}

func PushMethod(messageReceiver string, messageIdServer string) {

	//***查找是需要推送
	isreaded := whetherNeedPushThisMessage(messageReceiver, messageIdServer)
	// fmt.Println("isreaded = ", isreaded)

	//***查找历史推送number
	pushNumber := addNumberOfPushes(messageReceiver)

	if isreaded == false {
		deviceToken, voideSetting, err := getUserDeviceToken(messageReceiver)
		fmt.Println("deviceToken = ", deviceToken, "voideSetting = ", voideSetting, "pushNumber = ", pushNumber)
		if deviceToken != "" && err == nil {
			//**开始推送
			// fmt.Sprintf("%d", pushNumber+1)
			voiceFile := "cat.mp3"
			if voideSetting == "dog" {
				voiceFile = "dog.mp3"
			}
			Http_Post(deviceToken, "您有一条新私信 📧", fmt.Sprintf("%d", pushNumber+1), voiceFile)

		}
	}

}

func Http_Post(token string, text string, pushNumber string, voideSetting string) error {
	//post的body内容,当前为json格式
	reqbody := `
        {
  		"notifications":
  				 [{
          			"tokens": ["` + token + `"],
      				"platform": 1,
      				"message": "` + text + `",
      				"badge" : ` + pushNumber + `,
      				"topic":"cn.ideas-lab.looker.ios-app",
      				"sound": "` + voideSetting + `",
   					"production": true,
      				"development":false
    				}]
		}`
	fmt.Println(reqbody)
	//创建请求
	postReq, err := http.NewRequest("POST",
		address,                    //post链接
		strings.NewReader(reqbody)) //post内容

	if err != nil {
		fmt.Println("POST请求:创建请求失败", err)
		return err
	}

	//增加header
	postReq.Header.Set("Content-Type", "application/json; encoding=utf-8")

	//执行请求
	client := &http.Client{}
	resp, err := client.Do(postReq)
	if err != nil {
		fmt.Println("POST请求:创建请求失败", err)
		return err
	} else {
		//读取响应
		body, err := ioutil.ReadAll(resp.Body) //此处可增加输入过滤
		if err != nil {
			fmt.Println("POST请求:读取body失败", err)
			return err
		}

		fmt.Println("POST请求:创建成功", string(body))
	}
	defer resp.Body.Close()
	return nil
}
