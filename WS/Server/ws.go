package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
)

func generateAcceptKey(secWebSocketKey string) string {
	GUID := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	hash := sha1.Sum([]byte(secWebSocketKey + GUID))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func handleWebSocket(conn net.Conn) {
	fmt.Println("Client connected via WebSocket")
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed")
			break
		}

		message := decodeWebSocketFrame(buf[:n])
		fmt.Println("Received:", message)

		sendWebSocketMessage(conn, "Hello WebSocket client!")
	}
	conn.Close()
}

func decodeWebSocketFrame(frame []byte) string {
	payloadLength := frame[1] & 127
	mask := frame[2:6]
	payload := frame[6 : 6+payloadLength]
	decoded := make([]byte, payloadLength)
	for i := 0; i < int(payloadLength); i++ {
		decoded[i] = payload[i] ^ mask[i%4]
	}
	return string(decoded)
}

func sendWebSocketMessage(conn net.Conn, message string) {
	frame := []byte{0x81, byte(len(message))}
	frame = append(frame, []byte(message)...)
	conn.Write(frame)
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Upgrader header required", http.StatusUpgradeRequired)
		return
	}

	secWebSocketKey := r.Header.Get("Sec-WebSocket-Key")
	acceptKey := generateAcceptKey(secWebSocketKey)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Hijack error", http.StatusInternalServerError)
		return
	}

	response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n\r\n", acceptKey)
	conn.Write([]byte(response))
	handleWebSocket(conn)
}

func main() {
	http.HandleFunc("/ws", handleConnection)
	fmt.Println("WebSocket server running on ws://localhost:2808/ws")
	http.ListenAndServe(":2808", nil)
}
