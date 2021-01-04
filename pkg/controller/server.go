package controller

import (
	"net"
	"net/http"
	"strconv"

	"k8s.io/klog"
)

func StartHTTPServer(argHealthyIP string, argHealthyPort uint) {
	klog.Info("Starting event http service")
	klog.Info(http.ListenAndServe(net.JoinHostPort(argHealthyIP, strconv.Itoa(int(argHealthyPort))), nil))
}
