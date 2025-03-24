package main

import (
	"controll-me-daddy/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/bendahl/uinput"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	joystick, err := uinput.CreateGamepad("/dev/uinput", []byte("Virtual Joystick"), 0x1234, 0x5678)
	if err != nil {
		log.Fatalf("Failed to create virtual joystick: %v", err)
	}
	defer joystick.Close()

	joystick.ButtonDown(304) // Max right

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// Try to parse as CoordinateMessage first
		var coordMsg models.CoordinateMessage
		if err := json.Unmarshal(message, &coordMsg); err == nil {
			if coordMsg.Side == "left" {
				joystick.LeftStickMove(float32(coordMsg.X), float32(coordMsg.Y))
			} else {
				joystick.RightStickMove(float32(coordMsg.X), float32(coordMsg.Y))
			}
			fmt.Printf("Received coordinates: X=%.2f, Y=%.2f, Side=%.2f\n", coordMsg.X, coordMsg.Y, coordMsg.Side)
			continue
		}

		// If not CoordinateMessage, try ButtonMessage
		var btnMsg models.ButtonMessage
		if err := json.Unmarshal(message, &btnMsg); err == nil {
			fmt.Printf("Received button: Number=%d, On=%v\n", btnMsg.Key, btnMsg.Value)
			// Process button press/release
			if btnMsg.Value == 1 {
				joystick.ButtonDown(btnMsg.Key)
			} else if btnMsg.Value == -1 {
				joystick.ButtonDown(btnMsg.Key)
			} else {
				joystick.ButtonUp(btnMsg.Key)
			}
			continue
		}

		fmt.Printf("Received: %s\\n", message)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
