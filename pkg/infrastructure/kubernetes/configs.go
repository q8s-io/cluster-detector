package kubernetes

import (
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"

	"github.com/q8s-io/cluster-detector/pkg/entity"
	"github.com/q8s-io/cluster-detector/pkg/entity/dto"
)

const (
	APIVersion                = "v1"
	defaultUseServiceAccount  = false
	defaultServiceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultInClusterConfig    = true
	kubeConfig                = "./configs/kubeconfig"
)

func getConfigOverrides(uri *url.URL) (*clientcmd.ConfigOverrides, error) {
	kubeConfigOverride := clientcmd.ConfigOverrides{
		ClusterInfo: api.Cluster{},
	}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}
	opts := uri.Query()
	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}
	return &kubeConfigOverride, nil
}

func GetInfoFromUrl(uri string) (*entity.UrlInfo, error) {
	// init default value
	var urlInfo = entity.UrlInfo{
		InClusterConfig:   defaultInClusterConfig,
		Insecure:          false,
		UseServiceAccount: defaultUseServiceAccount,
		Server:            "",
	}
	requestURI, parseErr := url.ParseRequestURI(uri)
	if parseErr != nil {
		klog.Error(parseErr)
		return nil, parseErr
	}
	// kubeConfigOverride := clientcmd.ConfigOverrides{
	// 	ClusterInfo: api.Cluster{},
	// }
	if len(requestURI.Scheme) != 0 && len(requestURI.Host) != 0 {
		urlInfo.Server = fmt.Sprintf("%s://%s", requestURI.Scheme, requestURI.Host)
	}
	opts := requestURI.Query()

	if len(opts["inClusterConfig"]) > 0 {
		inClusterConfig, err := strconv.ParseBool(opts["inClusterConfig"][0])
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		urlInfo.InClusterConfig = inClusterConfig
	}

	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		// kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
		urlInfo.Insecure = insecure
	}

	if len(opts["useServiceAccount"]) >= 1 {
		useServiceAccount, err := strconv.ParseBool(opts["useServiceAccount"][0])
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		urlInfo.UseServiceAccount = useServiceAccount
	}

	if !urlInfo.InClusterConfig {
		authFile := ""
		if len(opts["auth"]) > 0 {
			authFile = opts["auth"][0]
		}
		urlInfo.AuthFile = authFile
	}

	return &urlInfo, nil

}

func GetKubeClientConfig(uri string) (*rest.Config, error) {
	urlInfo, err := GetInfoFromUrl(uri)
	if err != nil {
		return nil, err
	}
	return dto.ConvertKubeCfg(urlInfo)
}

func GetKubernetesClient(kubeConfig *rest.Config) (kubernetes.Interface, error) {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return kubeClient, nil
}

func GetKubernetesDynamicClient(uri string) (client dynamic.Interface, err error) {
	kubeConfig, err := GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}
	kubeClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}
