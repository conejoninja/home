---
date: 2017-06-12T17:45:11+01:00
title: Home
weight: 10
---

To log all the sensor data and be able to process it later, we use what we call *Home*. It's written in [Go](http://golang.org/) and made of different configurable *module*. The two principal tasks are to log the sensor
data in the *Logger* and serve it via the *API*.
 
 
## Logger

[Source code](https://github.com/conejoninja/home/tree/master/logger)

Save sensor data and other information from the MQTT network in a database.


## API

[Source code](https://github.com/conejoninja/home/tree/master/api)

Create a HTTP server and listen for requests.

## Configuration

By default, *Home* will use environment variables if available for its configuration, and if it's not the case, it will read the setting from a [config.yml](https://github.com/conejoninja/home/blob/master/cmd/config-sample.yml) file in the same folder. You could use a mix of env vars and config.yml. I choose this to be able to create a [docker image](https://hub.docker.com/r/conejo/home/) and configure it later via docker-compose (or docker run) without using a custom image with the file. I know, it's discouraged and not the best practices, but I also needed to make it work.


### config.yml options

* db_path: path to your database
* mqtt_server: your MQTT server
* mqtt_port: MQTT port (9001 by default)
* mqtt_protocol: MQTT protocol (ws by default)
* mqtt_user: MQTT user
* mqtt_password: MQTT password
* mqtt_client_id: client id, it could be anything
* api_enabled: to enable the API or only log sensor data
* api_port: API port
* timezone: Important if your server is not in the same timezone as you (example "Europe/Madrid")
* websocket_enabled: create a websocket server and publish all the messages to see realtime information
* websocket_port: websocket port
* tg_enabled: use a Telegram bot to notify you events on your network
* tg_token: Telegram bot token
* tg_chats: Telegram chat's ID to notify


## Environment variables 

Enviroment variables are exactly the same as the options of config.yml but in *upper case*, ie. db_path => DB_PATH.