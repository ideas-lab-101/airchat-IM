package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/robfig/cron"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
	xiaomipush "xiaomi_push"
)

const (
	address   = "http://127.0.0.1:8088/api/push"
)

var (
	andorid_push_timer *cron.Cron = nil
	aliasArr []string
)

func ProcessGeneralMethod(messageReceiver string, messageIdServer string, pushcode int32) {
	go func() {
		PushMethod(messageReceiver, messageIdServer, pushcode)
	}()
}

func ResetPushNumber(messageCreatorBody string) {
	messageCreator := gjson.Get(messageCreatorBody, "messageCreator").String()
	// fmt.Println("Push messageCreator", messageCreatorBody)
	resetNumberOfPushes(messageCreator)
}

func PushMethod(messageReceiver string, messageIdServer string, pushcode int32) {
	deviceToken, voideSetting, osType, err := getUserDeviceToken(messageReceiver)
	//fmt.Println("deviceToken = ", deviceToken, "osType = ", osType)

	if osType == "Android" {
		if len(aliasArr) < 1000 {
			aliasArr = append(aliasArr, deviceToken)
		}
		pushTimer()
	}else{
		//***查找是需要推送
		go func() {
			//time.AfterFunc(delayTime * time.Second, func() {
			//	isReaded := whetherNeedPushThisMessage(messageReceiver, messageIdServer)
			//	if  isReaded == false{
			//	}
			//})
			//***查找历史推送number
			pushNumber := addNumberOfPushes(messageReceiver)
			if deviceToken != "" && err == nil {
				//**开始推送
				voiceValue := "cat.mp3"
				if voideSetting != ""{
					voiceValue = voideSetting
				}
				Http_iOS_Post(deviceToken, "您有一条新私信 📧", fmt.Sprintf("%d", pushNumber+1), voiceValue, fmt.Sprintf("%d", pushcode))
			}
		}()
	}
}

func OtherPushMethod(receiveAccount string, content string) {
	deviceToken, voiceSetting, osType, err := getUserDeviceToken(receiveAccount)
	if deviceToken != "" && err == nil {
		//**开始推送
		voiceValue := "cat.mp3"
		if voiceSetting != ""{
			voiceValue = voiceSetting
		}
		if osType == "Android" {
			httpMiPushPostOne(deviceToken, content);
		}else{
			Http_iOS_Post(deviceToken, content, "1", voiceValue, fmt.Sprintf("%d", "1"))
		}
	}
}


/***
*
* 安卓推送极光免费的有api限制，所以采用了定时器设置
* 安卓定时器方法
* 每五秒查询一次，筛选出需要推送的数据，再统一推送
*
***/

//开启推送定时器
func pushTimer(){
	spec := "* * * * * *"  //分 时 日 月 星期 要运行的命令
	if  andorid_push_timer == nil{
		andorid_push_timer = cron.New()
		andorid_push_timer.AddFunc(spec, timerCheckMethod)
		andorid_push_timer.Start()
		select {}
	}
}

//每秒钟执行一次
func timerCheckMethod() { //安卓推送使用定时器
	if len(aliasArr) > 0 {
		//***小米推送
		httpMiPushPost()
		//***极光推送
		jPushPostMethod()
	}
	//**清空 列表
	aliasArr = aliasArr[0:0]
}

/***
*
* 小米推送
*
***/
//****推送一条
func httpMiPushPostOne(alias, content string) error{
	var client = xiaomipush.NewClient("CULJYF3/cBXND2BsTTgL6Q==", []string{"com.android.crypt.chatapp"})
	var msg1 = xiaomipush.NewAndroidMessage(content, "点击查看").SetPassThrough(1).SetPayload(content).SetNotifyType(-1).SetTimeToLive(863000000).AddExtra("notify_foreground", "0").AddExtra("notify_effect","1")
	result, err :=  client.SendToAlias(context.Background(),msg1, alias)
	fmt.Println("result = ", result, " err = ", err)
	return nil
}


func httpMiPushPost() error { //开始推送
	var client = xiaomipush.NewClient("CULJYF3/cBXND2BsTTgL6Q==", []string{"com.android.crypt.chatapp"})
	var msg1 = xiaomipush.NewAndroidMessage("你有一条新私信", "点击查看").SetPassThrough(0).SetPayload("你有一条新私信").SetNotifyType(-1).SetTimeToLive(863000000).AddExtra("notify_foreground", "0")
	result, err :=  client.SendToAliasList(context.Background(),msg1, aliasArr)
	fmt.Println("result = ", result, " err = ", err)
	return nil
}

/***
*
* 极光推送
*
***/
func jPushPostMethod(){  //拼接所有需要推送的alias
	pushAlias := ""
	for i:=0; i < len(aliasArr); i++{
		alia := aliasArr[i];
		pushAlias = pushAlias + "\"" + alia + "\"" + ","
	}
	pushAlias = pushAlias[0:len(pushAlias) - 1]
	httpJpushPost(pushAlias)
}

func httpJpushPost(alias string) error { //开始推送
	reqbody := `
        {
  			"platform": ["android"],
    		"audience": {
					"alias" : [` + alias + `]
			},
        	"message": {
        			"msg_content": "你有一条新私信",
        			"content_type": "text",
        			"title": "你有一条新私信"
    		},
			"options": {
					"time_to_live": 863000
			}
		}`
	fmt.Println("reqbody = ", reqbody)
	//创建请求
	postReq, err := http.NewRequest("POST",
		"https://api.jpush.cn/v3/push", //post链接
		strings.NewReader(reqbody))     //post内容

	if err != nil {
		fmt.Println("POST请求:创建请求失败", err)
		return err
	}

	//增加header
	app_key := "4e418685ef61c5e358786f64:1dabacef0eb60e6ea49b1fa3"
	app_key_bytes := []byte(app_key)

	appKey := "Basic " + base64.StdEncoding.EncodeToString(app_key_bytes)
	fmt.Println("appKey = ", appKey)
	postReq.Header.Set("Content-Type", "application/json; encoding=utf-8")
	postReq.Header.Add("Authorization", appKey)

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


/***
*
* iOS推送 每一条消息都推送，直接推送
*
***/
func Http_iOS_Post(token string, text string, pushNumber string, voideSetting string, pushcode string) error {
	//post的body内容,当前为json格式
	reqbody := `
        {
  		"notifications":
  				 [{
          			"tokens": ["` + token + `"],
      				"platform": 1,
      				"message": "` + text + `",
      				"badge" : ` + pushNumber + `,
      				"pushcode" : ` + pushcode + `,
      				"topic":"Cryeye.Inc.ChatCare",
      				"sound": "` + voideSetting + `",
   					"production": true,
      				"development":false
    				}]
		}`
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
