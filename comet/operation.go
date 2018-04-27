package main

import (
	log "github.com/thinkboy/log4go"
	"github.com/tidwall/gjson"
	"goim/libs/define"
	"goim/libs/proto"
	"time"
	//***abc自己的代码
	// "fmt"
)

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*proto.Proto) error
	PushOffLineMessage(*proto.Proto) error
	ClientGetMessageSuucss(*proto.Proto) error
	// Connect used for auth user and return a subkey, roomid, hearbeat.
	Connect(*proto.Proto) (string, int32, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string, int32) error
}

type DefaultOperator struct {
}

func (operator *DefaultOperator) Operate(p *proto.Proto) (err error) {
	var (
		body []byte
	)
	if p.Operation == define.OP_SEND_SMS {
		//1. 处理收到的消息push + cache
		var pProcess *proto.Proto = new(proto.Proto)
		pProcess.Operation = p.Operation
		pProcess.Body = p.Body
		pProcess.SeqId = p.SeqId
		pProcess.Ver = p.Ver

		//2. 通知客户端消息发送成功
		p.Operation = define.OP_SEND_SMS_SUCCESS
		messageIdClient := gjson.Get(string(p.Body), "MessageIdClient").String()
		p.Body = []byte("{\"Code\":1, \"MessageIdClient\": \"" + messageIdClient + "\"}")
		log.Info("send sms proto: %v", p.String())

		go func() {
			err = processMessage(pProcess)
		}()

	} else if p.Operation == define.OP_CLIENT_SMS_RECALLED {
		//***撤回一个消息
		var pProcess *proto.Proto = new(proto.Proto)
		pProcess.Operation = p.Operation
		pProcess.Body = p.Body
		pProcess.SeqId = p.SeqId
		pProcess.Ver = p.Ver

		//2. 通知客户端消息发送成
		p.Operation = define.OP_SEND_SMS_SUCCESS
		p.Body = []byte("{\"Code\":1}")

		go func() {
			recalledOneMessage(pProcess)
		}()

	} else if p.Operation == define.OP_CLIENT_RECALLED_ONE_SUCCESS {
		//***撤回一个消息成功
		recalledOneMessageSuccess(p)

	} else if p.Operation == define.OP_CLIENT_SMS_RECALLED_ALL {
		//***撤回所有消息

	} else if p.Operation == define.OP_CLIENT_ISTYPING {
		//***正在打印
		isTypeingNoti(p)

	} else if p.Operation == define.OP_CLIENT_RESETPUSHNUMBER {
		//***重置小红点
		resetPushNumber(p)

	} else if p.Operation == define.OP_TEST {
		log.Debug("test operation: %s", body)
		p.Operation = define.OP_TEST_REPLY
		p.Body = []byte("{\"Test\":\"come on\"}")
	} else {
		return ErrOperation
	}
	return nil
}

func (operator *DefaultOperator) PushOffLineMessage(p *proto.Proto) (err error) {
	err = pushAllOffLineMessages(p)
	return
}

func (operator *DefaultOperator) ClientGetMessageSuucss(p *proto.Proto) (err error) {
	err = clientGetMessagesReply(p)
	return
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string, rid int32) (err error) {
	var has bool
	if has, err = disconnect(key, rid); err != nil {
		return
	}
	if !has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}
