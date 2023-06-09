package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "github.com/tarm/serial"
    "bufio"
    "strings"
    "strconv"
    "fmt"
    "os"
    "time"
    "math"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

func sendMqtt(measurements []*Measurement, mqttClient mqtt.Client, config Config) {
    hostname, err := os.Hostname()
    if err != nil { log.Fatal(err) }
    for channelIndex, channel := range config.Channels {
        var value []byte
        var err error
        if channel.Phases != 1 {
            var values = make(map[string]string)
            for phase := 1; phase <= channel.Phases; phase++ {
                key := fmt.Sprintf("L%d", phase)
                values[key] = fmt.Sprint(measurements[channelIndex].Values[phase-1])
            }
            value, err = json.Marshal(values)
        } else {
            value, err = json.Marshal(fmt.Sprint(measurements[channelIndex].Values[0]))
        }
        if err != nil { log.Fatal(err) }
        topic := fmt.Sprintf("%s/%s/%s", hostname, config.MqttBroker.Path, channel.Topic)
        token := mqttClient.Publish(topic, 0, true, value)
        token.Wait()
        if token.Error() != nil { log.Fatal(token.Error()) }
    }
    log.Println("Sent all values to Mqtt topics")
}

type HomeAssistantAutoDiscoveryMessage struct {
    DeviceClass     string  `json:"device_class"`
    Name            string  `json:"name"`
    StateTopic      string  `json:"state_topic"`
    Unit            string  `json:"unit_of_measurement"`
    ValueTemplate   string  `json:"value_template"`
    UniqueId        string  `json:"unique_id"`
}

func advertiseHomeAssistant(mqttClient mqtt.Client, config Config) {
    hostname, err := os.Hostname()
    if err != nil { log.Fatal(err) }

    for _, channel := range config.Channels {
        if channel.Phases != 1 {
            for phase := 0; phase < channel.Phases; phase++ {
                message := new(HomeAssistantAutoDiscoveryMessage)
                message.DeviceClass = channel.Measurement
                message.Name = fmt.Sprintf("%s_L%d", channel.Topic, phase + 1)
                message.StateTopic = fmt.Sprintf("%s/%s/%s", hostname, config.MqttBroker.Path, channel.Topic)
                message.Unit = unitForMeasurement(channel.Measurement)
                message.ValueTemplate = fmt.Sprintf("{{ value_json.L%d }}", phase + 1)
                message.UniqueId = message.Name

                jsonMessage, err := json.Marshal(message)
                if err != nil { log.Fatal(err) }
                token := mqttClient.Publish(fmt.Sprintf("homeassistant/sensor/%s/%s_L%d/config", hostname, channel.Topic, phase + 1), 0, true, jsonMessage)
                token.Wait()
                if token.Error() != nil { log.Fatal(token.Error()) }
            }
        } else {
            message := new(HomeAssistantAutoDiscoveryMessage)
            message.DeviceClass = channel.Measurement
            message.Name = channel.Topic
            message.StateTopic = fmt.Sprintf("%s/%s/%s", hostname, config.MqttBroker.Path, channel.Topic)
            message.Unit = unitForMeasurement(channel.Measurement)
            message.ValueTemplate = "{{ value_json }}"
            message.UniqueId = message.Name

            jsonMessage, err := json.Marshal(message)
            if err != nil { log.Fatal(err) }
            token := mqttClient.Publish(fmt.Sprintf("homeassistant/sensor/%s/%s/config", hostname, channel.Topic), 0, true, jsonMessage)
            token.Wait()
            if token.Error() != nil { log.Fatal(token.Error()) }
        }
    }
    log.Println("HomeAssistant auto discovery sent")
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    log.Println("Mqtt connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    log.Printf("Mqtt connection lost: %v", err)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    log.Printf("Mqtt received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

func connectMqtt(config Config) mqtt.Client {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MqttBroker.Host, config.MqttBroker.Port))
    opts.SetClientID("RpictGoBridge")
    opts.SetUsername(config.MqttBroker.User)
    opts.SetPassword(config.MqttBroker.Password)
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler

    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    return client
}

type Measurement struct {
    Values      []float64
}

func unitForMeasurement(measurement string) string {
    switch measurement {
        case "voltage": return "V"
        case "frequency": return "Hz"
        case "power": return "W"
        case "current": return "A"
        default: return "?"
    }
}

func parseRpictLine(line string, config Config) []*Measurement {
    values := strings.Split(line, " ")[1:]

    var parsedValues []*Measurement = make([]*Measurement, len(config.Channels))

    i := 0
    for channelIndex, channel := range config.Channels {

        measurement := new(Measurement)
        measurement.Values = make([]float64, channel.Phases)
        log.Print(fmt.Sprintf("%s ", channel.Topic))

        for phase := 0; phase < channel.Phases; phase++ {
            phaseValue, err := strconv.ParseFloat(values[i], 64)
            if err != nil { log.Fatal(err) }
            i++

            if channel.Measurement == "power" || channel.Measurement == "current" {
                if phaseValue != 0 {
                    phaseValue *= -1
                }
                if channel.Measurement == "power" {
                    phaseValue = math.Round(phaseValue)
                }
            }

            measurement.Values[phase] = phaseValue
            log.Print(fmt.Sprintf("  L%v: %v ", phase+1, phaseValue))
        }

        parsedValues[channelIndex] = measurement
    }

    log.Println("")
    return parsedValues
}

func listen(config Config, client mqtt.Client) {
    serialConfig := &serial.Config{ Name: config.Rpict.Device, Baud: config.Rpict.Baudrate }
    log.Println(serialConfig)
    stream, err := serial.OpenPort(serialConfig)
    if err != nil {
            log.Fatal(err)
    }

    scanner := bufio.NewScanner(stream)
    for scanner.Scan() {
        line := scanner.Text()
        messages := parseRpictLine(line, config)
        sendMqtt(messages, client, config)
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}

func main() {
    log.Println("Reading config")
    config := readConfig()

    log.Println("Config read, connecting MQTT")
    mqttClient := connectMqtt(config)

    log.Println("Starting advertisement loop")
    go func() {
        for {
            advertiseHomeAssistant(mqttClient, config)
            time.Sleep(60 * time.Second)
        }
    }()

    log.Println("Starting MQTT listener")
    listen(config, mqttClient)
}

type Config struct {
    Rpict       struct {
        Device      string
        Baudrate    int
    }
    MqttBroker  struct {
        Host        string
        Port        int
        Path        string
        User        string
        Password    string
    }
    Channels    []ConfigChannel
}

type ConfigChannel struct {
    Phases      int
    Measurement string
    Description string
    Topic       string
}

func readConfig() Config {
	content, err := ioutil.ReadFile("config.json")
    if err != nil { log.Fatal("Could not read config: ", err) }

    var config Config
    err = json.Unmarshal(content, &config)
    if err != nil { log.Fatal("Could not parse config: ", err) }

    log.Println(config)

    return config
}

