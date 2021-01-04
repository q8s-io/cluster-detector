package entity

import (
	"time"
)

const (
	MinFrequency              = 5 * time.Second
	MaxValidPort              = 65535
	APIVersion                = "v1"
	DefaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	ContentType               = "application/vnd.kubernetes.protobuf"
)

type UrlInfo struct {
	InClusterConfig   bool
	Insecure          bool
	UseServiceAccount bool
	Server            string
	AuthFile          string
}
