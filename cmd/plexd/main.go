// plexd - broker daemon. Forwards traffic between input socket and subscribers.
// Copyright (C) Hexawolf <hexawolf@hexanet.dev>
// See LICENSE file for more info
package main

import (
	"fmt"
	"github.com/Hexawolf/plex"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"os/signal"
)

func logSetup(lc *viper.Viper) {
	if !lc.GetBool("LocalTime") {
		log.SetFlags(log.Flags() | log.LUTC)
	}
	if lc.GetBool("Debug") {
		log.SetFlags(log.Flags() | log.Lshortfile)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "<config>")
		os.Exit(1)
	}
	// Load config and watch for changes made by plexctl or user
	viper.SetDefault("Listen", ":18833")
	viper.SetConfigFile(os.Args[1])
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		return
	}
	viper.WatchConfig()
	logSetup(viper.Sub("Log"))

	mp, err := plex.NewPlex(viper.GetString("Listen"), uint16(viper.GetInt("Buffer")), nil)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		err = mp.ListenUDP()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// Subscribe routes
	routes := viper.GetStringSlice("Route")
	for _, v := range routes {
		raddr, err := net.ResolveUDPAddr("", v)
		if err != nil { log.Println(err); continue }
		conn, err := net.DialUDP("", nil, raddr)
		if err != nil { log.Println(err); continue }
		err = mp.SubscribeUDP(v, conn)
		if err != nil { log.Println(err); continue }
	}
	
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	mp.Close()
}
