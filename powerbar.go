package main

import (
	"fmt"
	"net/http"
	"net/url"
)

const powerbar_host = "powerbar.ti"

//     curl -d state=$STATE http://$MLP_HOST:$MLP_PORT/$BAR/$SOCKET
func PowerSocketSet(socketname string, state bool) error {
	state_str := "On"
	if !state {
		state_str = "Off"
	}

	v := url.Values{}
	v.Set("state", state_str)

	resp, err := http.PostForm("http://"+powerbar_host+"/"+socketname, v)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid status code: " + resp.Status)
	}

	return nil
}
