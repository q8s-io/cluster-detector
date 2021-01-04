package config

import (
	"github.com/BurntSushi/toml"
	"k8s.io/klog"
)

func Load(confPath string) error {
	// init runtime
	if _, err := toml.DecodeFile(confPath, &Config); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
