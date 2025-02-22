package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/username/my-chat-app/handlers"
)

func main() {
	// 로그인 및 WebSocket 핸들러를 등록합니다.
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/ws", handlers.WSHandler)

	fmt.Println("서버 실행 중 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
