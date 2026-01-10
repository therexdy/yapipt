package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"yapipt/pkg"

	"github.com/gorilla/websocket"
)


type ClientConn struct {
	user string
	WSConn *websocket.Conn
	WSConnMutex sync.Mutex
	CloseReaderRoutine bool
}

func loadEnv(ENV_VAR string) (string, error){
	env_var := os.Getenv(ENV_VAR)
	if(env_var==""){
		return "", errors.New(ENV_VAR + " not in env")
	}
	return env_var, nil
}

func (R *Runtime)saveEnv() error {
	var err error

	R.TCPServePort, err = loadEnv("SERVER_TCP_PORT")
	if err != nil {
		return err
	}
	pkg.LogInfo("SERVER_TCP_PORT="+R.TCPServePort)
	return nil
}

type Runtime struct{
	TCPServePort string
	HubMutex sync.Mutex
	WSConnHub map[string]*ClientConn
	BroadcastChan chan []byte
}

func (R *Runtime)BroadcastMsgData(raw_bytes []byte) {
	var msg_json pkg.MsgDataJSON
	err := json.Unmarshal(raw_bytes, &msg_json)
	if err!= nil {
		pkg.LogClientError("Unmarshal Error for message_data")
		return 
	}
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.WSConn.WriteJSON(msg_json)
		CC.WSConnMutex.Unlock()
	}
}

func (R *Runtime)BroadcastMsgIndct(raw_bytes []byte) {
	var msg_json pkg.MsgIndctJSON
	err := json.Unmarshal(raw_bytes, &msg_json)
	if err!= nil {
		pkg.LogClientError("Unmarshal Error for message_data")
		return 
	}
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.WSConn.WriteJSON(msg_json)
		CC.WSConnMutex.Unlock()
	}
}

func InitRuntime(env_file string) (*Runtime, error) {
	var R Runtime

	err := R.saveEnv()
	if err != nil {
		return &R, err
	}

	R.WSConnHub = make(map[string]*ClientConn)

	R.BroadcastChan = make(chan []byte)

	go func(R *Runtime) {
		var raw_bytes []byte
		for{
			raw_bytes = <- R.BroadcastChan
			if string(raw_bytes)=="" {
				continue
			} else if string(raw_bytes)=="Close" {
				break
			}
			var envlp pkg.Envelop
			err = json.Unmarshal(raw_bytes, &envlp)
			if err!=nil{
				pkg.LogClientError("Unmarshal Error for rawBytes from client")
			}
			switch envlp.Type {
			case pkg.MsgData:
				R.BroadcastMsgData(raw_bytes)
			case pkg.MsgIndct:
				R.BroadcastMsgIndct(raw_bytes)
			default:
				pkg.LogWarn("Marshal unknown JSON format")
			}
		}
		pkg.LogInfo("Broadcast GoRoutine Closed")
	}(&R)

	return &R, nil
}

func (R *Runtime) DeInitRuntime() {
	R.BroadcastChan <- []byte("Close")
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.CloseReaderRoutine = true
		CC.WSConnMutex.Unlock()
	}
}

