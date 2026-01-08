package pkg

const (
	MsgData="msg_data"
	Control="control"
	MsgIndct="msg_indct"
)

type Envelop struct {
	Type string `json:"type"`
}

type MsgFrmClntJSON struct{
	Type string `json:"type"`
	User string `json:"user"`
	Msg string `json:"msg"`
	SentTime string `json:"sent_time"`
}

