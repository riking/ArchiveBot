package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/riking/ArchiveBot/listener/discord"
	"io/ioutil"
	"os"
)

var flagShardID = flag.Int("shard", -1, "shard ID of this bot")
var flagNoHttp = flag.Bool("nohttp", true, "TODO")

type Config struct {
	Discord  json.RawMessage `json:"Discord"`
	Uploader *struct{}
}

func main() {
	flag.Parse()

	confBytes, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Println("Please copy config.yml.example to config.yml and fill out the values")
		os.Exit(2)

		return
	}
	var conf Config
	err = yaml.Unmarshal(confBytes, &conf)

	conf.Uploader = new(struct{}) // TODO
	var db *sql.DB = nil

	l, err := discord.NewListener(conf.Discord, db, conf.Uploader)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	err = l.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	if !*flagNoHttp {
		panic("unimplemented!")
	} else {
		select {}
	}
}
