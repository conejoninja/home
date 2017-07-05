package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"log"
	"time"

	"strconv"
	"strings"

	"errors"

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

type apiconfig struct {
	db_path, web_user, web_password, web_port, mqtt_proto, mqtt_server, mqtt_port, mqtt_user, mqtt_password, mqtt_client_id string
}

// MQTT
var mqtt_client mqtt.Client
var token mqtt.Token

// STORAGE
var db storage.Storage

type sensorResponse map[string]map[string][]common.Value
type metaResponse map[string]common.Meta

func sensor(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {

	ids := strings.Split(ps.ByName("ids"), ";")
	period := ps.ByName("period")
	response := make(sensorResponse)

	for _, id := range ids {
		if _, ok := response[id]; !ok {
			response[id] = make(map[string][]common.Value)
		}
		start, end := getPeriod(period, 0)
		response[id]["current"] = db.GetValuesBetweenTime(id, start, end)
		start, end = getPeriod(period, -1)
		response[id]["past"] = db.GetValuesBetweenTime(id, start, end)
	}

	valStr, _ := json.Marshal(response)
	fmt.Fprint(res, string(valStr))
}

func meta(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {

	ids := strings.Split(ps.ByName("ids"), ";")
	response := make(metaResponse)

	start := time.Now()
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	for _, id := range ids {
		response[id] = db.GetMeta([]byte(id + "-day-" + strconv.Itoa(int(start.Unix()))))
	}

	valStr, _ := json.Marshal(response)
	fmt.Fprint(res, string(valStr))

}

func devices(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	devices := db.GetDevices()
	devsjson, err := json.Marshal(devices)
	if err != nil {
		fmt.Fprint(res, "{\"error\":\"failed\"}")
		return
	}

	fmt.Fprint(res, string(devsjson))
}

func event(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	countStr := ps.ByName("count")
	count := 10
	if countStr != "" {
		c, err := strconv.Atoi(countStr)
		if err == nil {
			count = c
		}
	}

	evt := db.GetLastEvents(id, count)
	evtjson, err := json.Marshal(evt)
	if err != nil {
		fmt.Fprint(res, "{\"type\":\"error\",\"message\":\"failed\"}")
		return
	}

	fmt.Fprint(res, string(evtjson))
}

func call(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	device := ps.ByName("device")
	function := ps.ByName("function")
	req.ParseForm()
	var f common.Method
	f.Name = function
	params := make([]common.Value, len(req.Form))
	p := 0
	for k, v := range req.Form {
		if len(v) > 0 {
			params[p].Id = k
			params[p].Value = v[0]
		}
	}
	f.Params = params
	methodStr, _ := json.Marshal(f)

	err := mqttcall(device+"-call", "["+string(methodStr)+"]", true)
	if err != nil {
		fmt.Fprintf(res, "{\"type\":\"error\",\"message\":\"%s\"}", err)
		return
	}
	fmt.Fprint(res, "{\"type\":\"success\",\"message\":\"Function called\"}")
}

func mqttcall(topic, payload string, retained bool) error {
	tries := 0
	for tries < 5 {
		token := mqtt_client.Publish(topic, 0, retained, payload)
		token.Wait()
		if token.Error() != nil {
			if token = mqtt_client.Connect(); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
				tries++
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return errors.New("Not connected")
}

func cors(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		h(res, req, ps)
		return
	})
}

func getPeriod(period string, current int) (start time.Time, end time.Time) {
	start = time.Now()
	switch period {
	case "week":
		weekday := int(start.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		weekday = weekday - 1
		start = start.Add(-1 * time.Duration(weekday) * 24 * time.Hour)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7*current)
		end = start.Add(7 * 24 * time.Hour).Add(-1 * time.Second)
		break
	case "month":
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, current, 0)
		end = start.AddDate(0, 1, -1).Add(-1 * time.Second)
		break
	case "day":
	default:
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, current)
		end = start.Add(24 * time.Hour).Add(-1 * time.Second)
		break
	}
	return
}

func Start() {
	cfg := readConfig()

	db = storage.NewBadger(cfg.db_path)

	router := httprouter.New()
	router.GET("/sensor/:ids", cors(sensor))
	router.GET("/meta/:ids", cors(meta))
	router.GET("/event/:id", cors(event))
	router.GET("/event/:id/:count", cors(event))
	router.GET("/devices", cors(devices))
	router.POST("/call/:device/:function", cors(call))

	opts := mqtt.NewClientOptions().AddBroker(cfg.mqtt_proto + "://" + cfg.mqtt_server + ":" + cfg.mqtt_port)
	opts.SetClientID(cfg.mqtt_client_id)
	if cfg.mqtt_user != "" {
		opts.SetUsername(cfg.mqtt_user)
	}
	if cfg.mqtt_password != "" {
		opts.SetPassword(cfg.mqtt_password)
	}
	mqtt_client = mqtt.NewClient(opts)
	if token = mqtt_client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}

	fmt.Println("API started...")
	log.Fatal(http.ListenAndServe(":"+cfg.web_port, router))
}

func readConfig() (cfg apiconfig) {
	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	cfg.db_path = os.Getenv("DB_PATH")

	var wu, wp bool
	cfg.web_user, wu = os.LookupEnv("WEB_USER")
	cfg.web_password, wp = os.LookupEnv("WEB_PASSWORD")
	cfg.web_port = os.Getenv("WEB_PORT")
	cfg.mqtt_proto = os.Getenv("MQTT_PROTOCOL")
	cfg.mqtt_server = os.Getenv("MQTT_SERVER")
	cfg.mqtt_port = os.Getenv("MQTT_PORT")
	cfg.mqtt_user = os.Getenv("MQTT_USER")
	cfg.mqtt_password = os.Getenv("MQTT_PASSWORD")
	cfg.mqtt_client_id = os.Getenv("MQTT_CLIENT_ID")

	if cfg.mqtt_proto == "" {
		cfg.mqtt_proto = fmt.Sprint(viper.Get("mqtt_protocol"))
	}
	if cfg.mqtt_server == "" {
		cfg.mqtt_server = fmt.Sprint(viper.Get("mqtt_server"))
	}
	if cfg.mqtt_port == "" {
		cfg.mqtt_port = fmt.Sprint(viper.Get("mqtt_port"))
	}
	if cfg.mqtt_user == "" {
		cfg.mqtt_user = fmt.Sprint(viper.Get("mqtt_user"))
	}
	if cfg.mqtt_password == "" {
		cfg.mqtt_password = fmt.Sprint(viper.Get("mqtt_password"))
	}
	if cfg.mqtt_client_id == "" {
		cfg.mqtt_client_id = fmt.Sprint(viper.Get("mqtt_client_id"))
	}

	if cfg.mqtt_proto == "" {
		cfg.mqtt_proto = "ws"
	}
	if cfg.mqtt_port == "" {
		cfg.mqtt_port = "9001"
	}
	if cfg.mqtt_client_id == "" {
		cfg.mqtt_client_id = "webapi"
	}

	if cfg.db_path == "" {
		cfg.db_path = fmt.Sprint(viper.Get("db_path"))
	}
	if !wu {
		cfg.web_user = fmt.Sprint(viper.Get("web_user"))
	}
	if !wp {
		cfg.web_password = fmt.Sprint(viper.Get("web_password"))
	}
	if cfg.web_port == "" {
		cfg.web_port = fmt.Sprint(viper.Get("web_port"))
	}

	if cfg.db_path == "" {
		cfg.db_path = "./db"
	}
	if cfg.web_port == "" {
		cfg.web_port = "8080"
	}

	return

}
