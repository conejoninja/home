package logger

import (
	"fmt"
	"os"

	"encoding/json"

	"log"
	"net/http"

	"time"

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	"github.com/conejoninja/home/telegram"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

var db storage.Storage
var c mqtt.Client
var cfg common.HomeConfig

// MQTT
var subscriptions map[string]bool
var token mqtt.Token

// WEBSOCKETS
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Start is the entrypoint for the logger
func Start(homecfg common.HomeConfig, dbcon storage.Storage, mqttclient mqtt.Client) {
	cfg = homecfg
	db = dbcon
	c = mqttclient

	subscriptions = make(map[string]bool)

	restartDevices()

	// Discover new devices when they connect to the network
	if token = c.Subscribe("discovery", 0, discoveryHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token = c.Subscribe("events", 0, eventsHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if cfg.WS.Enabled {
		http.HandleFunc("/ws", handleConnections)

		go handleMessages()

		go func() {
			for {
				go echo("WebSocket server started on: " + cfg.WS.Port)
				err := http.ListenAndServe(":"+cfg.WS.Port, nil)
				if err != nil {
					log.Fatal("ListenAndServe: ", err)
					fmt.Println("(╯°□°)╯ API server failed, restarting in...")
				}
				time.Sleep(5 * time.Second)
			}
		}()

	}
	fmt.Println("END")
}

func restartDevices() {
	devices := db.GetDevices()
	for _, device := range devices {
		subscriptions[device.ID] = true
		go echo("Subscribed to " + device.ID)
		if token = c.Subscribe(device.ID, 0, defaultHandler); token.WaitTimeout(10*time.Second) && token.Error() != nil {
			subscriptions[device.ID] = false
			go echo(fmt.Sprintln(token.Error()))
			os.Exit(1)

		}
	}
}

var discoveryHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go echo("[" + msg.Topic() + "] " + string(msg.Payload()))
	var device common.Device
	err := json.Unmarshal(msg.Payload(), &device)
	if err == nil {
		db.AddDevice([]byte(device.ID), device)
		if v, ok := subscriptions[device.ID]; !ok || !v {
			subscriptions[device.ID] = true
			if token = c.Subscribe(device.ID, 0, defaultHandler); token.Wait() && token.Error() != nil {
				subscriptions[device.ID] = false
				fmt.Println(token.Error())
				os.Exit(1)
			}

		}
	} else {
		fmt.Println(err)
	}
}

var eventsHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go echo("[" + msg.Topic() + "] " + string(msg.Payload()))
	var evt common.Event
	err := json.Unmarshal(msg.Payload(), &evt)
	if err == nil {
		db.AddEvent(evt.ID, evt)
		telegram.NotifyEvent(evt)
	} else {
		fmt.Println(err)
	}
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go echo("[" + msg.Topic() + "] " + string(msg.Payload()))
	var values []common.Value
	err := json.Unmarshal(msg.Payload(), &values)
	if err == nil {
		for _, value := range values {
			db.AddValue(msg.Topic(), value)
			datetime := time.Now()
			if value.Time != nil && !(*value.Time).IsZero() {
				datetime = *value.Time
			}
			CalculateMetaAll(msg.Topic()+"-"+value.ID, datetime)
		}
	} else {
		fmt.Println(err)
	}
}

/**
 * WEBSOCKETS
 */
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	clients[ws] = true

}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func echo(s string) {
	t := time.Now()
	s = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d ",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second()) + s
	fmt.Println(s)
	broadcast <- []byte(s)
}
