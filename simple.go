package main

import (
	//"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

const serverssl = "irc.oftc.net:6667"
const ownnick = "spacephone"

const commandprefix = "!"

//const serverssl = "irc.oftc.net:6697"

type cmdFunc = func(*irc.Event, string, string) error

var commands = map[string]cmdFunc{
	"say":        say,
	"silence":    silence,
	"spacestate": spacestate,
	"alert":      alert,
	"battery":    battery,
	"vibrate":    vibrate,
	"help":       help,
}

var irccon = irc.IRC("", "") // FIXME

var silenced = time.Unix(0, 0)

func say(e *irc.Event, parsed_message, reply string) error {
	if time.Now().Before(silenced) {
		return fmt.Errorf("Currently silenced")
	}
	Espeak(parsed_message)
	return nil
}

func silence(e *irc.Event, parsed_message, reply string) error {
	duration := time.Duration(1) * time.Minute
	if len(parsed_message) > 1 {
		dur, err := time.ParseDuration(parsed_message)
		if err != nil {
			return err
		}
		if dur.Minutes() > 5 {
			return fmt.Errorf("Duration too long")
		}

		duration = dur
	}
	silenced = time.Now().Add(duration)
	return nil
}

func spacestate(e *irc.Event, parsed_message, reply string) error {
	resp, err := http.Get("http://techinc.nl/space/spacestate")
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	state, err := ioutil.ReadAll(resp.Body)

	irccon.Privmsg(reply, string(state))

	return nil
}

func alert(e *irc.Event, parsed_message, reply string) error {
	go Espeak(parsed_message)
	go MqttSay(parsed_message, e.Nick)
	return nil
}

func battery(e *irc.Event, parsed_message, reply string) error {
	go func() {
		percentage, err := BatteryPercentage()
		if err != nil {
			irccon.Privmsg(reply, fmt.Sprintf("Error getting battery percentage: %v", err))
			return
		}

		irccon.Privmsg(reply, fmt.Sprintf("Percentage: %d", percentage))

	}()
	return nil
}

func vibrate(e *irc.Event, parsed_message, reply string) error {
	Vibrate()
	return nil
}

func help(e *irc.Event, parsed_message, reply string) error {
	irccon.Privmsg(reply, commandprefix+"{say,silence,spacestate,alert,battery}")
	return nil
}

func privmsg(e *irc.Event) {
	user_or_chan := e.Arguments[0]

	// XXX: This is really ugly and will not work if we get a different nick
	// somehow
	if user_or_chan == ownnick {
		user_or_chan = e.Nick
	}

	msg := e.Arguments[1]

	if strings.HasPrefix(msg, commandprefix) {
		s := strings.SplitN(msg[1:], " ", 2)
		cmd := s[0]
		parsed_msg := ""

		if len(s) > 1 {
			parsed_msg = s[1]
		}

		if cmdfunc, found := commands[cmd]; found {
			err := cmdfunc(e, parsed_msg, user_or_chan)
			if err != nil {
				irccon.Privmsg(user_or_chan, fmt.Sprintf("Error: %v", err))
			} else {
				//irccon.Privmsg(user_or_chan, fmt.Sprintf("OK/Executed"))
			}
		} else {
			irccon.Privmsg(user_or_chan, "Invalid command")
			log.Println("Unknown command:", cmd)
		}
	}

}

func main() {
	MqttInit()

	ircnick1 := ownnick
	irccon = irc.IRC(ircnick1, ownnick)
	irccon.VerboseCallbackHandler = false
	irccon.Debug = false
	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join("#techinc")
		irccon.Join("#techinc-spacephone")
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	irccon.AddCallback("PRIVMSG", privmsg)
	err := irccon.Connect(serverssl)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
	irccon.Loop()
}
