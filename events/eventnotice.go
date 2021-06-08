package events

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//第一个输入nodename，第二个输入device
func EventNotice(args ...string) error {
	//新建clientset客户端
	clientset, err := NewClientset()
	if err != nil {
		fmt.Errorf("NewClientset err :%v", err)
		return err
	}
	//新建 event对象

	event := &apiv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mapper-imu-" + args[0],
		},
		InvolvedObject: apiv1.ObjectReference{
			Kind:      "pod",
			Name:      "mapper-imu-"+args[0]+"."+rand.String(10),
			Namespace: "default",
		},
		FirstTimestamp: metav1.Time{time.Now()},
		Message:        args[1]+"设备不可用",
		Reason:         args[1]+"disconnect",
		Type:           "Warning",
	}

	//创建event
	result, err := clientset.CoreV1().Events("default").Create(event)
	if err != nil {
		logrus.Error(err)
		return err
	}
	fmt.Printf("event:%v\n", result.Message)
	return nil
}
