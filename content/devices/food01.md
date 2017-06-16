---
date: 2017-06-16T17:45:11+01:00
title: Rabbit food dispenser
weight: 10
---

IoT enabled rabbit food dispenser, automatic and controllable via Internet. 

**Github repository:** https://github.com/conejoninja/home_food

![rabbit food dispenser](/images/food0101.jpg "Rabbit food 01")



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
* Some wire

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




{{< youtube 4CCGwA13kiI >}}



WORK IN PROGRESS