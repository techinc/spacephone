package main

import (
	"os/exec"
	"strconv"
	"strings"
)

func BatteryPercentage() (int64, error) {
	//c := exec.Command("/bin/sh", "-c", "cat /sys/class/power_supply/BAT0/uevent | grep 'POWER_SUPPLY_CAPACITY=' | egrep -o '[0-9]+'")
	c := exec.Command("/bin/sh", "-c", "lshal | egrep -o 'percentage = [0-9]+' | egrep -o '[0-9]+'")

	out, err := c.Output()
	if err != nil {
		return 0, err
	}

	output := string(out)
	output = strings.Replace(output, "\n", "", -1)
	val, err := strconv.ParseInt(output, 10, 64)

	return val, err
}
