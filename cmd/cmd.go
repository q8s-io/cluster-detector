package cmd

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/infrastructure/config"
)

var (
	confPath       = flag.String("conf", "./configs/pro.toml", "The path of config.")
	argFrequency   = flag.Duration("frequency", 30*time.Second, "The resolution at which detector pushes source to sinks")
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip detector health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port resource health check listens on")
)

func Run() {
	quitChannel := make(chan struct{}, 0)

	// init config
	config.Init(*confPath)

	klog.InitFlags(nil)
	defer klog.Flush()

	if err := validateFlags(); err != nil {
		klog.Info(err)
	}
	flag.Parse()

	// Start watch resource of cluster
	if config.Config.EventsConfig.Enabled == true {
		go RunEventsWatch()
	}
	if config.Config.DeleteInspectionConfig.Enabled == true {
		go RunUnusedInspection()
	}
	if config.Config.NodeInspectionConfig.Enabled == true {
		go RunNodeInspection()
	}
	if config.Config.PodInspectionConfig.Enabled == true {
		go RunPodInspection()
	}

	go startHTTPServer()

	<-quitChannel
}

func validateFlags() error {
	var minFrequency = 5 * time.Second
	if *argHealthzPort > 65534 {
		return fmt.Errorf("invalid port supplied for healthz %d", *argHealthzPort)
	}
	if *argFrequency < minFrequency {
		return fmt.Errorf("frequency needs to be no less than %s, supplied %s", minFrequency, *argFrequency)
	}
	return nil
}

func startHTTPServer() {
	klog.Info("Starting eventer http service")
	klog.Info(http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil))
}
