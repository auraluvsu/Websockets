package ws

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
)

func generateAcceptKey(secWebSocketKey string) string {
	GUID := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	hash := sha1.Sum([]byte(secWebSocketKey + GUID))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func readFullPayload(conn net.Conn, payload []byte) error {
	totalRead := 0
	for totalRead < len(payload) {
		n, err := conn.Read(payload[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	return nil
}

func handleWebSocket(conn net.Conn) {
	fmt.Println("Client connected via WebSocket")
	for {
		header := make([]byte, 2)
		_, err := conn.Read(header)
		if err != nil {
			fmt.Println("Connection closed")
			break
		}

		mask := (header[1] >> 7) & 1
		payloadLen := int(header[1] & 0x7F)

		if payloadLen == 126 {
			extended := make([]byte, 2)
			conn.Read(extended)
			payloadLen = int(binary.BigEndian.Uint16(extended))
		} else if payloadLen == 127 {
			extended := make([]byte, 8)
			conn.Read(extended)
			payloadLen = int(binary.BigEndian.Uint64(extended))
		}

		maskKey := make([]byte, 4)
		if mask == 1 {
			conn.Read(maskKey)
		}

		payload := make([]byte, payloadLen)
		err = readFullPayload(conn, payload)
		if err != nil {
			fmt.Println("Error reading payload:", err)
		}

		if mask == 1 {
			for i := 0; i < payloadLen; i++ {
				payload[i] ^= maskKey[i%4]
			}
		}
		message := string(payload)

		fmt.Println("Received:", message)

		sendWebSocketMessage(conn, "Hello WebSocket client!")
	}
	conn.Close()
}

func sendWebSocketMessage(conn net.Conn, message string) {
	payload := []byte(message)
	payloadLen := len(payload)

	frame := []byte{0x81}
	if payloadLen <= 125 {
		frame = append(frame, byte(payloadLen))
	} else if payloadLen <= 65535 {
		frame = append(frame, 126)
		extended := make([]byte, 2)
		binary.BigEndian.PutUint16(extended, uint16(payloadLen))
		frame = append(frame, extended...)
	} else {
		frame = append(frame, 127)
		extended := make([]byte, 8)
		binary.BigEndian.PutUint64(extended, uint64(payloadLen))
		frame = append(frame, extended...)
	}
	frame = append(frame, payload...)
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