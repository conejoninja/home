package main

import (
	"fmt"
	"os"

	"encoding/json"

	"log"
	"net/http"

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

// STORAGE
var db storage.Storage

// MQTT
var subscriptions map[string]bool
var token mqtt.Token
var c mqtt.Client

// wEBSOCKETS
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
		os.Exit(1)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	proto := fmt.Sprint(viper.Get("mqtt_protocol"))
	server := fmt.Sprint(viper.Get("mqtt_server"))
	port := fmt.Sprint(viper.Get("mqtt_port"))
	user := fmt.Sprint(viper.Get("mqtt_user"))
	password := fmt.Sprint(viper.Get("mqtt_password"))
	client_id := fmt.Sprint(viper.Get("mqtt_client_id"))

	if proto == "" {
		proto = "ws"
	}
	if port == "" {
		port = "9001"
	}
	if client_id == "" {
		client_id = "logger"
	}

	ws_enabled := false
	if fmt.Sprint(viper.Get("websocket_enabled")) == "1" || fmt.Sprint(viper.Get("websocket_enabled")) == "true" {
		ws_enabled = true
	}
	ws_port := fmt.Sprint(viper.Get("websocket_port"))
	if ws_port == "" {
		ws_port = "8055"
	}

	db = storage.NewBadger("./db")

	opts := mqtt.NewClientOptions().AddBroker(proto + "://" + server + ":" + port)
	opts.SetClientID(client_id)
	if user != "" {
		opts.SetUsername(user)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetDefaultPublishHandler(defaultHandler)

	subscriptions = make(map[string]bool)

	c = mqtt.NewClient(opts)
	if token = c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token)
		fmt.Println(token.Error())
		panic(token.Error())
	}

	restartDevices()

	fmt.Println("Starting " + client_id + " ...")
	// Discover new devices when they connect to the network
	if token := c.Subscribe("discovery", 0, discoveryHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if ws_enabled {
		// Configure websocket route
		http.HandleFunc("/ws", handleConnections)

		// Start listening for incoming chat messages
		go handleMessages()

		// Start the server on localhost port 8000 and log any errors
		print("http server started on: " + ws_port)
		err := http.ListenAndServe(":"+ws_port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}

	for {

	}

	defer c.Disconnect(250)

}

func restartDevices() {
	devices := db.GetDevices()
	for _, device := range devices {
		subscriptions[device.Id] = true
		print("Subscribed to " + device.Id)
		if token = c.Subscribe(device.Id, 0, nil); token.Wait() && token.Error() != nil {
			subscriptions[device.Id] = false
			print(fmt.Sprintln(token.Error()))
			os.Exit(1)
		}
	}
}

var discoveryHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	print("[" + msg.Topic() + "] " + string(msg.Payload()))
	var device common.Device
	err := json.Unmarshal(msg.Payload(), &device)
	if err == nil {
		db.AddDevice([]byte(device.Id), device)
		if v, ok := subscriptions[device.Id]; !ok || !v {
			subscriptions[device.Id] = true
			if token = c.Subscribe(device.Id, 0, nil); token.Wait() && token.Error() != nil {
				subscriptions[device.Id] = false
				fmt.Println(token.Error())
				os.Exit(1)
			}

		}
	} else {
		fmt.Println(err)
	}
}

var defaultHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	print("[" + msg.Topic() + "] " + string(msg.Payload()))
	var value common.Value
	err := json.Unmarshal(msg.Payload(), &value)
	if err == nil {
		db.AddValue(msg.Topic(), value)
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

func print(s string) {
	fmt.Println(s)
	broadcast <- []byte(s)
}
