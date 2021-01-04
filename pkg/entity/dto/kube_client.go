package dto

import (
	"fmt"
	"io/ioutil"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
)

func ConvertKubeCfg(urlInfo *entity.UrlInfo) (*rest.Config, error) {
	var kubeConfig *rest.Config
	var err error

	if urlInfo.InClusterConfig {
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		if urlInfo.Server != "" {
			kubeConfig.Host = urlInfo.Server
		}
		kubeConfig.GroupVersion = &schema.GroupVersion{Version: entity.APIVersion}
		kubeConfig.Insecure = urlInfo.Insecure
		if urlInfo.Insecure {
			kubeConfig.TLSClientConfig.CAFile = ""
		}
	} else {
		if urlInfo.AuthFile != "" {
			// Load structured kubeconfig data from the given path.
			loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: urlInfo.AuthFile}
			loadedConfig, err := loader.Load()
			if err != nil {
				klog.Error(err)
				return nil, err
			}
			// Flatten the loaded data to a particular restclient.Config based on the current context.
			if kubeConfig, err = clientcmd.NewNonInteractiveClientConfig(
				*loadedConfig,
				loadedConfig.CurrentContext,
				&clientcmd.ConfigOverrides{},
				loader).ClientConfig(); err != nil {
				klog.Error(err)
				return nil, err
			}
		} else {
			kubeConfig = &rest.Config{
				Host: urlInfo.Server,
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: urlInfo.Insecure,
				},
			}
			kubeConfig.GroupVersion = &schema.GroupVersion{Version: entity.APIVersion}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes master url specified")
	}
	if urlInfo.UseServiceAccount {
		// If a readable service account token exists, then use it
		if contents, err := ioutil.ReadFile(entity.DefaultServiceAccountFile); err == nil {
			kubeConfig.BearerToken = string(contents)
		}
	}
	kubeConfig.ContentType = entity.ContentType
	return kubeConfig, nil
}
