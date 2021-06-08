package events

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EventNotice(string2 string) error {
	//新建clientset客户端
	clientset, err := NewClientset()
	if err != nil {
		fmt.Errorf("NewClientset err :%v", err)
		return err
	}
	//新建 event对象

	event := &apiv1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mapper-imu-" + string2,
		},
		InvolvedObject: apiv1.ObjectReference{
			Kind:      "pod",
			Name:      "mapper-imu-"+string2+"."+rand.String(10),
			Namespace: "default",
		},
		FirstTimestamp: metav1.Time{time.Now()},
		Message:        "imu 设备不可用",
		Reason:         "imu disconnect",
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
