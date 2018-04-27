package main

import (
	"encoding/json"
	"fmt"
	log "github.com/thinkboy/log4go"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
)

type MessageReceive struct {
	Content         string //内容
	MessageCreator  string //消息创建者
	MessageReceiver string //消息接受者
	MessageIdClient string //消息id，客户端生成
	MessageIdServer string //消息id，服务器生成
	MsgType         int64  //消息类型
	FileUrl         string //附件url
	MessageSendTime string //消息发送时间
	IsSendSuccess   bool   //是否被客户端接收成功
	MessageTag      bool   //消息是否是我发的
}

func processMessage(messageBody string) (err error) {
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	//1. 获得消息的  创建者  和 接收者
	messageCreator := gjson.Get(messageBody, "MessageCreator").String()
	messageReceiver := gjson.Get(messageBody, "MessageReceiver").String()

	// log.Warn("messageBody = ", messageBody)

	//2. 根据  messageReceiver 判断当前用户是否存在于连接池里
	messsgeReceiveId, err := changeUserNameToUserId(messageReceiver)
	if err != nil {
		log.Warn("changeUserNameToUserId err：", err)
		return
	}

	subKeyDic := genSubKey(messsgeReceiveId)

	//3. 修改需要 推送／保存 的消息体
	// fmt.Println(">>>", messageBody)
	var mR MessageReceive
	err = json.Unmarshal([]byte(messageBody), &mR)
	if err != nil {
		log.Warn("string -> json 结构不对：", messageBody, " - ", err)
		return
	}
	mR.IsSendSuccess = true
	mR.MessageTag = true
	mR.MessageIdServer = "qxProject" + messageCreator + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	bodyBytes, err = json.Marshal(mR)
	if err != nil {
		log.Warn("json -> string消息结构不对：", messageBody, " - ", err)
		return
	}
	messageBody = string(bodyBytes)
	// fmt.Println("messageBody结果 = ", messageBody)
	// bodyBytes = []byte(messageBody)

	if len(subKeyDic) > 0 {
		// fmt.Println("有客户端在线->", subKeyDic)
		//2. 缓存消息体 (isreaded = true)
		go func() {
			// /1. 先进行消息的推送
			for serverId, keys = range subKeyDic {
				fmt.Println("serverId = ", serverId)
				if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
					return
				}
			}
			if err = insertMessage(messageCreator, messageReceiver, messageBody, false); err != nil {
				log.Warn("消息插入失败:", err)
				return
			}

		}()

	} else {
		// fmt.Println("无客户端在线")
		//1. 缓存消息体 (isreaded = false)
		go func() {
			if err = insertMessage(messageCreator, messageReceiver, messageBody, false); err != nil {
				log.Warn("消息插入失败:", err)
				return
			}
		}()
	}

	//****推送消息
	ProcessAfterDelay(messageReceiver, mR.MessageIdServer)

	return
}

func getOffLineMessage(authBody string) (err error) {
	var (
		serverId          int32
		keys              []string
		bodyBytes         []byte
		recalledBodyBytes []byte
	)

	// 1. 离线消息推送
	messageReceiver := gjson.Get(authBody, "UserName").String()
	messsgeReceiveId, err := changeUserNameToUserId(messageReceiver)

	if err != nil {
		return
	}

	//***推送所有的离线消息
	go func() {
		subKeyDic := genSubKey(messsgeReceiveId)
		//1. 查询所有的离线消息
		resultMessageArr := []MessageMongo{}
		resultMessageArr, err = findAllUnreadMessage(authBody)
		// fmt.Println("离线查询结果：", resultMessageArr)
		if len(resultMessageArr) > 0 {
			for _, value := range resultMessageArr {
				// jsons, errs := json.Marshal(value)
				if err == nil {
					// fmt.Println("查询结果：", string(value.MessageBody))
					bodyBytes = []byte(string(value.MessageBody))
					// fmt.Println("查询结果：", bodyBytes)
					// 进行消息的推送
					if len(subKeyDic) > 0 {
						for serverId, keys = range subKeyDic {
							if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
								return
							}
						}
					}
				}
			}
		}

	}()
	//****推送所有的撤回消息
	go func() {
		time.AfterFunc(3.*time.Second, func() {
			subKeyDic2 := genSubKey(messsgeReceiveId)
			//1. 查询所有的离线消息
			resultRecalledArr := []MessageRecalled{}
			resultRecalledArr, err = getAllOfflineRecalledMsg(authBody)
			// fmt.Println("所有离线recalled消息: ", resultRecalledArr)
			if len(resultRecalledArr) > 0 {
				for _, value := range resultRecalledArr {
					if err == nil {
						recalledBodyBytes = []byte("{\"OnceMsg\":1, \"OP\":18, \"MessageCreator\": \"" + string(value.MessageCreator) + "\", \"MessageIdClient\":\"" + string(value.MessageIdClient) + "\"}")
						if len(subKeyDic2) > 0 {
							for serverId, keys = range subKeyDic2 {
								if err = mpushKafka(serverId, keys, recalledBodyBytes); err != nil {
									return
								}
							}
						}

					}
				}
			}
		})
	}()

	return
}

func clientGetMsgSuccess(getMsgSuccessInfo string) (err error) {
	// fmt.Println("收到参数是：", getMsgSuccessInfo)
	go func() {
		err = updateMessageReadSate(getMsgSuccessInfo)
	}()
	return
}

func isTypeingMethod(IsTypeingInfo string) (err error) {
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	WhoIsTypeing := gjson.Get(IsTypeingInfo, "WhoIsTypeing").String()
	TypeingMsgReceiver := gjson.Get(IsTypeingInfo, "TypeingMsgReceiver").String()

	messsgeReceiveId, err := changeUserNameToUserId(TypeingMsgReceiver)

	if err != nil {
		return
	}

	subKeyDic := genSubKey(messsgeReceiveId)
	if len(subKeyDic) > 0 {
		bodyBytes = []byte("{\"OnceMsg\":1 ,\"OP\":20, \"WhoIsTypeing\":\"" + WhoIsTypeing + "\"}")
		for serverId, keys = range subKeyDic {
			if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
				return
			}
		}

	}

	return
}

func recalledOneMethod(RecalledOneInfo string) (err error) {
	fmt.Println("recalledOneMethod 执行了", RecalledOneInfo)
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	messageCreator := gjson.Get(RecalledOneInfo, "MessageCreator").String()
	messageReceiver := gjson.Get(RecalledOneInfo, "MessageReceiver").String()
	messageIdClient := gjson.Get(RecalledOneInfo, "MessageIdClient").String()

	messsgeReceiveId, err := changeUserNameToUserId(messageReceiver)

	if err != nil {
		return
	}

	//***推送recalled消息
	subKeyDic := genSubKey(messsgeReceiveId)
	if len(subKeyDic) > 0 {
		bodyBytes = []byte("{\"OnceMsg\":1, \"OP\":18, \"MessageCreator\": \"" + messageCreator + "\", \"MessageIdClient\":\"" + messageIdClient + "\"}")
		for serverId, keys = range subKeyDic {
			if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
				return
			}
		}
	}

	//添加一条离线的recalled消息
	go func() {
		addRecalledMessageTypeOne(messageCreator, messageReceiver, messageIdClient)
	}()

	return
}
func recalledOneSuccessMethod(RecalledOneInfo string) {
	// fmt.Println("recalledOneSuccessMwthod 执行了", RecalledOneInfo)
	MessageCreator := gjson.Get(RecalledOneInfo, "MessageCreator").String()
	MessageIdClient := gjson.Get(RecalledOneInfo, "MessageIdClient").String()
	fmt.Println("MessageCreator = ", MessageCreator, "MessageIdClient = ", MessageIdClient)

	resetRecalledState(MessageCreator, MessageIdClient)

}
