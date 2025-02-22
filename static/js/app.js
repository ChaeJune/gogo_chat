document.addEventListener('DOMContentLoaded', () => {
    // WebSocket 서버 주소 (환경에 맞게 수정)
    const socket = new WebSocket('ws://localhost:8080/ws');
  
    // DOM 요소 참조
    const messagesContainer = document.getElementById('messages');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');
  
    // WebSocket 연결 성공 이벤트
    socket.addEventListener('open', () => {
      console.log('WebSocket 연결 성공');
    });
  
    // 서버로부터 메시지 수신 이벤트
    socket.addEventListener('message', (event) => {
      console.log('서버로부터 메시지 수신:', event.data);
      appendMessage(event.data, 'server-message');
    });
  
    // WebSocket 오류 처리
    socket.addEventListener('error', (error) => {
      console.error('WebSocket 오류:', error);
    });
  
    // WebSocket 연결 종료 이벤트
    socket.addEventListener('close', () => {
      console.log('WebSocket 연결 종료');
    });
  
    // 메시지 전송 함수
    function sendMessage() {
      const message = messageInput.value.trim();
      if (message && socket.readyState === WebSocket.OPEN) {
        // 사용자의 메시지를 화면에 출력
        appendMessage(message, 'user-message');
        // 서버로 메시지 전송
        socket.send(message);
        // 입력창 초기화
        messageInput.value = '';
      }
    }
  
    // 메시지를 채팅 화면에 추가하는 함수
    function appendMessage(text, type) {
      const messageElem = document.createElement('div');
      messageElem.classList.add('message', type);
      messageElem.innerHTML = `<p>${text}</p>`;
      messagesContainer.appendChild(messageElem);
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  
    // 전송 버튼 클릭 이벤트
    sendButton.addEventListener('click', sendMessage);
  
    // Enter 키 입력 시 메시지 전송 이벤트
    messageInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        sendMessage();
      }
    });
  });
  