package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"log"
	"time"

	"github.com/conejoninja/home/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

var db_path, web_user, web_password, web_port string

// STORAGE
var db storage.Storage

func sensor(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {

	id := ps.ByName("id")
	now := time.Now()

	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekday = weekday - 1

	//start := now.Add(-1 * time.Duration(weekday) * 24 * time.Hour)
	start := now.Add(-24 * time.Hour)

	values := db.GetValuesBetweenTime(id, start, now)

	valStr, _ := json.Marshal(values)
	fmt.Fprint(res, string(valStr))
}

func meta(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
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

func cors(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		h(res, req, ps)
		return
	})
}

func Start() {
	db_path, web_user, web_password, web_port = readConfig()

	db = storage.NewBadger(db_path)

	router := httprouter.New()
	router.GET("/sensor/:id", cors(sensor))
	router.GET("/meta/:id", meta)
	router.GET("/devices", devices)

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
