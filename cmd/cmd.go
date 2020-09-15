package cmd

import (
	"flag"
	"fmt"
	"github.com/q8s-io/cluster-detector/configs"
	"github.com/q8s-io/cluster-detector/pkg/common/process"
	"k8s.io/klog"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	argFrequency = flag.Duration("frequency", 30*time.Second, "The resolution at which detector pushes source to sinks")
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip detector health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port resource health check listens on")

	confPath = flag.String("conf", "./configs/pro.toml", "The path of config.")
)

func Run() {
	quitChannel := make(chan struct{}, 0)
	// 初始化配置
	process.Init(*confPath)
	log.Println(configs.Config.PodInspectionConfig)

	klog.InitFlags(nil)
	defer klog.Flush()

	klog.Infof(strings.Join(os.Args, " "))
	if err := validateFlags(); err != nil {
		klog.Fatal(err)
	}
	flag.Parse()

	// Start watch resource of cluster
	if configs.Config.EventsConfig.Enabled == true {
		go RunEventsWatch()
	}
	if configs.Config.NodeInspectionConfig.Enabled == true {
		go RunNodeInspection()
	}
	if configs.Config.PodInspectionConfig.Enabled == true {
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
		return fmt.Errorf("frequency needs to be no less than %s, supplied %s", minFrequency,
			*argFrequency)
	}

	return nil
}

func startHTTPServer() {
	klog.Info("Starting eventer http service")
	klog.Fatal(http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil))
}