package basicPrepare

import (
	"fmt"
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
)

func InitEnv(confPath string, argFrequency time.Duration, argHealthyPort uint) error {
	if err := validateFlags(argHealthyPort, argFrequency); err != nil {
		klog.Error(err)
		return err
	}

	if err := config.Load(confPath); err != nil {
		return err
	}
	return nil
}

func validateFlags(argHealthyPort uint, argFrequency time.Duration) error {
	if argHealthyPort > entity.MaxValidPort {
		return fmt.Errorf("invalid port supplied for healthy %d", argHealthyPort)
	}
	if argFrequency < entity.MinFrequency {
		return fmt.Errorf("frequency needs to be no less than %s, supplied %s", entity.MinFrequency, argFrequency)
	}
	return nil
}
