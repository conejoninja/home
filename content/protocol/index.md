---
date: 2017-06-12T17:45:11+01:00
title: Protocol
weight: 20
---

Simple protocol to discover and communicate with IoT devices on a network. The devices communicate with each other through the same central server(s) (a single server or a mesh), that may or may not be in the same local network. This makes possible the communication of devices in different physical networks.
 
 
## Discovery of devices

Devices that join the network will broadcast their description. The message will be send to the **discovery** channel. 

### Discovery message structure 
```json
ID: Unique ID of the device (e.g. temp-31415)
Name: Name of the device (e.g. ACME Temperature Sensor)
Version: Version of the protocol the device is using
Date: Optional field to force the date of the message
Out[] : Array of data types the device publish to their channel
    ID: Unique ID of the sensor (e.g. t1)
    Name: Name of the data (e.g. temperature)
    Type: Primitive type (number, text, complex object....). Default type is number, so we could save some data.
    Min? Max? : Optional fields yet to be defined in case we want to better represent them?
Methods[] : Array of methods available to communicate with the device
    Name: Method’s name (e.g. setColor, On, Off,...)
    Params[]: Array of parameters for the method
        Name: Param’s id/name
        Type: Primitive type (number, text,....). Default type is number, so we could save some data.
        Min? Max?: Optional limits to the value
```        
        
        
### Short Discovery Message / Handshake
To avoid polluting discovery channel with devices that frequently enter sleep mode or that reconnect to the network several times, devices may send only their ID to the discovery channel, then it may be asked to publish the full discovery message by sending the message “whoareyou” to “{ID}-events” channel
 
 
## Sensor information messages
A device could publish their sensor data to their own channel {ID}. A message could be composed of the data of one or more sensors, in an array.
```json
[]{
    ID: Sensor’s ID (e.g. t1)
    Value: Value of the sensor (e.g. 27.4)
    Unit?: If any (e.g. “ºC”)
    Date?: Optional field to force the date of the message
}
```

Example:
```json
[
    {ID: “t1”, value: 27.4, unit: “ºC”}, // temperature 
    {ID: “h1”, value: 54, unit: “%”} // humidity
]
```
 
 
## Method calling messages

To control the devices, we could call the methods described in their discovery message on the channel **{device_id}-call**. A message could have one or more calls to different methods in the same message. 
```json
[]{
    Name: Method’s name (e.g. On)
    Parameters?: Parameters of the method (optional)
}
``` 
 
 
## General Event Channel
Devices should be able to emit events to the **events** channel,  to communicate messages not related to the sensor data. For example, if a required action failed, resulted in success or when it ended.

```json
ID: Unique ID of the device (emitter) (e.g. temp-31415)
Message: Text of the message (e.g. “Task ended at @ and lasted @”)
Priority: 0 (OK), 1 (WARNING), 2 (ERROR)
Extra[]: {
    Key: Name of the data (e.g. duration)
    Value: Data (e.g. 1m35s)
}
Date?: Optional field to force the date of the message
```