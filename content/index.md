---
date: 2017-06-11T18:28:13+01:00
title: Home Automation project
type: index
weight: 90
---

Amateur project to automate a few task in my home and communicate with some IoT devices. I used a mix of languages, frameworks and platforms. I will try to document as much as possible, but remember, this is an amateur project. 




## Global view

In order to interact together, devices communicates with each other through a MQTT server using a custom JSON-based protocol. Besides the devices, there are a few "modules" too, those modules are programs that consume the MQTT messages, perform actions based of information and allow human interaction (Web, App).


![Global view](https://conejoninja.github.io/home/images/diagram.png)

Devices on the left size of the image (rover, food, device x) are connected to through wifi or ethernet to the home network. Modules on the right (logger, cc, dashboard) reside in an external server along with the MQTT server, but they could live in a Raspberry Pi or similar inside the home network.

## Protocol

I decided to created my own protocol sinces I didn't find any that suits my needs and was easy to use. The protocol is best described at the [protocol section](https://conejoninja.github.io/home/protocol/).


## MQTT server

For the MQTT server I choose to use [Mosquitto](https://mosquitto.org/) inside a [Docker](http://docker.com/) running on a external server.

## Logger

Logger is running program that listen to everything on the network and saves the important information (sensor data, events and devices' descriptions) to a database, for later consumption and stats. Logger is written in Go.

## Database

For the database I choose to use [Badger](https://github.com/dgraph-io/badger), Go-native key-value database. 

## Command and Control (CC)

Here is where the real magic happens, Go program that automates tasks based on the information stored in the database, such as if it's too hot, switch on the fan.

## Dashboard / API

The dashboard/API to consume the information off of the database and show some nice charts to the users (web, mobile app) and allow them to manually perform actions or request additional data.
 
 
## Devices

### Food-01
Food-01 is a rabbit automatic feeder. It dispenses food twice a day, small quantity on breakfast and a bigger one at dinner time. It communicates the status, if there was any problem during its operation. It accepts an user command to dispense extra food at any given time. It also sends temperature/humidity  