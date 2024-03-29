package proto

type ConnArg struct {
	Token  string
	Server int32
}

type ConnReply struct {
	Key    string
	RoomId int32
}

type DisconnArg struct {
	Key    string
	RoomId int32
}

type DisconnReply struct {
	Has bool
}

type DeliverMessageArg struct {
	Message         string
	MessageSendTime string
}

type DeliverMessageReply struct {
	HasError    bool
	ErrorString string
}

type PushOfflineMessageArg struct {
	PushInfo string
}

type PushOfflineMessageReply struct {
	HasError    bool
	ErrorString string
}

type ClientGetSuccessMessageArg struct {
	GetMsgSuccessInfo string
}

type ClientGetSuccessMessageArgReply struct {
	HasError    bool
	ErrorString string
}

type ClientResetPushNumberArg struct {
	ResetPushNumberInfo string
}

type ClientResetPushNumberArgReply struct {
	ErrorString string
}

type ClientIsTypeingArg struct {
	IsTypeingInfo string
}

type ClientIsTypeingArgReply struct {
	ErrorString string
}

type ClientRecalledOneArg struct {
	RecalledOneInfo string
}

type ClientRecalledOneArgReply struct {
	ErrorString string
}

type ClientRecalledOneSuccessArg struct {
	RecalledOneSuccessInfo string
}

type ClientRecalledOneSuccessArgReply struct {
	ErrorString string
}

//****这个是http的私信调用
type HttpSpacialImMessageArgs struct {
	ReceiveAccount string
	MessageKind    string
}

type HttpSpacialImMessageReply struct {
	ErrorString string
}

type HttpSpacialMsgResetArgs struct {
	MsgResetInfo string
}

type HttpSpacialMsgResetReply struct {
	ErrorString string
}
