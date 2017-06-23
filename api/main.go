package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"os"

	"github.com/conejoninja/home/storage"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

type Key int

const MyKey Key = 0

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var db_path, web_user, web_password, web_port string

// STORAGE
var db storage.Storage

func login(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Access-Control-Allow-Origin", "*")

	req.ParseForm()
	user := req.Form.Get("user")
	pass := req.Form.Get("pass")

	if user != web_user || pass != web_password {
		fmt.Fprint(res, "{\"error\":\"failed\"}")
		return
	}

	expireToken := time.Now().Add(time.Hour * 1).Unix()
	expireCookie := time.Now().Add(time.Hour * 1)

	claims := Claims{
		"myusername",
		jwt.StandardClaims{
			ExpiresAt: expireToken,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, _ := token.SignedString([]byte("secret"))

	cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true}
	http.SetCookie(res, &cookie)

	fmt.Fprint(res, "{\"success\":\"enjoy your flight\"}")
}

// middleware to protect private pages
func validate(page http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		cookie, err := req.Cookie("Auth")
		if err != nil {
			http.NotFound(res, req)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return []byte("secret"), nil
		})
		if err != nil {
			http.NotFound(res, req)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page(res, req.WithContext(ctx))
		} else {
			http.NotFound(res, req)
			return
		}
	})
}

func sensor(res http.ResponseWriter, req *http.Request) {
}

func devices(res http.ResponseWriter, req *http.Request) {
	devices := db.GetDevices()
	devsjson, err := json.Marshal(devices)
	if err != nil {
		fmt.Fprint(res, "{\"error\":\"failed\"}")
		return
	}

	fmt.Println(string(devsjson))
	fmt.Fprint(res, string(devsjson))
}

func logout(res http.ResponseWriter, req *http.Request) {
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now()}
	http.SetCookie(res, &deleteCookie)
	return
}

func main() {
	db_path, web_user, web_password, web_port = readConfig()

	db = storage.NewBadger(db_path)

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", validate(logout))

	http.HandleFunc("/sensor", sensor)
	http.HandleFunc("/devices", devices)

	http.ListenAndServe(":"+web_port, nil)
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
