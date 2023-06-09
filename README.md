# rpict-mqtt-bridge

This is a very simple functional, untested and "as-is" public domain implementation of an MQTT bridge reading Lechacals RPICT Raspberry Pi hats.

My personal setup has a "RPICT3V4v5" main hat, measuring the 3 phase voltages, then 4 extra hats all "RPICT8v5", measuring currents.

I use this from the Raspberry Pi running raspbian, broadcasting to a Home Assistant server running an MQTT broker.

If you have awesome suggestions, please open up an issue, or even better: a pull request. This project is low priority for me though: "if it works, it works" (so forking might be the better option, lol).

## Usage

You can start the service with `go run service.go`. I recommend running it in a GNU screen.
This project supports Home Assistant MQTT auto discovery and will broadcast discovery message every 60 seconds.

## Config

The config has three main sections

```json
{
    "Rpict": {
        "Device": "/dev/ttyAMA0",      // Where can we find the serial socket for RPICT
        "Baudrate": 38400              // Baudrate to use
    },
    "MqttBroker": {
        "Host": "<hostname>",          // Where can we find the MQTT broker
        "Port": 1337,
        "Path": "RPICT",               // Path to use for broadcasting all channels / measuremnts
        "User": "<mqtt_user>",         // Mqtt broker authentication
        "Password": "<mqtt_password>"
    },
    "Channels": [
        // Measurement is a Home Assistant sensor device class: https://www.home-assistant.io/integrations/sensor/
        { "Phases": 3, "Measurement": "voltage", "Description": "Net voltages", "Topic": "mains_voltage" },
        { "Phases": 3, "Measurement": "frequency", "Description": "Net frequenties", "Topic": "mains_frequency" },
        { "Phases": 3, "Measurement": "power", "Description": "Zonnepanelen", "Topic": "pv" },
        { "Phases": 3, "Measurement": "power", "Description": "Schuur", "Topic": "shed" },
        { "Phases": 3, "Measurement": "power", "Description": "Laadpaal", "Topic": "carcharger" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L1.4", "Topic": "GL1_4" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L1.5", "Topic": "GL1_5" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L1.6", "Topic": "GL1_6" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L2.1", "Topic": "GL2_1" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L2.2", "Topic": "GL2_2" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L2.3", "Topic": "GL2_3" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L3.5", "Topic": "GL3_5" },
        { "Phases": 1, "Measurement": "power", "Description": "Groep L3.6", "Topic": "GL3_6" },
        { "Phases": 1, "Measurement": "current", "Description": "Net aarde", "Topic": "mains_ground" }
    ]
}
```
