package main

import (
	"github.com/fhs/gompd/mpd"
)

func getClient() (*mpd.Client, error) {
	return mpd.Dial("tcp", "mpd.ti:6600")
}

func MpdConsume() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	err = c.Consume(false)
	if err != nil {
		return err
	}

	return nil
}

func MpdPause(pause bool) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	err = c.Pause(pause)
	if err != nil {
		return err
	}

	return nil
}

func MpdPing() error {
	c, err := getClient()
	if err != nil {
		return err
	}

	err = c.Ping()
	if err != nil {
		return err
	}

	return nil
}

// 2018/08/30 22:21:55 attrs: map[song:9 random:0 single:0 consume:0 playlist:99 playlistlength:19 mixrampdb:0.000000 volume:-1 repeat:0 nextsong:10 nextsongid:196 state:stop songid:195]
func MpdStatus() (mpd.Attrs, error) {
	c, err := getClient()
	if err != nil {
		return nil, err
	}

	return c.Status()
}

func MpdCurrentSong() (mpd.Attrs, error) {
	c, err := getClient()
	if err != nil {
		return nil, err
	}

	return c.CurrentSong()
}
