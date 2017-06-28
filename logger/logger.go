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

func Start() {
	db_path, proto, server, port, user, password, client_id, ws_enabled, ws_port := readConfig()

	db = storage.NewBadger(db_path)

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
		fmt.Println(token.Error())
		panic(token.Error())
	}

	restartDevices()

	go echo("Starting " + client_id + " ...")
	// Discover new devices when they connect to the network
	if token = c.Subscribe("discovery", 0, discoveryHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token = c.Subscribe("events", 0, eventsHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if ws_enabled {
		// Configure websocket route
		http.HandleFunc("/ws", handleConnections)

		// Start listening for incoming chat messages
		go handleMessages()

		// Start the server on localhost port 8000 and log any errors
		go echo("http server started on: " + ws_port)
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
		go echo("Subscribed to " + device.Id)
		token = c.Subscribe(device.Id, 0, nil)
		if token = c.Subscribe(device.Id, 0, nil); token.WaitTimeout(10*time.Second) && token.Error() != nil {
			subscriptions[device.Id] = false
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

var eventsHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	go echo("[" + msg.Topic() + "] " + string(msg.Payload()))
	var evt common.Event
	err := json.Unmarshal(msg.Payload(), &evt)
	if err == nil {
		db.AddEvent(evt.Id, evt)
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
			if !value.Time.IsZero() {
				datetime = value.Time
			}
			CalculateMetaAll(msg.Topic()+"-"+value.Id, datetime)
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

func readConfig() (db_path, proto, server, port, user, password, client_id string, ws_enabled bool, ws_port string) {
	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	db_path = os.Getenv("DB_PATH")

	proto = os.Getenv("MQTT_PROTOCOL")
	server = os.Getenv("MQTT_SERVER")
	port = os.Getenv("MQTT_PORT")
	user = os.Getenv("MQTT_USER")
	password = os.Getenv("MQTT_PASSWORD")
	client_id = os.Getenv("MQTT_CLIENT_ID")

	ws_enabled_str := os.Getenv("WEBSOCKET_ENABLED")
	ws_port = os.Getenv("WEBSOCKET_PORT")

	if proto == "" {
		proto = fmt.Sprint(viper.Get("mqtt_protocol"))
	}
	if db_path == "" {
		db_path = fmt.Sprint(viper.Get("db_path"))
	}
	if server == "" {
		server = fmt.Sprint(viper.Get("mqtt_server"))
	}
	if port == "" {
		port = fmt.Sprint(viper.Get("mqtt_port"))
	}
	if user == "" {
		user = fmt.Sprint(viper.Get("mqtt_user"))
	}
	if password == "" {
		password = fmt.Sprint(viper.Get("mqtt_password"))
	}
	if client_id == "" {
		client_id = fmt.Sprint(viper.Get("mqtt_client_id"))
	}
	if ws_enabled_str == "" {
		ws_enabled_str = fmt.Sprint(viper.Get("websocket_enabled"))
	}
	if ws_port == "" {
		ws_port = fmt.Sprint(viper.Get("websocket_port"))
	}

	if db_path == "" {
		db_path = "./db"
	}
	if proto == "" {
		proto = "ws"
	}
	if port == "" {
		port = "9001"
	}
	if client_id == "" {
		client_id = "logger"
	}

	ws_enabled = false
	if ws_enabled_str == "1" || ws_enabled_str == "true" {
		ws_enabled = true
	}
	if ws_port == "" {
		ws_port = "8055"
	}

	return

}
