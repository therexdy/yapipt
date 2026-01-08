package pkg

import (
	"fmt"
	"time"
)

const timeLayout string = "2006-01-02 15:04:05"

func LogInfo(s string){
	fmt.Println("[I] - " + time.Now().Format(timeLayout) + " " + s)
}

func LogWarn(s string){
	fmt.Println("\t[W] - " + time.Now().Format(timeLayout) + " " + s)
}

func LogError(s string){
	fmt.Println("\n\t[E] - " + time.Now().Format(timeLayout) + " " + s + "\n")
}

func LogClientError(s string){
	fmt.Println("[E] - " + time.Now().Format(timeLayout) + " " + s)
}
