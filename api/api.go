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

	"github.com/conejoninja/home/common"
	"github.com/conejoninja/home/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

var db_path, web_user, web_password, web_port string

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
		fmt.Println(id + "-" + strconv.Itoa(int(start.Unix())), start, strconv.Itoa(int(time.Now().Unix())), time.Now())
		response[id] = db.GetMeta([]byte(id + "-day-" + strconv.Itoa(int(start.Unix()))))
		fmt.Println(response[id])
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

	fmt.Println(string(devsjson))
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
		fmt.Fprint(res, "{\"error\":\"failed\"}")
		return
	}

	fmt.Fprint(res, string(evtjson))
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
	db_path, web_user, web_password, web_port = readConfig()

	db = storage.NewBadger(db_path)

	router := httprouter.New()
	router.GET("/sensor/:ids", cors(sensor))
	router.GET("/meta/:ids", cors(meta))
	router.GET("/event/:id", cors(event))
	router.GET("/event/:id/:count", cors(event))
	router.GET("/devices", cors(devices))

	fmt.Println("API started...")
	log.Fatal(http.ListenAndServe(":"+web_port, router))
}

func readConfig() (db_path, web_user, web_password, web_port string) {
	if _, err := os.Stat("./config.yml"); err != nil {
		fmt.Println("Error: config.yml file does not exist")
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	db_path = os.Getenv("DB_PATH")

	var wu, wp bool
	web_user, wu = os.LookupEnv("WEB_USER")
	web_password, wp = os.LookupEnv("WEB_PASSWORD")
	web_port = os.Getenv("WEB_PORT")

	if db_path == "" {
		db_path = fmt.Sprint(viper.Get("db_path"))
	}
	if !wu {
		web_user = fmt.Sprint(viper.Get("web_user"))
	}
	if !wp {
		web_password = fmt.Sprint(viper.Get("web_password"))
	}
	if web_port == "" {
		web_port = fmt.Sprint(viper.Get("web_port"))
	}

	if db_path == "" {
		db_path = "./db"
	}
	if web_port == "" {
		web_port = "8080"
	}

	return

}
