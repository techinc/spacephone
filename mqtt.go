package main

import (
	"encoding/json"
	"fmt"
	//"os"
	//"os/signal"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

var mqttcli = client.New(&client.Options{
	// Define the processing of the error handler.
	ErrorHandler: func(err error) {
		fmt.Println(err)
	},
})

func MqttInit() {
	// Set up channel on which to send signal notifications.
	//sigc := make(chan os.Signal, 1)
	//signal.Notify(sigc, os.Interrupt, os.Kill)

	// Connect to the MQTT Server.
	err := mqttcli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  "mqtt.ti:1883",
		ClientID: []byte("example-client"),
	})
	if err != nil {
		panic(err)
	}
}

func MqttSay(message, nick string) error {
	payload := map[string]interface{}{
		"text": message,
		"who":  nick,
	}
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish a message.
	return mqttcli.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte("ledslie/alert/1/spacealert"),
		Message:   j,
	})
}

func MqttIrc(src, msg string) error {
	payload := map[string]interface{}{
		"user": src,
		"msg":  msg,
	}
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish a message.
	return mqttcli.Publish(&client.PublishOptions{
		QoS: mqtt.QoS0,
		// TODO: Channel hardcoded, should not relay privmsg and so on
		TopicName: []byte("irc/techinc"),
		Message:   j,
	})
}

func MqttIrcJoin(who string) error {
	payload := map[string]interface{}{
		"join": who,
	}
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish a message.
	return mqttcli.Publish(&client.PublishOptions{
		QoS: mqtt.QoS0,
		// TODO: Channel hardcoded, should not relay privmsg and so on
		TopicName: []byte("irc/techinc"),
		Message:   j,
	})
}

func MqttIrcPart(who string) error {
	payload := map[string]interface{}{
		"part": who,
	}
	j, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish a message.
	return mqttcli.Publish(&client.PublishOptions{
		QoS: mqtt.QoS0,
		// TODO: Channel hardcoded, should not relay privmsg and so on
		TopicName: []byte("irc/techinc"),
		Message:   j,
	})
}

func MqttStop() {
	defer mqttcli.Terminate()
}
