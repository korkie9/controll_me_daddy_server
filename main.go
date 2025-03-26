package main

import (
	"controll-me-daddy/models"
	"encoding/json"
	"fmt"
	"github.com/bendahl/uinput"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func sendHatEvent(joystick uinput.Gamepad, key int, state int) error {
	if key == 16 {
		if state == -1 {
			return joystick.HatPress(uinput.HatLeft)
		}
		if state == 1 {
			return joystick.HatPress(uinput.HatRight)
		}
		if state == 0 {
			joystick.HatRelease(uinput.HatRight)
			return joystick.HatRelease(uinput.HatLeft)
		}
	}
	if key == 17 {
		if state == 1 {
			return joystick.HatPress(uinput.HatUp)
		}
		if state == -1 {
			return joystick.HatPress(uinput.HatDown)
		}
		if state == 0 {
			joystick.HatRelease(uinput.HatUp)
			return joystick.HatRelease(uinput.HatDown)
		}
	}
	return nil
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

			// Detailed logging for D-pad events
			if btnMsg.Key == 16 || btnMsg.Key == 17 {
				fmt.Printf("D-Pad Event: Key=%d, Value=%d\n", btnMsg.Key, btnMsg.Value)
				err := sendHatEvent(joystick, btnMsg.Key, btnMsg.Value)
				if err != nil {
					fmt.Printf("Error in D-Pad event: %v\n", err)
				}
				continue
			}

			// Regular button handling
			if btnMsg.Value == 1 {
				joystick.ButtonDown(btnMsg.Key)
			} else {
				joystick.ButtonUp(btnMsg.Key)
			}
			continue
		}

		fmt.Printf("Received: %s\n", message)
	}
}

func main() {
	joystick, err := uinput.CreateGamepad("/dev/uinput", []byte("Detailed Virtual Joystick"), 0x045E, 0x028E)
	if err != nil {
		log.Fatalf("Failed to create virtual joystick: %v", err)
	}
	defer joystick.Close()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, joystick)
	})

	fmt.Println("WebSocket server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
