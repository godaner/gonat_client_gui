package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
)

var Remote_ip, Server_ip, Crypt, CryptKey string
var Remote_port int
var Debug bool
var Config_path string

type config struct {
	Debug      bool   `json:"debug"`
	RemotePort int    `json:"remote_port"`
	RemoteIp   string `json:"remote_ip"`
	ServerIp   string `json:"server_ip"`
	Crypt      string `json:"crypt"`
	CryptKey   string `json:"crypt_key"`
}

func Load() {

	flag.Parse()

	if Config_path != "" {
		file, err := os.Open(Config_path)
		if err != nil {
			return
		}

		b, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		cfg := config{}
		err = json.Unmarshal(b, &cfg)
		if err != nil {
			panic(err)
		}

		Remote_ip = cfg.RemoteIp
		Remote_port = cfg.RemotePort
		Server_ip = cfg.ServerIp
		Debug = cfg.Debug
		Crypt = cfg.Crypt
		CryptKey = cfg.CryptKey

	} else {

	}
}
