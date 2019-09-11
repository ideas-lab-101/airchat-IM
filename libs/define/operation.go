package define

const (
	// handshake
	OP_HANDSHAKE       = int32(0)
	OP_HANDSHAKE_REPLY = int32(1)
	// heartbeat
	OP_HEARTBEAT       = int32(2)
	OP_HEARTBEAT_REPLY = int32(3)
	// send text messgae
	OP_SEND_SMS               = int32(4)
	OP_SEND_SMS_REPLY         = int32(5)
	OP_SEND_SMS_SUCCESS       = int32(15)
	OP_OFFLINE_SMS            = int32(16)
	OP_CLIENT_SMS_GET_SUCCESS = int32(17)

	OP_CLIENT_SMS_RECALLED         = int32(18)
	OP_CLIENT_SMS_RECALLED_ALL     = int32(19)
	OP_CLIENT_ISTYPING             = int32(20)
	OP_CLIENT_RESETPUSHNUMBER      = int32(21)
	OP_CLIENT_RECALLED_ONE_SUCCESS = int32(22)
	OP_CLIENT_RECESPECIAL_MSG      = int32(23)

	// kick user
	OP_DISCONNECT_REPLY = int32(6)

	// auth user
	OP_AUTH       = int32(7)
	OP_AUTH_REPLY = int32(8)

	// handshake with sid
	OP_HANDSHAKE_SID       = int32(9)
	OP_HANDSHAKE_SID_REPLY = int32(10)

	// raw message
	OP_RAW = int32(11)
	// room
	OP_ROOM_READY = int32(12)
	// proto
	OP_PROTO_READY  = int32(13)
	OP_PROTO_FINISH = int32(14)

	// for test
	OP_TEST       = int32(254)
	OP_TEST_REPLY = int32(255)
)
