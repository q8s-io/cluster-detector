package cmd

import (
	"flag"
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/controller"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/basicPrepare"
	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
)

var (
	confPath       = flag.String("conf", "./configs/pro.toml", "The path of config.")
	argFrequency   = flag.Duration("frequency", 30*time.Second, "The resolution at which detector pushes source to sinks")
	argHealthyIP   = flag.String("healthy-ip", "0.0.0.0", "ip detector health check service uses")
	argHealthyPort = flag.Uint("healthy-port", 8084, "port resource health check listens on")
)

func Run() {
	flag.Parse()
	klog.InitFlags(nil)
	defer klog.Flush()
	// init config
	if err := basicPrepare.InitEnv(*confPath, *argFrequency, *argHealthyPort); err != nil {
		return
	}
	runApps()
}

func runApps() {
	cfg := config.Config
	quitChannel := make(chan struct{}, 0)
	defer close(quitChannel)

	controller.RunKafka()
	// Start watch resource of cluster
	if cfg.EventsConfig.Enabled == true {
		controller.RunEventsWatch()
	}
	// if cfg.DeleteInspectionConfig.Enabled == true {
	// 	go controller.RunUnusedInspection()
	// }
	// if config.Config.NodeInspectionConfig.Enabled == true {
	// 	go controller.RunNodeInspection()
	// }
	// if cfg.PodInspectionConfig.Enabled == true {
	// 	go controller.RunPodInspection()
	// }
	go controller.Sink()
	go controller.StartHTTPServer(*argHealthyIP, *argHealthyPort)
	<-quitChannel
}
