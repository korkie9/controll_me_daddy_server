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

func sendButtonEvent(joystick uinput.Gamepad, key int, state int) error {
	if state == 1 || state == -1 {
		return joystick.ButtonDown(key)
	} else {
		return joystick.ButtonUp(key)
	}
}

func sendHatEvent(joystick uinput.Gamepad, key int, state int) error {
	if key == 16 {
		if state == 1 {
			fmt.Printf("Logging left presss")
			return joystick.HatPress(uinput.ButtonDpadLeft)

		}
		if state == -1 {
			fmt.Printf("Logging right presss")
			return joystick.HatPress(uinput.ButtonDpadRight)
		}
		if state == 0 {
			joystick.HatRelease(uinput.ButtonDpadLeft)
			return joystick.HatRelease(uinput.ButtonDpadLeft)
		}
	}
	if key == 17 {
		if state == 1 {
			return joystick.HatPress(uinput.ButtonDpadUp)
		}
		if state == -1 {
			return joystick.HatPress(uinput.ButtonDpadDown)
		}
		if state == 0 {
			joystick.HatRelease(uinput.ButtonDpadUp)
			return joystick.HatRelease(uinput.ButtonDpadDown)
		}
	}
	return joystick.HatRelease(uinput.HatUp)
}

func wsHandler(w http.ResponseWriter, r *http.Request, joystick uinput.Gamepad) {

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

		var coordMsg models.CoordinateMessage
		if err := json.Unmarshal(message, &coordMsg); err == nil {
			if coordMsg.Side != "" {
				if coordMsg.Side == "left" {
					joystick.LeftStickMove(float32(coordMsg.X), float32(coordMsg.Y))
				} else {
					joystick.RightStickMove(float32(coordMsg.X), float32(coordMsg.Y))
				}
				fmt.Printf("Received coordinates: X=%.2f, Y=%.2f, Side=%s\n", coordMsg.X, coordMsg.Y, coordMsg.Side)
				continue
			}
		}

		var btnMsg models.ButtonMessage
		if err := json.Unmarshal(message, &btnMsg); err == nil {
			fmt.Printf("Received button: Number=%d, On=%v\n", btnMsg.Key, btnMsg.Value)
			if btnMsg.Key == 16 || btnMsg.Key == 17 {
				sendHatEvent(joystick, btnMsg.Key, btnMsg.Value)
				continue
			}
			if btnMsg.Value == 1 {
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

	joystick, err := uinput.CreateGamepad("/dev/uinput", []byte("Virtual Joystick"), 0x1234, 0x5678)
	if err != nil {
		log.Fatalf("Failed to create virtual joystick: %v", err)
	}
	defer joystick.Close()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, joystick) // Pass as interface value
	})
	fmt.Println("WebSocket server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
