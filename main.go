package main

import (
	"flag"
	"fmt"
	"os"

	"MonophobiaServer/GameServer"

	log "github.com/sirupsen/logrus"
)

var FlagLogLevel, FlagIP string
var FlagPort int

func init() {
	flag.StringVar(&FlagLogLevel, "log", "info", "Set log level ( none, info, error, debug )")
	flag.StringVar(&FlagIP, "ip", "0.0.0.0", "Set IP for the server")
	flag.IntVar(&FlagPort, "port", 1338, "Set Port for the server")
}
func main() {

	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true,
		ForceColors:  true,
	})
	log.SetOutput(os.Stdout)

	lv, err := log.ParseLevel(FlagLogLevel)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		log.SetLevel(lv)
	}

	var server GameServer.GameServer = GameServer.GameServer{} //{IP: FlagIP, Port: int64(FlagPort)}
	server.SetAddress(FlagIP, FlagPort)
	server.GameVersion = "0.1.1"
	server.Start()

}
