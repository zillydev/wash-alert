package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (use cautiously)
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex = sync.Mutex{}

type Machine struct {
	IsMachine     bool
	MachineNumber int
	Status        string
	TimeBooked    *time.Time
	HoursBooked   int
	MinutesBooked int
	Timer         string
}

type Floor struct {
	FloorNumber int
	FloorName   string
	Rows        int
	Columns     int
	Grid        [][]Machine
}

type BookingData struct {
	FloorNumber  int
	RowNumber    int
	ColumnNumber int
	Hours        int
	Minutes      int
	TimeBooked   string
	Token        string
}

type SendData struct {
	Type    string
	Message any
}

type ReceiveData struct {
	Type    string
	Message any
}

type LocationData struct {
	FloorNumber  int
	RowNumber    int
	ColumnNumber int
}

type PushSubscription struct {
	Endpoint string `json:"endpoint"`
}

type SubscriptionData struct {
	Token        string
	Subscription PushSubscription
}

const (
	// Your VAPID keys generated previously
	vapidPublicKey  = "g4fE9vQ0uuRDr_VgljWa-C-B90SglTFcxDc4cta7Wc8"
	vapidPrivateKey = "BGmy17rlHYEqnUtcECwxCt67SFjirG1g-3kYgthDaYWBCbOqZ2eQ2t_jADUsTgZpw60LGvetKhcrQoqua3HQb4k"
)

var floors = []Floor{
	{
		FloorName: "5th Floor",
		Rows:      2,
		Columns:   5,
		Grid: [][]Machine{
			{Machine{true, 1, "Empty", nil, 0, 0, ""}, Machine{true, 2, "Empty", nil, 0, 0, ""}, Machine{true, 3, "Empty", nil, 0, 0, ""}, Machine{true, 4, "Not Functional", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}},
			{Machine{false, 1, "", nil, 0, 0, ""}, Machine{true, 5, "Empty", nil, 0, 0, ""}, Machine{true, 6, "Empty", nil, 0, 0, ""}, Machine{true, 7, "Empty", nil, 0, 0, ""}, Machine{true, 8, "Empty", nil, 0, 0, ""}},
		},
	},
	{
		FloorName: "Ground Floor",
		Rows:      3,
		Columns:   4,
		Grid: [][]Machine{
			{Machine{true, 1, "Empty", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{true, 2, "Empty", nil, 0, 0, ""}},
			{Machine{true, 3, "Empty", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{true, 4, "Empty", nil, 0, 0, ""}},
			{Machine{true, 5, "Empty", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{false, 1, "", nil, 0, 0, ""}, Machine{true, 6, "Empty", nil, 0, 0, ""}},
		},
	},
}

func send(messageType string, message any) {
	data := SendData{Type: messageType, Message: message}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err = client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			fmt.Println("Error sending message:", err)

			// Remove the client from the list if there's an error
			client.Close()
			delete(clients, client)
			return
		}
	}
}

var subscriptions = make(map[string]string)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	send("data", floors)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)

			// Remove the client from the list
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()

			break
		}
		println("Received:", string(data))
		var response ReceiveData
		err = json.Unmarshal(data, &response)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		if response.Type == "subscribe" {
			subscription, err := json.Marshal(response.Message)
			if err != nil {
				fmt.Println("Error marshalling subscription data:", err)
				return
			}

			var subscriptionData SubscriptionData
			err = json.Unmarshal(subscription, &subscriptionData)
			if err != nil {
				fmt.Println("Error unmarshaling subscription data:", err)
				return
			}

			subscriptions[subscriptionData.Token] = subscriptionData.Subscription.Endpoint
		} else if response.Type == "book" {
			messageData, err := json.Marshal(response.Message)
			if err != nil {
				fmt.Println("Error marshalling message data:", err)
				return
			}

			var booking BookingData
			err = json.Unmarshal(messageData, &booking)
			if err != nil {
				fmt.Println("Error unmarshaling booking data:", err)
				return
			}

			// Process the received booking
			floors[booking.FloorNumber].Grid[booking.RowNumber][booking.ColumnNumber].Status = "Booked"
			floors[booking.FloorNumber].Grid[booking.RowNumber][booking.ColumnNumber].HoursBooked = booking.Hours
			floors[booking.FloorNumber].Grid[booking.RowNumber][booking.ColumnNumber].MinutesBooked = booking.Minutes
			timeBooked, err := time.Parse(time.RFC3339, booking.TimeBooked)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				return
			}
			floors[booking.FloorNumber].Grid[booking.RowNumber][booking.ColumnNumber].TimeBooked = &timeBooked

			send("booked", BookingData{
				FloorNumber:  booking.FloorNumber,
				RowNumber:    booking.RowNumber,
				ColumnNumber: booking.ColumnNumber,
				Hours:        booking.Hours,
				Minutes:      booking.Minutes,
				TimeBooked:   booking.TimeBooked,
			})

			duration := time.Duration(booking.Hours)*time.Hour + time.Duration(booking.Minutes)*time.Minute
			timer := time.NewTimer(duration)

			go func(floorNumber int, rowNumber int, columnNumber int) {
				<-timer.C
				fmt.Println("Timer expired!")
				floors[floorNumber].Grid[rowNumber][columnNumber].Status = "Done"
				floors[floorNumber].Grid[rowNumber][columnNumber].HoursBooked = 0
				floors[floorNumber].Grid[rowNumber][columnNumber].MinutesBooked = 0
				floors[floorNumber].Grid[rowNumber][columnNumber].TimeBooked = nil

				subscription, ok := subscriptions[booking.Token]
				if !ok {
					fmt.Println("Subscription not found for token:", booking.Token)
					return
				}

				sendPushNotification(subscription, "Timer Completed!")

				send("done", LocationData{
					FloorNumber:  floorNumber,
					RowNumber:    rowNumber,
					ColumnNumber: columnNumber,
				})
			}(booking.FloorNumber, booking.RowNumber, booking.ColumnNumber)
		} else if response.Type == "collect" {
			messageData, err := json.Marshal(response.Message)
			if err != nil {
				fmt.Println("Error marshalling message data:", err)
				return
			}

			var collect LocationData
			err = json.Unmarshal(messageData, &collect)
			if err != nil {
				fmt.Println("Error unmarshaling collect data:", err)
				return
			}

			floors[collect.FloorNumber].Grid[collect.RowNumber][collect.ColumnNumber].Status = "Empty"

			send("collected", LocationData{
				FloorNumber:  collect.FloorNumber,
				RowNumber:    collect.RowNumber,
				ColumnNumber: collect.ColumnNumber,
			})
		}
	}
}

func sendPushNotification(subscription string, payload string) error {
	// Create a push message
	msg := []byte(payload)

	s := &webpush.Subscription{}
	json.Unmarshal([]byte(subscription), s)

	// Send the push notification
	resp, err := webpush.SendNotification(msg, s, &webpush.Options{
		Subscriber:      "abc@localhost",
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		TTL:             60,
	})
	if err != nil {
		return fmt.Errorf("failed to send push notification: %v", err)
	}
	defer resp.Body.Close()

	// Check for success
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Println("Push notification sent successfully")
	} else {
		return fmt.Errorf("received non-success status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("VAPID Public Key:", publicKey)
	// fmt.Println("VAPID Private Key:", privateKey)

	http.HandleFunc("/ws", handleConnection)

	// Start the server
	log.Println("WebSocket server started at ws://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
