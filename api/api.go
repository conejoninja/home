package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log"
	"time"

	"strconv"
	"strings"

	"errors"

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/julienschmidt/httprouter"
)

type ApiConfig struct {
	Port     string
	TimeZone string
	location *time.Location
	Enabled  bool
}

type sensorResponse map[string]map[string][]common.Value
type metaResponse map[string]common.Meta

var db storage.Storage
var c mqtt.Client
var cfg ApiConfig

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
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, cfg.location)

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
		token := c.Publish(topic, 0, retained, payload)
		token.Wait()
		if token.Error() != nil {
			if token = c.Connect(); token.Wait() && token.Error() != nil {
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
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, cfg.location).AddDate(0, 0, 7*current)
		end = start.Add(7 * 24 * time.Hour).Add(-1 * time.Second)
		break
	case "month":
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, cfg.location).AddDate(0, current, 0)
		end = start.AddDate(0, 1, -1).Add(-1 * time.Second)
		break
	case "day":
	default:
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, cfg.location).AddDate(0, 0, current)
		end = start.Add(24 * time.Hour).Add(-1 * time.Second)
		break
	}
	return
}

func Start(apicfg ApiConfig, dbcon storage.Storage, mqttclient mqtt.Client) {
	cfg = apicfg
	db = dbcon
	c = mqttclient

	var err error
	cfg.location, err = time.LoadLocation(cfg.TimeZone)
	if err != nil {
		cfg.location = time.UTC
	}

	router := httprouter.New()
	router.GET("/sensor/:ids", cors(sensor))
	router.GET("/meta/:ids", cors(meta))
	router.GET("/event/:id", cors(event))
	router.GET("/event/:id/:count", cors(event))
	router.GET("/devices", cors(devices))
	router.POST("/call/:device/:function", cors(call))

	go func() {
		for {
			fmt.Println("API started...")
			log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
			fmt.Println("(╯°□°)╯ API server failed, restarting in...")
			time.Sleep(5 * time.Second)
		}
	}()
}
