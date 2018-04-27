package main

import (
	"errors"
	log "github.com/thinkboy/log4go"
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"
	"time"
	//***abc自己的代码
	// "fmt"
)

var (
	logicRpcClient *xrpc.Clients
	logicRpcQuit   = make(chan struct{}, 1)

	logicService                        = "RPC"
	logicServicePing                    = "RPC.Ping"
	logicServiceConnect                 = "RPC.Connect"
	logicServiceDisconnect              = "RPC.Disconnect"
	logicServiceDeliverMessage          = "RPC.DeliverMessage"
	logicServicePushOffLineMessage      = "RPC.PushOffLineMessage"
	logicServiceClientGetSuccessMessage = "RPC.ClientGetSuccessMessage"
	logicServiceResetPushNumber         = "RPC.ClientResetPushNumber"
	logicServiceIsTypeing               = "RPC.ClientIsTypeing"
	logicServiceRecalledOne             = "RPC.ClientRecalledOne"
	logicServiceRecalledOneSuccess      = "RPC.ClientRecalledOneSuccess"
)

func InitLogicRpc(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		rpcOptions    []xrpc.ClientOptions
	)
	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		options := xrpc.ClientOptions{
			Proto: network,
			Addr:  addr,
		}
		rpcOptions = append(rpcOptions, options)
	}
	// rpc clients
	logicRpcClient = xrpc.Dials(rpcOptions)
	// ping & reconnect
	logicRpcClient.Ping(logicServicePing)
	log.Info("init logic rpc: %v", rpcOptions)
	return
}

func connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	var (
		arg   = proto.ConnArg{Token: string(p.Body), Server: Conf.ServerId}
		reply = proto.ConnReply{}
	)

	if err = logicRpcClient.Call(logicServiceConnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = 5 * 60 * time.Second
	return
}

func disconnect(key string, roomId int32) (has bool, err error) {
	var (
		arg   = proto.DisconnArg{Key: key, RoomId: roomId}
		reply = proto.DisconnReply{}
	)
	if err = logicRpcClient.Call(logicServiceDisconnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	has = reply.Has
	return
}

//***消息的处理
func processMessage(p *proto.Proto) (err error) {
	var (
		arg   = proto.DeliverMessageArg{Message: string(p.Body)}
		reply = proto.DeliverMessageReply{}
	)

	if err = logicRpcClient.Call(logicServiceDeliverMessage, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceDeliverMessage, arg, err)
		return
	}

	hasError := reply.HasError
	if hasError {
		err = errors.New(reply.ErrorString)
	} else {
		err = nil
	}
	return
}

//***拉取离线消息
func pushAllOffLineMessages(p *proto.Proto) (err error) {
	var (
		arg   = proto.PushOfflineMessageArg{PushInfo: string(p.Body)}
		reply = proto.PushOfflineMessageReply{}
	)
	if err = logicRpcClient.Call(logicServicePushOffLineMessage, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServicePushOffLineMessage, arg, err)
		return
	}
	hasError := reply.HasError
	if hasError {
		err = errors.New(reply.ErrorString)
	} else {
		err = nil
	}
	return
}

//*** 消息接收成功
func clientGetMessagesReply(p *proto.Proto) (err error) {
	var (
		arg   = proto.ClientGetSuccessMessageArg{GetMsgSuccessInfo: string(p.Body)}
		reply = proto.ClientGetSuccessMessageArgReply{}
	)
	if err = logicRpcClient.Call(logicServiceClientGetSuccessMessage, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceClientGetSuccessMessage, arg, err)
		return
	}
	hasError := reply.HasError

	if hasError {
		err = errors.New(reply.ErrorString)
	} else {
		err = nil
	}
	return
}

//*** 撤回消息
func recalledOneMessage(p *proto.Proto) (err error) {
	var (
		arg   = proto.ClientRecalledOneArg{RecalledOneInfo: string(p.Body)}
		reply = proto.ClientRecalledOneArgReply{}
	)
	if err = logicRpcClient.Call(logicServiceRecalledOne, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceRecalledOne, arg, err)
		return
	}

	if reply.ErrorString == "" {
		err = nil
	} else {
		err = errors.New(reply.ErrorString)
	}

	return
}

func recalledOneMessageSuccess(p *proto.Proto) (err error) {
	var (
		arg   = proto.ClientRecalledOneSuccessArg{RecalledOneSuccessInfo: string(p.Body)}
		reply = proto.ClientRecalledOneSuccessArgReply{}
	)
	if err = logicRpcClient.Call(logicServiceRecalledOneSuccess, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceRecalledOneSuccess, arg, err)
		return
	}

	if reply.ErrorString == "" {
		err = nil
	} else {
		err = errors.New(reply.ErrorString)
	}

	return
}

//*** 正在输入消息
func isTypeingNoti(p *proto.Proto) (err error) {

	var (
		arg   = proto.ClientIsTypeingArg{IsTypeingInfo: string(p.Body)}
		reply = proto.ClientIsTypeingArgReply{}
	)
	if err = logicRpcClient.Call(logicServiceIsTypeing, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceIsTypeing, arg, err)
		return
	}

	if reply.ErrorString == "" {
		err = nil
	} else {
		err = errors.New(reply.ErrorString)
	}

	return
}

//*** 更新小红点
func resetPushNumber(p *proto.Proto) (err error) {
	var (
		arg   = proto.ClientResetPushNumberArg{ResetPushNumberInfo: string(p.Body)}
		reply = proto.ClientResetPushNumberArgReply{}
	)
	if err = logicRpcClient.Call(logicServiceResetPushNumber, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceResetPushNumber, arg, err)
		return
	}

	if reply.ErrorString == "" {
		err = nil
	} else {
		err = errors.New(reply.ErrorString)
	}

	return

}
