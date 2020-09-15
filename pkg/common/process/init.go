package process

import (
	"github.com/q8s-io/cluster-detector/configs"
	"github.com/q8s-io/cluster-detector/pkg/common/qlog"
	"log"

	"github.com/BurntSushi/toml"
)

func Init(confPath string) {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &configs.Config); err != nil {
		log.Println(err)
		return
	}
	// init log
	qlog.InitLog()
}
