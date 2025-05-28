package main

import (
	"controll-me-daddy/models"
	"controll-me-daddy/util"
	"encoding/json"
	"fmt"
	"github.com/bendahl/uinput"
	"github.com/gorilla/websocket"
	"github.com/mdp/qrterminal/v3"
	"log"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func sendHatEvent(joystick uinput.Gamepad, key int, state int) error {
	if key == 16 {
		if state == 1 {
			return joystick.HatPress(uinput.HatLeft)
		}
		if state == -1 {
			return joystick.HatPress(uinput.HatRight)
		}
		if state == 0 {
			joystick.HatRelease(uinput.HatRight)
			return joystick.HatRelease(uinput.HatLeft)
		}
	}
	if key == 17 {
		if state == -1 {
			return joystick.HatPress(uinput.HatUp)
		}
		if state == 1 {
			return joystick.HatPress(uinput.HatDown)
		}
		if state == 0 {
			joystick.HatRelease(uinput.HatUp)
			return joystick.HatPress(uinput.HatDown)
		}
	}
	return nil
}

func sendMenu(joystick uinput.Gamepad, key int, state int) error {
	fmt.Printf("received menu input: X=%s, Y=%s", key, state)
	if key == 315 {
		if state == 1 {
			fmt.Printf("Pressign pause here 1")
			return joystick.ButtonDown(uinput.ButtonStart)

		}
		if state == 0 {

			fmt.Printf("Releasing pause here 1")
			return joystick.ButtonUp(uinput.ButtonStart)
		}
	}
	if key == 314 {
		if state == 1 {
			return joystick.ButtonUp(uinput.ButtonSelect)
		}
		if state == 0 {
			return joystick.ButtonUp(uinput.ButtonSelect)
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

			if btnMsg.Key == 16 || btnMsg.Key == 17 {
				err := sendHatEvent(joystick, btnMsg.Key, btnMsg.Value)
				if err != nil {
					fmt.Printf("Error in D-Pad event: %v\n", err)
				}
				continue
			}
			if btnMsg.Key == 314 || btnMsg.Key == 315 {
				err := sendMenu(joystick, btnMsg.Key, btnMsg.Value)
				if err != nil {
					fmt.Printf("Error in Menu event: %v\n", err)
				}
				continue
			}

			// Regular button handling
			if btnMsg.Value == 1 {
				fmt.Printf("Logging button press event event: %v\n", btnMsg.Value)
				joystick.ButtonDown(btnMsg.Key)
			} else {
				fmt.Printf("Logging button release event event: %v\n", btnMsg.Value)
				joystick.ButtonUp(btnMsg.Key)
			}
			continue
		}

		fmt.Printf("Received: %s\n", message)
	}
}

func main() {
	ip, err := util.GetLocalIP()
	if err != nil {
		log.Fatalln("No IP", err)
	}
	websocketUrl := "ws://" + ip + ":8080/ws"

	config := qrterminal.Config{
		HalfBlocks: true,
		Level:      qrterminal.M,
		Writer:     os.Stdout,
	}
	qrterminal.GenerateWithConfig(websocketUrl, config)

	joystick, err := uinput.CreateGamepad("/dev/uinput", []byte("Socket Joystick"), 0x045E, 0x028E)
	if err != nil {
		log.Fatalf("Failed to create virtual joystick: %v", err)
	}
	defer joystick.Close()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, joystick)
	})

	fmt.Println(websocketUrl)
	fmt.Println("WebSocket server started on :8080. Scan QR from Control Me Daddy Mobile app to connect")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
