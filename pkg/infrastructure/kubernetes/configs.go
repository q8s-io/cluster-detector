package kubernetes

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicClient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdAPI "k8s.io/client-go/tools/clientcmd/api"
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
		ClusterInfo: clientcmdAPI.Cluster{},
	}
	if len(uri.Scheme) != 0 && len(uri.Host) != 0 {
		kubeConfigOverride.ClusterInfo.Server = fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
	}
	opts := uri.Query()
	if len(opts["insecure"]) > 0 {
		insecure, err := strconv.ParseBool(opts["insecure"][0])
		if err != nil {
			return nil, err
		}
		kubeConfigOverride.ClusterInfo.InsecureSkipTLSVerify = insecure
	}
	return &kubeConfigOverride, nil
}

func GetKubeClientConfig(uri *url.URL) (*restClient.Config, error) {
	var (
		kubeConfig *restClient.Config
		err        error
	)
	opts := uri.Query()
	configOverrides, err := getConfigOverrides(uri)
	if err != nil {
		return nil, err
	}
	inClusterConfig := defaultInClusterConfig
	if len(opts["inClusterConfig"]) > 0 {
		inClusterConfig, err = strconv.ParseBool(opts["inClusterConfig"][0])
		if err != nil {
			return nil, err
		}
	}
	if inClusterConfig {
		kubeConfig, err = restClient.InClusterConfig()
		if err != nil {
			return nil, err
		}
		if configOverrides.ClusterInfo.Server != "" {
			kubeConfig.Host = configOverrides.ClusterInfo.Server
		}
		kubeConfig.GroupVersion = &schema.GroupVersion{Version: APIVersion}
		kubeConfig.Insecure = configOverrides.ClusterInfo.InsecureSkipTLSVerify
		if configOverrides.ClusterInfo.InsecureSkipTLSVerify {
			kubeConfig.TLSClientConfig.CAFile = ""
		}
	} else {
		authFile := ""
		if len(opts["auth"]) > 0 {
			authFile = opts["auth"][0]
		}
		if authFile != "" {
			// Load structured kubeconfig data from the given path.
			loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: authFile}
			loadedConfig, err := loader.Load()
			if err != nil {
				return nil, err
			}
			// Flatten the loaded data to a particular restclient.Config based on the current context.
			if kubeConfig, err = clientcmd.NewNonInteractiveClientConfig(
				*loadedConfig,
				loadedConfig.CurrentContext,
				&clientcmd.ConfigOverrides{},
				loader).ClientConfig(); err != nil {
				return nil, err
			}
		} else {
			kubeConfig = &restClient.Config{
				Host: configOverrides.ClusterInfo.Server,
				TLSClientConfig: restClient.TLSClientConfig{
					Insecure: configOverrides.ClusterInfo.InsecureSkipTLSVerify,
				},
			}
			kubeConfig.GroupVersion = &schema.GroupVersion{Version: APIVersion}
		}
	}
	if len(kubeConfig.Host) == 0 {
		return nil, fmt.Errorf("invalid kubernetes master url specified")
	}
	useServiceAccount := defaultUseServiceAccount
	if len(opts["useServiceAccount"]) >= 1 {
		useServiceAccount, err = strconv.ParseBool(opts["useServiceAccount"][0])
		if err != nil {
			return nil, err
		}
	}
	if useServiceAccount {
		// If a readable service account token exists, then use it
		if contents, err := ioutil.ReadFile(defaultServiceAccountFile); err == nil {
			kubeConfig.BearerToken = string(contents)
		}
	}
	kubeConfig.ContentType = "application/vnd.kubernetes.protobuf"
	return kubeConfig, nil
}

func GetKubernetesClient(uri *url.URL) (client kubernetes.Interface, err error) {
	/*kubeConfig, err := GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}*/
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Println(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}

func GetKubernetesDynamicClient(uri *url.URL) (client dynamicClient.Interface, err error) {
	kubeConfig, err := GetKubeClientConfig(uri)
	if err != nil {
		return nil, err
	}
	kubeClient, err := dynamicClient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}
