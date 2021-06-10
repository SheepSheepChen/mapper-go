package events

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"time"

	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)
var (
	Clientset   *kubernetes.Clientset
	err error
)
func init() {
	//新建clientset客户端
	Clientset, err = NewClientset()
	if err != nil {
		fmt.Errorf("NewClientset err :%v", err)
	}

}
//第一个输入nodename，第二个输入device
func EventNotice(args ...string) error {

	deploy, err := Clientset.AppsV1().Deployments("default").Get("modbus-mapper-common-"+args[0], metav1.GetOptions{})
	if err != nil {
		logrus.Error(err)
		return err
	}

	//新建 event对象
	event := &apiv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mapper-common-" + args[0]+"."+rand.String(10),
		},
		InvolvedObject: apiv1.ObjectReference{
			Kind:      "deployment",
			Name:      "mapper-common-"+args[0],
			Namespace: "default",
			UID:       deploy.ObjectMeta.UID,
		},
		FirstTimestamp: metav1.Time{time.Now()},
		Message:        args[1]+"设备不可用",
		Reason:         args[1]+"disconnect",
		Type:           "Warning",
	}

	//创建event
	result, err := Clientset.CoreV1().Events("default").Create(event)
	if err != nil {
		logrus.Error(err)
		return err
	}
	fmt.Printf("event:%v\n", result.Message)
	return nil
}
