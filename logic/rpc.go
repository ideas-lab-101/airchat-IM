package main

import (
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/thinkboy/log4go"
	//***abc自己的代码
	// "fmt"
)

func InitRPC(auther Auther) (err error) {
	var (
		network, addr string
		c             = &RPC{auther: auther}
	)
	rpc.Register(c)
	for i := 0; i < len(Conf.RPCAddrs); i++ {
		log.Info("start listen rpc addr: \"%s\"", Conf.RPCAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.RPCAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go rpcListen(network, addr)
	}
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", addr)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// RPC
type RPC struct {
	auther Auther
}

func (r *RPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

// Connect auth and registe login
func (r *RPC) Connect(arg *proto.ConnArg, reply *proto.ConnReply) (err error) {
	if arg == nil {
		err = ErrConnectArgs
		log.Error("Connect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)

	uid, reply.RoomId, err = r.auther.Auth(arg.Token)

	// fmt.Println("鉴权后的结果>", uid, reply.RoomId, "|", "arg = ", arg, "err = ", err)

	if err == nil {
		// fmt.Println("session 新建")
		if seq, err = connect(uid, arg.Server, reply.RoomId); err == nil {
			reply.Key = encode(uid, seq)
		}

	}

	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(arg *proto.DisconnArg, reply *proto.DisconnReply) (err error) {
	if arg == nil {
		err = ErrDisconnectArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = decode(arg.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", arg.Key, err)
		return
	}
	reply.Has, err = disconnect(uid, seq, arg.RoomId)
	return
}

func (r *RPC) DeliverMessage(arg *proto.DeliverMessageArg, reply *proto.DeliverMessageReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("DeliverMessageArg() error(%v)", err)
		return
	}

	err = processMessage(string(arg.Message), string(arg.MessageSendTime))
	if err != nil {
		reply.HasError = true
	}

	return
}

func (r *RPC) PushOffLineMessage(arg *proto.PushOfflineMessageArg, reply *proto.PushOfflineMessageReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("PushOffLineMessage() error(%v)", err)
		return
	}

	err = getOffLineMessage(string(arg.PushInfo))
	if err != nil {
		reply.HasError = true
	}

	return
}

func (r *RPC) ClientGetSuccessMessage(arg *proto.ClientGetSuccessMessageArg, reply *proto.ClientGetSuccessMessageArgReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("ClientGetSuccessMessage() error(%v)", err)
		return
	}
	err = clientGetMsgSuccess(string(arg.GetMsgSuccessInfo))
	if err != nil {
		reply.HasError = true
	}
	return
}

func (r *RPC) ClientResetPushNumber(arg *proto.ClientResetPushNumberArg, reply *proto.ClientResetPushNumberArgReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("logicServiceResetPushNumber() error(%v)", err)
		return
	}

	ResetPushNumber(string(arg.ResetPushNumberInfo))

	//***忽略错误处理

	return
}

func (r *RPC) ClientRecalledOne(arg *proto.ClientRecalledOneArg, reply *proto.ClientRecalledOneArgReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("logicServiceRecalledOne() error(%v)", err)
		return
	}

	recalledOneMethod(string(arg.RecalledOneInfo))

	//***忽略错误处理

	return
}

func (r *RPC) ClientRecalledOneSuccess(arg *proto.ClientRecalledOneSuccessArg, reply *proto.ClientRecalledOneSuccessArgReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("ClientRecalledOneSuccess() error(%v)", err)
		return
	}

	recalledOneSuccessMethod(string(arg.RecalledOneSuccessInfo))

	//***忽略错误处理

	return
}

func (r *RPC) ClientIsTypeing(arg *proto.ClientIsTypeingArg, reply *proto.ClientIsTypeingArgReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("logicServiceIsTypeing() error(%v)", err)
		return
	}

	isTypeingMethod(string(arg.IsTypeingInfo))

	//***忽略错误处理

	return
}

func (r *RPC) HttpSpacialMsgReset(arg *proto.HttpSpacialMsgResetArgs, reply *proto.HttpSpacialMsgResetReply) (err error) {
	if arg == nil {
		err = ErrDeliverMessageArg
		log.Error("logicServiceIsTypeing() error(%v)", err)
		return
	}
	resetSpacialMessage(string(arg.MsgResetInfo))
	return
}
