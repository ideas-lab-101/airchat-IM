package main

import (
	"github.com/tidwall/gjson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "log"
	log "github.com/thinkboy/log4go"
	"strconv"
	"time"
	//***abc自己的代码
	"fmt"
)

type MessageMongo struct {
	Usercreaterchatid  string
	Userreceiverchatid string
	MessageBody        string
	Messagegettime     string
	MessageIdClient    string
	MessageIdServer    string
	Isreaded           bool
	LastPushedTime     string //是否被推送过
}

type MessagePush struct {
	MessagePushId  string
	PushNumber     int32
	LastPushedTime string
}

type MessageRecalled struct {
	MessageCreator  string
	MessageReceiver string
	MessageIdClient string
	RecalledType    int32 //默认是0，根据MessageCreator和MessageIdClient撤回一条消息
	IsRecalled      bool
}

type SpecialMessage struct {
	ReceiveAccount  string
	CreateAccount  string
	MessageKind     string
	MessageIdClient string
	Content         string
	IsReceived      bool
}

type UserInfomation struct {
	Account       string
	DeviceToken   string
	UserId        int64
	VoiceSettings string
	OsType        string
}

const (
	databaseUrl    = "127.0.0.1:27017"
	database       = "qxProject"
	source         = "qxProject"
	sourcePush     = "qxPush"
	sourceRecall   = "qxRecall"
	sourceSpecial  = "qxSpecial"
	sourceUserInfo = "qxUserInfo"
	username       = "qxUser"
	password       = "Aa135412."
)

var (
	golalSession *mgo.Session
)

/******
   ****
   ****  持久化缓存一些聊天中数据
   ****
******/

func insertUserInfo(userName string, deviceToken string, userId int64, voiceSettings string, os_type string) {
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}

	defer session.Close()
	c := session.DB(database).C(sourceUserInfo)

	//fmt.Println("userName = ", userName, " deviceToken = ", deviceToken, " userId = ", userId, " voiceSettings = ", voiceSettings)
	err = c.Update(bson.M{"account": userName}, bson.M{"$set": bson.M{"devicetoken": deviceToken, "voicesettings": voiceSettings, "ostype" : os_type}})
	if err != nil {
		log.Debug("userInfo消息状态更新错误:", err)

		err = c.Insert(&UserInfomation{userName, deviceToken, userId, voiceSettings, os_type})
		if err != nil {
			log.Debug("userInfo消息插入错误:", err)
			return
		}
	}
}

func getUserId(userName string) (userInfo UserInfomation, err error) {
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceUserInfo)
	err = c.Find(bson.M{"account": userName}).One(&userInfo)

	return
}

func getUserDeviceTokenMoethod(userName string) (userInfo UserInfomation, err error) {
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceUserInfo)
	err = c.Find(bson.M{"account": userName}).One(&userInfo)

	return
}

/******
   ****
   ****  消息的处理
   ****
******/

func insertMessage(messageCreator string, messageReceiver string, messageBody string, isreaded bool) (err error) {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	err = inserData(session, messageCreator, messageReceiver, messageBody, isreaded)
	if err != nil {
		log.Debug("消息插入错误:", err)
		return
	}

	return

}

func findAllUnreadMessage(authBody string) (resultMessageArr []MessageMongo, err error) {
	// fmt.Println("getOffLineMessage = ", authBody)
	//1. 获得传过来的userid
	userName := gjson.Get(authBody, "UserName").String()
	// fmt.Println("userName = ", userName)

	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	resultMessageArr, err = findAllUnreadData(session, userName)
	if err != nil {
		log.Debug("所有消息查询错误:", err)
		// log.Fatal(err)
	}

	return
}

func updateMessageReadSate(getMsgSuccessInfo string) (err error) {
	messageIdClient := gjson.Get(getMsgSuccessInfo, "MessageIdClient").String()
	messageReceiver := gjson.Get(getMsgSuccessInfo, "MessageReceiver").String()

	// fmt.Println("消息状态更新参数: messageIdClient = ", messageIdClient, "  messageReceiver = ", messageReceiver)

	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		fmt.Println("数据库连接错误:", err)
		return
	}
	defer session.Close()

	err = updateData(session, messageIdClient, messageReceiver)
	if err != nil {
		log.Debug("消息状态更新错误:", err)
		fmt.Println("数据库更新错误:", err)
	}

	return
}

/******
   ****
   ****  mongodb的基本操作
   ****
******/

//*** 连接数据库
func getSession() (session *mgo.Session, err error) {
	//连接参数
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{databaseUrl},
		Direct:    false,
		Timeout:   time.Second * 10.,
		Database:  database,
		Source:    source,
		Username:  username,
		Password:  password,
		PoolLimit: 4096, // Session.SetPoolLimit
	}

	if golalSession == nil {
		golalSession, err = mgo.DialWithInfo(dialInfo)
	}

	if err == nil {
		// session, err = mgo.DialWithInfo(dialInfo)
		session = golalSession.Clone()

		// Optional. Switch the session to a monotonic behavior.
		session.SetMode(mgo.Monotonic, true)
	} else {
		log.Debug("数据库连接错误")
		// log.Fatal(err)
	}
	return
}

//****插入消息体
func inserData(session *mgo.Session, messageCreator string, messageReceiver string, messageBody string, isreaded bool) (err error) {
	c := session.DB(database).C(source)

	//创建消息的时间
	messagegetTime := gjson.Get(messageBody, "MessageSendTime").String()
	MessageIdClient := gjson.Get(messageBody, "MessageIdClient").String()
	MessageIdServer := gjson.Get(messageBody, "MessageIdServer").String()

	// fmt.Println("messageCreator = ", messageCreator, " messageReceiver = ", messageReceiver, " messageBody = ", messageBody, " MessageIdClient = ", MessageIdClient)
	err = c.Insert(&MessageMongo{messageCreator, messageReceiver, messageBody, messagegetTime, MessageIdClient, MessageIdServer, isreaded, ""})
	if err != nil {
		log.Debug("数据插入错误")
		// log.Fatal(err)
	}
	return
}

//***消息置位 未读->已读
func updateData(session *mgo.Session, messageIdClient string, messageReceiver string) (err error) {
	c := session.DB(database).C(source)
	err = c.Update(bson.M{"messageidclient": messageIdClient, "userreceiverchatid": messageReceiver, "isreaded": false}, bson.M{"$set": bson.M{"isreaded": true}})
	return
}

//****查询最新的消息(用于更新消息列表，暂时不用)
func findNewestData(session *mgo.Session, userreceiverchatid string) (resultMessage MessageMongo, err error) {

	c := session.DB(database).C(source)
	err = c.Find(bson.M{"userreceiverchatid": userreceiverchatid}).Sort("-messagegettime").One(&resultMessage)

	return
}

//****查询所有的未读的消息(用于更新消息列表)
func findAllUnreadData(session *mgo.Session, userreceiverchatid string) (resultMessageArr []MessageMongo, err error) {

	c := session.DB(database).C(source)
	err = c.Find(bson.M{"userreceiverchatid": userreceiverchatid, "isreaded": false}).Sort("messagegettime").All(&resultMessageArr)

	return
}

/*
//*** 删除过期的已读的消息(未完成)；overTime 是秒，消息已读了overTime秒后，可以删除
func deleteOverTimeReadedData(session *mgo.Session, overTime int32) (err error) {
	// c := session.DB(database).C(source)
	// err = c.Find(bson.M{"userchatid": userchatid}).Sort("messagegettime").All(&resultMessageArr)

	return
}

*/

/******
   ****
   ****  推送的处理
   ****
******/

//***推送个数 +1
func addNumberOfPushes(messagePushId string) int32 {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return 0
	}
	defer session.Close()

	var resultMessage MessagePush

	c := session.DB(database).C(sourcePush)
	err = c.Find(bson.M{"messagepushid": messagePushId}).One(&resultMessage)

	if err != nil {
		log.Debug("查询错误:", err)
		//****没查到就插入一个数据
		err = c.Insert(&MessagePush{messagePushId, 0, strconv.FormatInt(time.Now().UTC().UnixNano(), 10)})
		return 0
	} else {
		c.Update(bson.M{"messagepushid": messagePushId}, bson.M{"$set": bson.M{"pushnumber": resultMessage.PushNumber + 1, "lastpushedtime": strconv.FormatInt(time.Now().UTC().UnixNano(), 10)}})
	}

	return resultMessage.PushNumber
}

//***推送个数置为0
func resetNumberOfPushes(messagePushId string) {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourcePush)
	c.Update(bson.M{"messagepushid": messagePushId}, bson.M{"$set": bson.M{"pushnumber": 0}})
}

//***判断这个消息要不要推送
func whetherNeedPushThisMessage(messageReceiver string, messageIdServer string) bool {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return true
	}
	defer session.Close()

	var resultMessage MessageMongo

	c := session.DB(database).C(source)
	_ = c.Find(bson.M{"userreceiverchatid": messageReceiver, "messageidserver": messageIdServer}).One(&resultMessage)
	return resultMessage.Isreaded
}


func searchAndroidPushAccount(){


}

/******
   ****
   ****  撤回的处理
   ****
******/

//***添加一条消息撤回数据，用于离线推送
func addRecalledMessageTypeOne(messageCreator, messageReceiver string, messageIdClient string) (err error) {
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceRecall)
	err = c.Insert(&MessageRecalled{messageCreator, messageReceiver, messageIdClient, 0, false})
	if err != nil {
		log.Debug("Recalled消息插入错误:", err)
		return
	}

	err = updateData(session, messageIdClient, messageReceiver)
	if err != nil {
		log.Debug("Recalled消息状态更新错误:", err)
	}

	return
}

//***获取所有离线的撤回消息
func getAllOfflineRecalledMsg(recalledBody string) (resultMessageArr []MessageRecalled, err error) {
	//1. 获得传过来的userid
	messageCreator := gjson.Get(recalledBody, "UserName").String()
	// fmt.Println("userName = ", userName)

	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceRecall)
	err = c.Find(bson.M{"messagereceiver": messageCreator, "isrecalled": false}).All(&resultMessageArr)

	if err != nil {
		log.Debug("所有recalled消息查询错误:", err)
	}

	return
}

//***更新离线消息处理状况
func resetRecalledState(messageReceiver string, messageIdClient string) (err error) {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceRecall)
	err = c.Update(bson.M{"messagecreator": messageReceiver, "messageidclient": messageIdClient, "isrecalled": false}, bson.M{"$set": bson.M{"isrecalled": true}})

	if err != nil {
		log.Debug("UpdateAll错误:", err)
	}
	return
}

//******http 的推送调用
func addSpecialMessage(receiveAccount string, createAccount string,  messageKind string, messageIdClient string, content string) (err error) {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceSpecial)

	err = c.Insert(&SpecialMessage{receiveAccount, createAccount,messageKind, messageIdClient, content, false})
	if err != nil {
		log.Debug("addSpecialMessage 数据插入错误")
	}

	return
}
func getSpecialMessage(receiveAccount string) (resultMessageArr []SpecialMessage, err error) {

	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceSpecial)
	err = c.Find(bson.M{"receiveaccount": receiveAccount, "isreceived": false}).All(&resultMessageArr)

	if err != nil {
		log.Debug("所有recalled消息查询错误:", err)
	}

	return
}

func resetSpecialState(receiveAccount string, messageIdClient string) (err error) {
	//** 1.连接数据库
	session, err := getSession()
	if err != nil {
		log.Debug("数据库连接错误:", err)
		return
	}
	defer session.Close()

	c := session.DB(database).C(sourceSpecial)
	err = c.Update(bson.M{"receiveaccount": receiveAccount, "messageidclient": messageIdClient, "isreceived": false}, bson.M{"$set": bson.M{"isreceived": true}})

	if err != nil {
		log.Debug("UpdateAll错误:", err)
	}
	return
}
