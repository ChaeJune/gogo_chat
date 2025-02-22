package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

// 세션 스토어
var Store = sessions.NewCookieStore([]byte("super-secret-key"))

// WebSocket 업그레이더
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// OpenAI API 관련 구조체
type OpenAIRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type OpenAIChoice struct {
	Text string `json:"text"`
}

type OpenAIResponse struct {
	Choices []OpenAIChoice `json:"choices"`
}

// 최대 허용 요청 수 (예: 세션당 100회)
const maxRequestsPerSession = 25

// callOpenAI 함수는 사용자 프롬프트를 받아 OpenAI API를 호출합니다.
func callOpenAI(prompt string) (string, error) {
	url := "https://api.openai.com/v1/completions"
	reqBody := OpenAIRequest{
		Model:       "text-davinci-003",
		Prompt:      prompt,
		Temperature: 0.5,
		MaxTokens:   150,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer YOUR_OPENAI_API_KEY") // 실제 API 키로 교체

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp OpenAIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return "", err
	}

	if len(apiResp.Choices) > 0 {
		return apiResp.Choices[0].Text, nil
	}
	return "", fmt.Errorf("OpenAI API로부터 유효한 응답이 없습니다")
}

// WSHandler는 세션 인증 후 WebSocket 연결을 업그레이드하고,
// 요청 제한을 체크한 후 OpenAI API 호출 결과를 클라이언트에 전달합니다.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	// 세션 확인
	session, err := Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "세션을 불러올 수 없습니다", http.StatusInternalServerError)
		return
	}
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "인증되지 않았습니다", http.StatusUnauthorized)
		return
	}

	// WebSocket 업그레이드
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 에러:", err)
		return
	}
	defer conn.Close()

	// 세션에서 사용량 카운트 초기화 (없으면 0으로 시작)
	usage, ok := session.Values["usageCount"].(int)
	if !ok {
		usage = 0
	}

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("메시지 읽기 에러:", err)
			break
		}
		log.Printf("수신 메시지: %s", msg)

		// 사용량 체크: 제한 초과 시 호출 거부
		if usage >= maxRequestsPerSession {
			errMsg := "요청 제한에 도달했습니다."
			conn.WriteMessage(msgType, []byte(errMsg))
			continue
		}

		// 요청 처리: OpenAI API 호출
		response, err := callOpenAI(string(msg))
		if err != nil {
			log.Println("OpenAI 호출 에러:", err)
			conn.WriteMessage(msgType, []byte("오류가 발생했습니다."))
			continue
		}

		// 응답 전송
		if err := conn.WriteMessage(msgType, []byte(response)); err != nil {
			log.Println("메시지 전송 에러:", err)
			break
		}

		// 사용량 증가 후 세션에 저장
		usage++
		session.Values["usageCount"] = usage
		if err := session.Save(r, w); err != nil {
			log.Println("세션 저장 에러:", err)
			break
		}
	}
}
