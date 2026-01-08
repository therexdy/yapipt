package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"yapipt/pkg"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)


type ClientConn struct {
	user string
	WSConn *websocket.Conn
	CloseReaderRoutine bool
}

func loadEnv(ENV_VAR string) (string, error){
	env_var := os.Getenv(ENV_VAR)
	if(env_var==""){
		return "", errors.New(ENV_VAR + " not in .env")
	}
	return env_var, nil
}

func (R *Runtime)saveEnv(env_file string) error {
	godotenv.Load(env_file)
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
	WSProtoUpgrader websocket.Upgrader
	HubMutex sync.Mutex
	WSConnHub map[string]*ClientConn
	BroadcastChan chan []byte
}

func InitRuntime(env_file string) (*Runtime, error) {
	var R Runtime

	err := R.saveEnv(env_file)
	if err != nil {
		return &R, err
	}

	R.WSProtoUpgrader = websocket.Upgrader{}
	R.WSConnHub = make(map[string]*ClientConn)

	R.BroadcastChan = make(chan []byte)

	go func(R *Runtime) {
		var message_data []byte
		for{
			message_data = <- R.BroadcastChan
			if string(message_data)=="" {
				continue
			} else if string(message_data)=="Close" {
				break
			}
			var msg_json pkg.MsgFrmClntJSON
			err = json.Unmarshal(message_data, &msg_json)
			if err!= nil {
				pkg.LogError("Unmarshal Error for message_data")
				return 
			}
			for _, CC := range R.WSConnHub {
				CC.WSConn.WriteJSON(msg_json)
			}
		}
	}(&R)

	return &R, nil
}
