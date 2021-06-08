package events

import (
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"


)

var (
	Kubeconfig       *string
)


func init() {
	// init kubeconfig
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

func NewClientset() (*kubernetes.Clientset, error){
	config, err := clientcmd.BuildConfigFromFlags("", *Kubeconfig)
	if err != nil {
		return nil, err
	}
	// 实例化clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	return clientset,err
}
