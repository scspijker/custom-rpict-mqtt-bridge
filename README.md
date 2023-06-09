# rpict-mqtt-bridge

This is a very simple functional, untested and "as-is" public domain implementation of an MQTT bridge reading Lechacal's RPICT Raspberry Pi hats.

My personal setup have a "RPICT3V4v5" main hat, measuring the 3 phase voltages, then 4 extra hats all "RPICT8v5", measuring currents.

I use this from the Raspberry Pi running raspbian, sending to a Home Assistant server running an MQTT broker.

If you have awesome suggestions, please open up an issue, or even better: a pull request. This project is low priority for me though: "if it works, it works".

## Usage

You can start the service with `go run service.go`. I recommend running it in a GNU screen.
This project supports Home Assistant MQTT auto discovery and will broadcast discovery message every 60 seconds.

## Config

The config has three main sections

### Rpict

Where can we find the serial socket, and which baudrate is the RPICT configured at to open up the serial connection.

### MqttBroker

Where can we find the broker, at what path do you want to publish, and what are the username/password for access.

### Channels

You can measure 1 or more phases for each channel by setting the amount of phases.

The "Measurement" is a string representing a Home Assistant sensor device class.

Description is currently unused, but for my own sanity when reading the config.

Topic is the topic where the values should be published on MQTT.