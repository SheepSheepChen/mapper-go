package events

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
)

var (
	Kubeconfig *string
)

//func init() {
//	// init kubeconfig
//	if home := homedir.HomeDir(); home != "" {
//		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
//	} else {
//		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
//	}
//	flag.Parse()
//}


func NewClientset() (*kubernetes.Clientset, error) {
	//*Kubeconfig ="/root/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		return nil, err
		fmt.Println(err)
	}
	// 实例化clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	return clientset,err
}
