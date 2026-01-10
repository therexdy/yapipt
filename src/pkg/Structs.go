package pkg

const (
	MsgData="msg_data"
	Control="control"
	MsgIndct="msg_indct"
)

const (
	Joined="joined"
	Typing="typing"
	StoppedTyping="stopped_typing"
	Left="left"
)

type Envelop struct {
	Type string `json:"type"`
}

type MsgDataJSON struct{
	Type string `json:"type"`
	User string `json:"user"`
	Msg string `json:"msg"`
	SentTime string `json:"sent_time"`
}


type MsgIndctJSON struct{
	Type string `json:"type"`
	IndctType string `json:"indct_type"`
	User string `json:"user"`
}

