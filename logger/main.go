package main

import (
	"fmt"
	"os"

	"encoding/json"

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
)

var db storage.Storage
var subscriptions map[string]bool
var token mqtt.Token
var c mqtt.Client

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

	for {

	}

	defer c.Disconnect(250)

}

func restartDevices() {
	devices := db.GetDevices()
	for _, device := range devices {
		subscriptions[device.Id] = true
		fmt.Println("Subscribed to", device.Id)
		if token = c.Subscribe(device.Id, 0, nil); token.Wait() && token.Error() != nil {
			subscriptions[device.Id] = false
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}
}

var discoveryHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("DISCOVERY: %s\n", msg.Payload())
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
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
	var value common.Value
	err := json.Unmarshal(msg.Payload(), &value)
	if err == nil {
		db.AddValue(msg.Topic(), value)
	} else {
		fmt.Println(err)
	}
}
