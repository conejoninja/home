---
date: 2017-06-16T17:45:11+01:00
title: Rabbit feeder v2
weight: 10
menu:
    main:
        parent: devices
        url: devices/food02/
---

IoT enabled rabbit feeder, automatic and controllable via Internet. 

**Github repository:** https://github.com/conejoninja/home_food


## Idea and motivation

Learn from the experience and improve the previous versions of the feeder.

As with the old prototype, the feeder will wake up each 15 minutes, send the temperature and humidity to a server and check if it's time to feed the monsters. It's as autonomous as possible (it will feed them if
it's time, even if there's no internet connection). The feeder will notify if there's a problem with the temperature and humidity readings or if it's jammed. It could receive a command to dispense extra food even if it's not the time to.

![little rascals](https://conejoninja.github.io/home//images/rabbits.jpg "Baloo & Moneypenny")



## Components

* Wemos D1 Mini (contains a ESP12F, but any other similar board works great)
* DS1307 AT24C32 (DS1307 module with 32K EEPROM)
* Temperature & humidity DTH11 sensor (or any DHTXX)
* DRV8825 (or stepstick) driver
* Stepper motor (NEMA17)
* DC-DC step-down converter, from 12V to 5V. with 3.3V output (LM2596 for example)
* 3D printed parts from STL files
* 6V-12 power source
* Some M3 screws and nuts
* Some wire
* Some glue

## The circuit

![the circuit](https://github.com/conejoninja/home_food/raw/master/circuit_v2.png "The circuit")

The stepper motor (via DRV8825) is powered by the 12V power source.
From there, the voltage is converted to 5V to feed the Wemos D1 Mini (via USB, with is regulated) and the rest of components with a DC-DC converter (LM2596).



### Connections

In case the diagram is not clear enough, here are the connections of the ESP12F.

| Wemos D1 Mini pins   | Component |
| ------------- |-------------|
| RST | Wemos D0 pin (to be able to deepsleep) |
| A0 | - |
| D5 | DHT11 data pin |
| D6 | DRV8825 DIR |
| D7 | DRV8825 STP |
| D8 | DRV8825 SLP |
| 3V3 | - |
| 5V | 5V (from DC-DC converter) |
| GND | GND |
| D4 | - |
| D3 | - |
| D2 | RTC SDA |
| D1 | RTC SCL |
| RX | - |
| TX | - |


## Motor

The major change from the first prototype is replacing the dc motor with a stepper controlled by a DRV8825 (stepstick works well too). A set of gears has been included in the new design too to make it a little more powerful.
 


## DHT11 - Temperature and humidity sensor

![dht11](https://conejoninja.github.io/home/images/food/dht11.jpg "dht11")

DHT11, DHT22 or DHT21, any of them will work fine. They are cheap, being the DHT11 the cheapest among them, work well on 3.3V and 5V. We'll take 10 readings and calculate the average of both, temperature and humidity, to minimize wrong readings. Some of those readings, the first ones, may fail because the sensor needs to *warm up*, waiting 1 second between readings allow use to have at least a few good samples.


## DS1307 AT24C32 - RTC Clock with EEPROM memory

![rtc clock](https://conejoninja.github.io/home/images/food/ds1307at24c32.jpg "RTC Clock")

We want to keep track of the time to be able to use the *alarms* and serve food whenever it's time to. We'll use the memory to know if the last time we failed and need to serve food again. Another possibility is to use the network time, but if the wifi is down for whatever reason we'll end up without a valid time.


There could be delays while connecting to the wifi network, or take more time to serve food for whatever reason. With the RTC it's possible to calculate the exact amount of time needed to wake up at the *exactly\** wanted time (at :00, :15, :30 and :45).


*\*the deepsleep of the ESP12F/Wemos is a little bit imprecise, it will wake up a few seconds earlier than it should*





## STL/3D files


STL files for 3D printing could be found in the [project repository](https://github.com/conejoninja/home_food/tree/master/3Dfiles). They are quite printer-friendly, but supports are needed for a few of them. For everything food-related you should use PLA instead of ABS, as PLA is  Generally Recognized As Safe (GRAS) when used in contact with food by [some studies](http://www.sciencedirect.com/science/article/pii/027869159400145E), but there are a [few things](https://pinshape.com/blog/3d-printing-food-safe/) to have in mind.


Assembling is very easy and straight forward, but I made a video of it.

\
\


{{< youtube zQorAdMOrRM >}}



## Discovery JSON

During discovery, food01 will send the following description message:

```
{
    "id":"food02",
    "name":"Dulicomida 3000",
    "version":"2.0.0",
    "out":[
        {"id":"t1","name":"temperature"},
        {"id":"h1","name":"humidity"},
        {"id":"m1","name":"memory1"},
        {"id":"m2","name":"memory2"},
        {"id":"m3","name":"alarm1"},
        {"id":"m4","name":"alarm2"},
        {"id":"m5","name":"bigqty"},
        {"id":"m6","name":"smallqty"}
    ],
    "methods":[
        {"name":"food"},
        {"name":"ping"},
        {"name":"setmem","params":[{"name":"id"},{"name":"value"}]},
        {"name":"getmem"}
    ]
}
```

There are 8 output values. m1-m6 are explained in the [section below](#memory) and are only updated on request (by a *getmem* call). t1 (temperature) and h1 (humidity), are sensor data that are update every 15 minutes if an internet connection is available and the MQTT server is working. You could make 4 different request to the device. *ping* will reply with *pong* just to check the device is still alive. A *food* request will dispense a small amount of food. You could change the configuration of the device with *setmem* and get the actual values in memory with *getmem*. Since the device is mostly deepsleeping, it will execute the commands the next time it wakes up as long as the MQTT message was marked as persistent. You could send several commands at the same time in the same request, but if you make several request before it wakes up, only the last one will be executed.



## Memory

The rabbit feeder stores some settings (and other data) in a non-volatile memory chip, thanks to the DS1307 AT24C32. It can be configured remotely with the *setmem* function call, more information in the [protocol page (method-calling-messages)](/home/protocol/#method-calling-messages). There are six different settings:

* memory1: unix timestamp (from Jan 1st 2000) for next alarm1 to be executed. Example value: *553291190* 
* memory2: unix timestamp (from Jan 1st 2000) for next alarm2 to be executed. Example value: *553330790*
* alarm1: time for the first alarm (big amount of food) in 24h format without symbols. Example value: *2030* (2030 = 20:30 = 8:30pm)
* alarm2: time for the second alarm (small amount of food) in 24h format without symbols. Example value: *2030* (745 = 7:45 = 7:45am)
* bigqty: minimum number of steps of the rotary encoder to dispense food for the first alarm
* smallqty: minimum number of steps of the rotary encoder to dispense food for the second alarm

**Note**: memory1 and memory2 are the timestamp from the Jan 1st of 2000, you need to substract 946684800 from the *real* timestamp. 
07/13/2017 @ 7:59pm (UTC) is equivalent to 1499975990 and for the device, it's 553291190.

