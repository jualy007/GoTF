package models

import (
	"fmt"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/jualy007/GoTF/blockchain/lighting"
	"github.com/jualy007/GoTF/config"
)

var Livenode = arraylist.New()

func init() {
	Healthcheck()
}

func Healthcheck() {
	Livenode.Clear()

	for key, value := range config.Cfg.Lnds {
		_, err := lighting.NewAdapter(value)

		if err != nil {
			fmt.Println("Lnd Node %v is not alive", key)
		} else {
			Livenode.Add(key)
		}
	}
}
