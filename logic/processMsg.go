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
	MessageCreator  string //消息创建者
	MessageReceiver string //消息接受者
	MessageIdClient string //消息id，客户端生成
	MessageIdServer string //消息id，服务器生成
	MessageSendTime string //消息发送时间,从服务器获取，用于展示
	Body_en         string //消息主体
	Key             string //消息解密Key
	IsSendSuccess   bool   //是否被客户端接收成功
	MessageTag      bool   //消息是否是我发的
}


func processMessage(messageBody string, messageSendTime string) (err error) {
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	//1. 获得消息的  创建者  和 接收者
	messageCreator := gjson.Get(messageBody, "MessageCreator").String()
	messageReceiver := gjson.Get(messageBody, "MessageReceiver").String()

	//2. 修改需要结构体
	var mR MessageReceive
	err = json.Unmarshal([]byte(messageBody), &mR)
	if err != nil {
		log.Debug("string -> json 结构不对：", messageBody, " - ", err)
		fmt.Println("string -> json 结构不对：", messageBody, " - ", err)
		return
	}
	mR.IsSendSuccess = true
	mR.MessageTag = true
	mR.MessageSendTime = messageSendTime
	mR.MessageIdServer = "qxProject" + messageCreator + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)

	bodyBytes, err = json.Marshal(mR)
	if err != nil {
		log.Debug("json -> string消息结构不对：", messageBody, " - ", err)
		fmt.Println("string -> json 结构不对：", messageBody, " - ", err)
		return
	}
	messageBody = string(bodyBytes)

	//3. 根据  messageReceiver 判断当前用户是否存在于连接池里
	messsgeReceiveId, err := changeUserNameToUserId(messageReceiver)
	if err != nil {
		log.Debug("changeUserNameToUserId err：", err)
		return
	}
	subKeyDic := genSubKey(messsgeReceiveId)

	if len(subKeyDic) > 0 {
		fmt.Println("有客户端在线->", subKeyDic)
		log.Debug("有客户端在线: ", subKeyDic)
		//2. 缓存消息体 (isreaded = true)
		go func() {
			// /1. 先进行消息的推送
			for serverId, keys = range subKeyDic {
				log.Debug("serverId = ", serverId)
				if len(keys) > 0{
					if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
						return
					}
				}
			}
			if err = insertMessage(messageCreator, messageReceiver, messageBody, false); err != nil {
				log.Debug("消息插入失败:", err)
				fmt.Println("消息插入失败:", err)
				return
			}

		}()

	} else {
		fmt.Println("无客户端在线")
		//1. 缓存消息体 (isreaded = false)
		go func() {
			if err = insertMessage(messageCreator, messageReceiver, messageBody, false); err != nil {
				log.Debug("消息插入失败:", err)
				return
			}
		}()
	}

	//****推送消息
	ProcessGeneralMethod(messageReceiver, mR.MessageIdServer, 1)

	return
}

func getOffLineMessage(authBody string) (err error) {
	var (
		serverId          int32
		keys              []string
		bodyBytes         []byte
		recalledBodyBytes []byte
		specialBodyBytes  []byte
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
					bodyBytes = []byte(string(value.MessageBody))
					// fmt.Println("subKeyDic：", subKeyDic)
					// 进行消息的推送
					if len(subKeyDic) > 0 {
						for serverId, keys = range subKeyDic {
							// fmt.Println("serverId = ", serverId, "keys = ", keys)
							if len(keys) > 0{
								if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
									return
								}
							}
						}
					}
				}
			}
		} else {
			bodyBytes = []byte("{\"OnceMsg\":1, \"OP\":21}")
			if len(subKeyDic) > 0 {
				for serverId, keys = range subKeyDic {
					fmt.Println("serverId = ", serverId , " keys = ", keys)
					if len(keys) > 0{
						if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
							return
						}
					}
				}
			}
		}

	}()

	//****推送所有的撤回消息
	go func() {
		time.AfterFunc(2.*time.Second, func() {
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
								if len(keys) > 0{
									if err = mpushKafka(serverId, keys, recalledBodyBytes); err != nil {
										return
									}
								}
							}
						}

					}
				}
			}
		})
	}()

	//****推送所有的spacial消息
	go func() {
		time.AfterFunc(1.*time.Second, func() {
			subKeyDic3 := genSubKey(messsgeReceiveId)
			//1. 查询所有的离线消息
			resultSpecialArr := []SpecialMessage{}
			resultSpecialArr, err = getSpecialMessage(messageReceiver)
			// fmt.Println("所有离线special消息: ", resultSpecialArr)
			if len(resultSpecialArr) > 0 {
				for _, value := range resultSpecialArr {
					if err == nil {
						specialBodyBytes = []byte("{\"OnceMsg\":1, \"OP\":23, \"MessageCreator\": \"" + string(value.ReceiveAccount) + "\", \"MessageKind\":\"" + string(value.MessageKind) + "\", \"MessageIdClient\":\"" + string(value.MessageIdClient) + "\", \"Content\":\"" + string(value.Content) +"\", \"MessageReceive\":\"" + string(value.CreateAccount) + "\"}")
						if len(subKeyDic3) > 0 {
							for serverId, keys = range subKeyDic3 {
								if len(keys) > 0{
									if err = mpushKafka(serverId, keys, specialBodyBytes); err != nil {
										return
									}
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
	go func() {
		err = updateMessageReadSate(getMsgSuccessInfo)
	}()
	// err = updateMessageReadSate(getMsgSuccessInfo)
	return
}

func isTypeingMethod(IsTypeingInfo string) (err error) {
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	WhoIsTyping := gjson.Get(IsTypeingInfo, "WhoIsTyping").String()
	TypingMsgReceiver := gjson.Get(IsTypeingInfo, "TypingMsgReceiver").String()

	messsgeReceiveId, err := changeUserNameToUserId(TypingMsgReceiver)

	if err != nil {
		return
	}

	subKeyDic := genSubKey(messsgeReceiveId)
	if len(subKeyDic) > 0 {
		bodyBytes = []byte("{\"OnceMsg\":1 ,\"OP\":20, \"WhoIsTyping\":\"" + WhoIsTyping + "\"}")
		for serverId, keys = range subKeyDic {
			if len(keys) > 0{
				if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
					return
				}
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
			if len(keys) > 0{
				if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
					return
				}
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
	log.Debug("MessageCreator = ", MessageCreator, "MessageIdClient = ", MessageIdClient)

	resetRecalledState(MessageCreator, MessageIdClient)

}

//******http 的推送调用
func sendHttpSpacialImMessage(receiveAccount string, createAccount string, messageKind string, content string) {
	// fmt.Println("receiveAccount = ", receiveAccount)
	// fmt.Println("messageKind = ", messageKind)
	messageIdClient := "qxProject" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	//添加一条离线的recalled消息
	go func() {
		addSpecialMessage(receiveAccount, createAccount, messageKind, messageIdClient, content)
	}()

	go func() {
		time.AfterFunc(2.*time.Second, func() {
			sendSpacialMessage(receiveAccount, createAccount, messageKind, messageIdClient, content)
		})
	}()
}

func sendSpacialMessage(receiveAccount string, createAccount string, messageKind string, messageIdClient string, content string) {
	var (
		serverId  int32
		keys      []string
		bodyBytes []byte
	)

	messsgeReceiveId, err := changeUserNameToUserId(receiveAccount)

	if err != nil {
		return
	}

	//***推送spacial消息
	subKeyDic := genSubKey(messsgeReceiveId)
	if len(subKeyDic) > 0 {
		bodyBytes = []byte("{\"OnceMsg\":1, \"OP\":23, \"MessageCreator\": \"" + receiveAccount + "\", \"MessageKind\":\"" + messageKind + "\", \"MessageIdClient\":\"" + messageIdClient + "\", \"Content\":\"" + content + "\", \"MessageReceive\":\"" + createAccount +"\"}")
		for serverId, keys = range subKeyDic {
			if len(keys) > 0{
				if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
					return
				}
			}
		}
	}
}

func resetSpacialMessage(msgResetInfo string) {
	// fmt.Println("recalledOneSuccessMwthod 执行了", RecalledOneInfo)
	MessageCreator := gjson.Get(msgResetInfo, "MessageCreator").String()
	MessageIdClient := gjson.Get(msgResetInfo, "MessageIdClient").String()

	// fmt.Println("MessageCreator =", MessageCreator, "MessageIdClient = ", MessageIdClient)
	resetSpecialState(MessageCreator, MessageIdClient)
}
