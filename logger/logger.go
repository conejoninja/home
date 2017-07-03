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

type loggerconfig struct {
	db_path, proto, server, port, user, password, client_id, ws_port string
	ws_enabled                                                       bool
}

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
	cfg := readConfig()

	db = storage.NewBadger(cfg.db_path)

	opts := mqtt.NewClientOptions().AddBroker(cfg.proto + "://" + cfg.server + ":" + cfg.port)
	opts.SetClientID(cfg.client_id)
	if cfg.user != "" {
		opts.SetUsername(cfg.user)
	}
	if cfg.password != "" {
		opts.SetPassword(cfg.password)
	}
	opts.SetDefaultPublishHandler(defaultHandler)

	subscriptions = make(map[string]bool)

	c = mqtt.NewClient(opts)
	if token = c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}

	restartDevices()

	go echo("Starting " + cfg.client_id + " ...")
	// Discover new devices when they connect to the network
	if token = c.Subscribe("discovery", 0, discoveryHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if token = c.Subscribe("events", 0, eventsHandler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	if cfg.ws_enabled {
		// Configure websocket route
		http.HandleFunc("/ws", handleConnections)

		// Start listening for incoming chat messages
		go handleMessages()

		// Start the server on localhost port 8000 and log any errors
		go echo("http server started on: " + cfg.ws_port)
		err := http.ListenAndServe(":"+cfg.ws_port, nil)
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
			if value.Time!=nil && !(*value.Time).IsZero() {
				datetime = *value.Time
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

func readConfig() (cfg loggerconfig) {
	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	cfg.db_path = os.Getenv("DB_PATH")

	cfg.proto = os.Getenv("MQTT_PROTOCOL")
	cfg.server = os.Getenv("MQTT_SERVER")
	cfg.port = os.Getenv("MQTT_PORT")
	cfg.user = os.Getenv("MQTT_USER")
	cfg.password = os.Getenv("MQTT_PASSWORD")
	cfg.client_id = os.Getenv("MQTT_CLIENT_ID")

	ws_enabled_str := os.Getenv("WEBSOCKET_ENABLED")
	cfg.ws_port = os.Getenv("WEBSOCKET_PORT")

	if cfg.proto == "" {
		cfg.proto = fmt.Sprint(viper.Get("mqtt_protocol"))
	}
	if cfg.db_path == "" {
		cfg.db_path = fmt.Sprint(viper.Get("db_path"))
	}
	if cfg.server == "" {
		cfg.server = fmt.Sprint(viper.Get("mqtt_server"))
	}
	if cfg.port == "" {
		cfg.port = fmt.Sprint(viper.Get("mqtt_port"))
	}
	if cfg.user == "" {
		cfg.user = fmt.Sprint(viper.Get("mqtt_user"))
	}
	if cfg.password == "" {
		cfg.password = fmt.Sprint(viper.Get("mqtt_password"))
	}
	if cfg.client_id == "" {
		cfg.client_id = fmt.Sprint(viper.Get("mqtt_client_id"))
	}
	if ws_enabled_str == "" {
		ws_enabled_str = fmt.Sprint(viper.Get("websocket_enabled"))
	}
	if cfg.ws_port == "" {
		cfg.ws_port = fmt.Sprint(viper.Get("websocket_port"))
	}

	if cfg.db_path == "" {
		cfg.db_path = "./db"
	}
	if cfg.proto == "" {
		cfg.proto = "ws"
	}
	if cfg.port == "" {
		cfg.port = "9001"
	}
	if cfg.client_id == "" {
		cfg.client_id = "logger"
	}

	cfg.ws_enabled = false
	if ws_enabled_str == "1" || ws_enabled_str == "true" {
		cfg.ws_enabled = true
	}
	if cfg.ws_port == "" {
		cfg.ws_port = "8055"
	}
	

	return

}

