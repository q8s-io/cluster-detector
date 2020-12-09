package kafka

import "k8s.io/klog"

type GologAdapterLogger struct {
}

func (GologAdapterLogger) Print(v ...interface{}) {
	klog.Info(v...)
}

func (GologAdapterLogger) Printf(format string, args ...interface{}) {
	klog.Infof(format, args...)
}

func (GologAdapterLogger) Println(v ...interface{}) {
	klog.Infoln(v...)
}
