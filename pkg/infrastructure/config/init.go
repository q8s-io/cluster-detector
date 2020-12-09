package config

import (
	"log"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/qlog"

	"github.com/BurntSushi/toml"
)

func Init(confPath string) {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &Config); err != nil {
		log.Println(err)
		return
	}
	// init log
	qlog.InitLog()
}
