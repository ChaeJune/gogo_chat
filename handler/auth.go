package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

// 세션 스토어를 공개 변수로 만들어 다른 핸들러에서도 재사용 가능하게 함
var Store = sessions.NewCookieStore([]byte("super-secret-key"))

// db는 PostgreSQL 데이터베이스 연결 객체입니다.
var db *sql.DB

// init 함수에서 데이터베이스 연결을 초기화합니다.
func init() {
	var err error
	// 연결 문자열은 환경에 맞게 수정
	connStr := "user=postgres password=yourpassword dbname=yourdbname sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("데이터베이스 연결 실패:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("데이터베이스 ping 실패:", err)
	}
}

// LoginHandler는 POST 방식으로 전달된 username과 password를 사용해 인증 후 세션에 인증 값을 저장합니다.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// POST 방식만 허용
	if r.Method != http.MethodPost {
		http.Error(w, "POST 방식만 허용됩니다", http.StatusMethodNotAllowed)
		return
	}

	// 폼 데이터 파싱
	if err := r.ParseForm(); err != nil {
		http.Error(w, "폼 파싱 오류", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "아이디와 비밀번호가 필요합니다", http.StatusBadRequest)
		return
	}

	// 데이터베이스에서 username에 해당하는 해시된 비밀번호 조회
	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)
	if err != nil {
		http.Error(w, "유효하지 않은 자격 증명입니다", http.StatusUnauthorized)
		return
	}

	// 입력받은 비밀번호와 데이터베이스에 저장된 해시 비교 (bcrypt 사용)
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		http.Error(w, "유효하지 않은 자격 증명입니다", http.StatusUnauthorized)
		return
	}

	// 인증 성공 시 세션에 값 저장
	session, err := Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "세션 오류", http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = true
	session.Values["username"] = username
	if err := session.Save(r, w); err != nil {
		http.Error(w, "세션 저장 실패", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "로그인 성공!")
}
