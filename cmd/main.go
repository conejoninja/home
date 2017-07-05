package main

import (
	"fmt"
	"os"

	"github.com/conejoninja/home/api"
	"github.com/conejoninja/home/logger"
	"github.com/conejoninja/home/storage"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	"time"
)

type MqttConfig struct {
	Protocol, Server, Port, User, Password, ClientId string
}

type HomeConfig struct {
	DBPath string
	Mqtt   MqttConfig
	WS     logger.WebsocketConfig
	Api    api.ApiConfig
}

// STORAGE
var db storage.Storage

// MQTT
var token mqtt.Token
var mqttclient mqtt.Client

func main() {
	cfg := readConfig()

	db = storage.NewBadger(cfg.DBPath)

	opts := mqtt.NewClientOptions().AddBroker(cfg.Mqtt.Protocol + "://" + cfg.Mqtt.Server + ":" + cfg.Mqtt.Port)
	opts.SetClientID(cfg.Mqtt.ClientId)
	if cfg.Mqtt.User != "" {
		opts.SetUsername(cfg.Mqtt.User)
	}
	if cfg.Mqtt.Password != "" {
		opts.SetPassword(cfg.Mqtt.Password)
	}
	//opts.SetDefaultPublishHandler(defaultHandler)

	mqttclient = mqtt.NewClient(opts)
	if token = mqttclient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}

	if cfg.Api.Enabled {
		api.Start(cfg.Api, db, mqttclient)
	}
	logger.Start(cfg.WS, db, mqttclient)

	for ;; {
		fmt.Println(time.Now(), "Still alive")
		time.Sleep(5 * time.Minute)

	}

}

func readConfig() (cfg HomeConfig) {
	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	/**
	 * DATABASE
	 */
	cfg.DBPath = os.Getenv("DB_PATH")
	if cfg.DBPath == "" {
		cfg.DBPath = fmt.Sprint(viper.Get("db_path"))
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "./db"
	}

	/**
	 * MQTT
	 */
	cfg.Mqtt.Protocol = os.Getenv("MQTT_PROTOCOL")
	cfg.Mqtt.Server = os.Getenv("MQTT_SERVER")
	cfg.Mqtt.Port = os.Getenv("MQTT_PORT")
	cfg.Mqtt.User = os.Getenv("MQTT_USER")
	cfg.Mqtt.Password = os.Getenv("MQTT_PASSWORD")
	cfg.Mqtt.ClientId = os.Getenv("MQTT_CLIENT_ID")

	if cfg.Mqtt.Protocol == "" {
		cfg.Mqtt.Protocol = fmt.Sprint(viper.Get("mqtt_protocol"))
	}
	if cfg.Mqtt.Server == "" {
		cfg.Mqtt.Server = fmt.Sprint(viper.Get("mqtt_server"))
	}
	if cfg.Mqtt.Port == "" {
		cfg.Mqtt.Port = fmt.Sprint(viper.Get("mqtt_port"))
	}
	if cfg.Mqtt.User == "" {
		cfg.Mqtt.User = fmt.Sprint(viper.Get("mqtt_user"))
	}
	if cfg.Mqtt.Password == "" {
		cfg.Mqtt.Password = fmt.Sprint(viper.Get("mqtt_password"))
	}
	if cfg.Mqtt.ClientId == "" {
		cfg.Mqtt.ClientId = fmt.Sprint(viper.Get("mqtt_client_id"))
	}
	if cfg.Mqtt.Protocol == "" {
		cfg.Mqtt.Protocol = "ws"
	}
	if cfg.Mqtt.Port == "" {
		cfg.Mqtt.Port = "9001"
	}
	if cfg.Mqtt.ClientId == "" {
		cfg.Mqtt.ClientId = "home-cmd"
	}

	/**
	 * API
	 */
	api_enabled_str := os.Getenv("API_ENABLED")
	cfg.Api.Port = os.Getenv("API_PORT")
	if api_enabled_str == "" {
		api_enabled_str = fmt.Sprint(viper.Get("api_enabled"))
	}
	if cfg.Api.Port == "" {
		cfg.Api.Port = fmt.Sprint(viper.Get("api_port"))
	}

	cfg.Api.Enabled = false
	if api_enabled_str == "1" || api_enabled_str == "true" {
		cfg.Api.Enabled = true
	}
	if cfg.Api.Port == "" {
		cfg.Api.Port = "80"
	}

	/**
	 * WEBSOCKETS
	 */
	ws_enabled_str := os.Getenv("WEBSOCKET_ENABLED")
	cfg.WS.Port = os.Getenv("WEBSOCKET_PORT")
	if ws_enabled_str == "" {
		ws_enabled_str = fmt.Sprint(viper.Get("websocket_enabled"))
	}
	if cfg.WS.Port == "" {
		cfg.WS.Port = fmt.Sprint(viper.Get("websocket_port"))
	}

	cfg.WS.Enabled = false
	if ws_enabled_str == "1" || ws_enabled_str == "true" {
		cfg.WS.Enabled = true
	}
	if cfg.WS.Port == "" {
		cfg.WS.Port = "8055"
	}

	return
}
