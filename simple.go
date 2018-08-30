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

// When is a battery considered low?
const battery_low = 42

// When to check if the battery is low
const battery_check_time = time.Second * 30

// How often to remind about a low battery
const battery_reminder_time = time.Second * 15

// When to check if the battery charging state has changed
const battery_charge_check_time = time.Second * 30

var battery_charging = false

const commandprefix = "!"

//const serverssl = "irc.oftc.net:6697"

type cmdFunc = func(*irc.Event, string, string) error
type recFunc = func() error

var commands = map[string]cmdFunc{
	"say":        say,
	"silence":    silence,
	"spacestate": spacestate,
	"alert":      alert,
	"battery":    battery,
	"vibrate":    vibrate,
	"mpd":        mpdf,
	"help":       help,
}

var irccon = irc.IRC("", "") // FIXME

var silenced = time.Unix(0, 0)
var last_battery_warning = time.Unix(0, 0)

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
		charging, err := BatteryCharging()
		if err != nil {
			irccon.Privmsg(reply, fmt.Sprintf("Error getting battery percentage: %v", err))
			return
		}

		is_charging := "yes"
		if !charging {
			is_charging = "no"
		}

		irccon.Privmsg(reply, fmt.Sprintf("Percentage: %d. Charging: %s", percentage, is_charging))

	}()
	return nil
}

func vibrate(e *irc.Event, parsed_message, reply string) error {
	Vibrate()
	return nil
}

func mpdf(e *irc.Event, parsed_message, reply string) error {
	if parsed_message == "ping" {
		err := MpdPing()
		if err != nil {
			return err
		}
		irccon.Privmsg(reply, "pong")
	} else if parsed_message == "status" {
		attrs, err := MpdStatus()
		if err != nil {
			return err
		}
		irccon.Privmsg(reply, fmt.Sprintf("%v", attrs))
	} else if parsed_message == "np" {
		attrs, err := MpdCurrentSong()
		if err != nil {
			return err
		}
		irccon.Privmsg(reply, fmt.Sprintf("%v", attrs))
	} else if parsed_message == "play" {
		err := MpdPause(false)
		if err != nil {
			return err
		}
		irccon.Privmsg(reply, "Now playing")
	} else if parsed_message == "pause" {
		err := MpdPause(true)
		if err != nil {
			return err
		}
		irccon.Privmsg(reply, "Now paused")
	}
	return nil
}

func help(e *irc.Event, parsed_message, reply string) error {
	irccon.Privmsg(reply, commandprefix+"{say,silence,spacestate,alert,battery,vibrate,mpd}")
	return nil
}

func recurring(interval time.Duration, fu recFunc) {
	for true {
		time.Sleep(interval)
		err := fu()
		if err != nil {
			log.Println("recurring error:", err)
		}
	}
}

func privmsg(e *irc.Event) {
	user_or_chan := e.Arguments[0]

	// XXX: This is really ugly and will not work if we get a different nick
	// somehow
	if user_or_chan == ownnick {
		user_or_chan = e.Nick
	}

	msg := e.Arguments[1]

	nameref := false
	if strings.HasPrefix(msg, ownnick+": ") {
		msg = msg[len(ownnick+" "):]
		nameref = true
	}

	if strings.HasPrefix(msg, commandprefix) || nameref {
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
	go recurring(battery_check_time, func() error {
		percentage, err := BatteryPercentage()
		if err != nil {
			return err
		}
		if percentage < battery_low {
			if !battery_charging && time.Now().After(last_battery_warning.Add(battery_reminder_time)) {
				last_battery_warning = time.Now()

				msg := fmt.Sprintf("WARNING: battery percentage is less than %d: %d", battery_low, percentage)
				irccon.Privmsg("#techinc-spacephone", msg)
				Espeak(msg)
			}
		}
		return nil
	})
	go recurring(battery_charge_check_time, func() error {
		charge, err := BatteryCharging()
		if err != nil {
			return err
		}
		if charge != battery_charging {
			msg := "battery is now "
			if charge {
				irccon.Privmsg("#techinc-spacephone", msg+"charging")
				Espeak(msg + "charging")
			} else {
				irccon.Privmsg("#techinc-spacephone", msg+"discharging")
				Espeak(msg + "discharging")
			}
		}
		battery_charging = charge

		return nil
	})
	irccon.Loop()
}
