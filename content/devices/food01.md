---
date: 2017-06-16T17:45:11+01:00
title: Rabbit food dispenser
weight: 10
menu:
    main:
        parent: devices
---

IoT enabled rabbit food dispenser, automatic and controllable via Internet. 

**Github repository:** https://github.com/conejoninja/home_food

![rabbit food dispenser](/images/food/food0101.jpg "Rabbit food 01")



## Idea and motivation

The main idea was to be able to feed these two little rascals even when I'm not at home. Twice a day, a small quantity at breakfast and a bigger one for dinner (non-spoiled rabbits are supposed to eat timothy during the day when they are not sleeping). I would like to also measure the temperature in case it gets too hot, since rabbits don't do well on hot weather. 

The dispenser will wake up each 15 minutes, send the temperature and humidity to a server and check if it's time to feed the monsters. It's as autonomous as possible (it will feed them if it's time, even if there's no internet connection). The dispenser will notify if there's a problem with the temperature and humidity readings or if it's jammed. It could receive a command to dispense extra food even if it's not the time to.

![little rascals](/images/rabbits.jpg "Baloo & Moneypenny")



## Components

* ESP12F (or any other similar board)
* DS1307 AT24C32 (DS1307 module with 32K EEPROM)
* Temperature & humidity DTH11 sensor (or any DHTXX)
* L293B / L293D half-bridge
* DC motor (one of those cheap geared yellow motors)
* Rotary encoder
* DC-DC buck converter with 3.3V output (YL-46, AMS1117 based)
* 2x 10K ohm resistors
* 4x 1N4007
* 3D printed parts from STL files
* 6V power source
* 6x M3 screws and nuts
* Some wire
* Some glue

## The circuit

![the circuit](https://github.com/conejoninja/home_food/raw/master/circuit.png "The circuit")

The DC motor and the L293D/L293B are powered by the 6V, any phone charger rated at 5V will work too without any issue, but the motor could definitely use that extra power in case there's a jam in the dispenser. From there, the voltage is converted to 3.3V to feed the ESP12F and the rest of components with a DC-DC cheap converter.

Note: I used a L293B instead of the pictured L293D. I also used a DS1307 module with 32k EEPROM memory from ebay, that works at 3.3v instead of the 5v depicted in the image. The DC motor is different too (see more in the motor section). I couldn't find Fritzing components for them. 
  
### Connections

In case the diagram is not clear enough, here are the connections of the ESP12F.

| ESP12F pins   | Component |
| ------------- |-------------|
| RST | ESP12F GPIO16 (to be able to deepsleep) | 
| ADC | - | 
| CH_PD / EN | 3.3V | 
| GPIO16 | ESP12F RST | 
| GPIO14 | DHT11 data pin | 
| GPIO12 | Rotary encoder pin A | 
| GPIO13 | Rotary encoder pin B | 
| VCC | 3.3V (from DC-DC converter) | 
| GND | GND | 
| GPIO15 | GND | 
| GPIO2 | L293B IN2 | 
| GPIO0 | 3.3V | 
| GPIO4 | RTC SDA | 
| GPIO5 | RTC SCL | 
| RXD | L293B IN1 | 
| TXD | - | 
       

Due to the limitations of the ESP12F, the RXD pin is used to drive the motor. This will cause some issues is using the Serial to debug the code (it's always on). I tried using pins GPIO9 and GPIO10 that are usually not in the adapter plate, but there are additional limitation using them, and I couldn't make them work. I added a flag in the code to enable/disable the motor/Serial. You can not use the Serial and the motor at the same time, unless you change the pins. 


## Motor

For the first design I used a 360 servo motor I had lying around, it turns out it didn't have enough force and jams from time to time. I ended up using one of the *famously cheap* yellow dc motors used in several robot kits. These motors take from 3V to 6V, and have enough force for our purpose.

![dc motor](/images/food/motor1.jpg "dc motor")


There's a small bump on one side of the motor enclosure, that side is the good one. The other is attached via a smooth rod and can get loose after a while.


![dc motor](/images/food/motor2.jpg "dc motor")


At code level, to avoid jams we move forward the motor a few miliseconds and then a little bit backwards, we repeat this a few times each time we need to dispense some food. In the case a jam happens (very rare, but it does) and avoid overheating of the motor, we limit the number of times we try to move the motor. If it failed to dispense the desired amount of food, we emit an alert and try to dispense next time we wake up (we don't change the alarm). If we succeed with our task, we set a new alarm for the next day.
 
 
 To control the motor we use a *H-bridge* with the L293B, you could use a L293D or any other similar component. 
 
 
![h bridge](/images/food/hbridge.jpg "h bridge") 
 
 
## Rotary encoder

![rotary encoder](/images/food/rotary.encoder.jpg "rotary encoder")

To control if the dispenser is jammed or how much food is already served we'll use a rotary encoder. One without a PCB and with the one-side flat knob. The rotary encoder uses interrupts to know when it has moved from one position to another. Check which of the pins on your board support interrupts, in the ESP12F, only the GPIO16 doesn't support them, but it doesn't matter to us since we're using GPIO16 to wake from deepsleep. On regular Arduino boards, only a few pins support them, check [this table](https://www.arduino.cc/en/Reference/AttachInterrupt) for more information.  


## DHT11 - Temperature and humidity sensor

![dht11](/images/food/dht11.jpg "dht11")

DHT11, DHT22 or DHT21, any of them will work fine. They are cheap, being the DHT11 the cheapest among them, work well on 3.3V and 5V. We'll take 10 readings and calculate the average of both, temperature and humidity, to minimize wrong readings. Some of those readings, the first ones, may fail because the sensor needs to *warm up*, waiting 1 second between readings allow use to have at least a few good samples.


## DS1307 AT24C32 - RTC Clock with EEPROM memory

![rtc clock](/images/food/ds1307at24c32.jpg "RTC Clock")

We want to keep track of the time to be able to use the *alarms* and serve food whenever it's time to. We'll use the memory to know if the last time we failed and need to serve food again. Another possibility is to use the network time, but if the wifi is down for whatever reason we'll end up without a valid time.


There could be delays while connecting to the wifi network, or take more time to serve food for whatever reason. With the RTC it's possible to calculate the exact amount of time needed to wake up at the *exactly\** wanted time (at :00, :15, :30 and :45).


*\*the deepsleep of the ESP12F is a little bit imprecise, it will wake up a few seconds earlier than it should*





## STL/3D files

STL files for 3D printing could be found in the [project repository](https://github.com/conejoninja/home_food/tree/master/3Dfiles). They are quite printer-friendly, but supports are needed for a few of them. For everything food-related you should use PLA instead of ABS, as PLA is  Generally Recognized As Safe (GRAS) when used in contact with food by [some studies](http://www.sciencedirect.com/science/article/pii/027869159400145E), but there are a [few things](https://pinshape.com/blog/3d-printing-food-safe/) to have in mind.


Looking back, there are a few things I would change from the model. Printing takes some time and I don't like wasting plastic, I made improvements and small modifications using existing parts when possible. For example, the rotary encoder wires are visible, which could be a problem with animals that bite most things at their reach. 

Assembling is very easy and straight forward, but I made a video of the disassembling/exploding the model.

\
\


{{< youtube 4CCGwA13kiI >}}

