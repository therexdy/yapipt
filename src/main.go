package main

import (
	"net/http"
	"yapipt/internal"
	"yapipt/pkg"
)

func main(){
	R, err := internal.InitRuntime("env")
	if err != nil {
		pkg.LogError("Error Loading env")
		return
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", R.InitWSConn)

	pkg.LogInfo("Started at :" + R.TCPServePort)
	err = http.ListenAndServe(":"+R.TCPServePort, mux)
	if err != nil{
		pkg.LogError("http server failed to start at /:" + R.TCPServePort)
	}
}
