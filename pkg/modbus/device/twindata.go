/*
Copyright 2020 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package device

import (
	"fmt"
	"github.com/kubeedge/mappers-go/pkg/events"
	"strconv"
	"strings"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/driver"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"k8s.io/klog"
)

// TwinData is the timer structure for getting twin/data.
type TwinData struct {
	Client             *driver.ModbusClient
	Name               string
	Type               string
	RegisterType       string
	Address            uint16
	Quantity           uint16
	Results            []byte
	Topic              string
	DeviceModel        string
	DeviceInstanceName string
}

// Run timer function.
func (td *TwinData) Run() error {
	var (
		err error
	)
	td.Results, err = td.Client.Get(td.RegisterType, td.Address, td.Quantity)
	var nodeName string
	if len(strings.Split(td.DeviceInstanceName, "-")) == 3 && strings.Split(td.DeviceInstanceName, "-")[2] != "" {
		nodeName = strings.Split(td.DeviceInstanceName, "-")[2]
	}
	// 访问失败之后，继续访问，访问10次，如果10次全部失败说明设备或者串口不可用，直接retuen
	if err != nil || td.Results==nil{
		for i := 0; i <= 9; i++ {
			fmt.Println("result=nil")
			if td.Results, err = td.Client.Get(td.RegisterType, td.Address, td.Quantity); err == nil {
				break
			}
			if i == 9 {
				events.EventNotice(nodeName)
				return fmt.Errorf("IMU设备不可用")
			}
		}
	}
	s1 := strings.Replace(fmt.Sprintf("%v", td.Results), "[", "", -1)
	s2 := strings.Replace(s1, "]", "", -1)
	splitS2 := strings.Split(s2, " ")
	// acceleration
	ss1 := splitS2[0]
	ss2 := splitS2[1]
	ss3 := splitS2[2]
	ss4 := splitS2[3]
	ss5 := splitS2[4]
	ss6 := splitS2[5]
	axh, _ := strconv.Atoi(ss1)
	axl, _ := strconv.Atoi(ss2)
	ayh, _ := strconv.Atoi(ss3)
	ayl, _ := strconv.Atoi(ss4)
	azh, _ := strconv.Atoi(ss5)
	azl, _ := strconv.Atoi(ss6)
	k := 16.0
	accX := float64(axh<<8|axl) / 32768.0 * k
	accY := float64(ayh<<8|ayl) / 32768.0 * k
	accZ := float64(azh<<8|azl) / 32768.0 * k
	if accX >= k {
		accX -= 2 * k
	}
	if accY >= k {
		accY -= 2 * k
	}
	if accZ >= k {
		accZ -= 2 * k
	}
	klog.V(2).Info("---------accX-----------", accX)
	klog.V(2).Info("---------accY-----------", accY)
	klog.V(2).Info("---------accZ-----------", accZ)
	// angularVelocity

	ss7 := splitS2[6]
	ss8 := splitS2[7]
	ss9 := splitS2[8]
	ss10 := splitS2[9]
	ss11 := splitS2[10]
	ss12 := splitS2[11]
	wxh, _ := strconv.Atoi(ss7)
	wxl, _ := strconv.Atoi(ss8)
	wyh, _ := strconv.Atoi(ss9)
	wyl, _ := strconv.Atoi(ss10)
	wzh, _ := strconv.Atoi(ss11)
	wzl, _ := strconv.Atoi(ss12)
	k = 2000.0
	wX := float64(wxh<<8|wxl) / 32768.0 * k
	wY := float64(wyh<<8|wyl) / 32768.0 * k
	wZ := float64(wzh<<8|wzl) / 32768.0 * k
	if wX >= k {
		wX -= 2 * k
	}
	if wY >= k {
		wY -= 2 * k
	}
	if wZ >= k {
		wZ -= 2 * k
	}
	klog.V(2).Info("---------wX-----------", wX)
	klog.V(2).Info("---------wY-----------", wY)
	klog.V(2).Info("---------wZ-----------", wZ)
	// magnetic
	ss13 := splitS2[12]
	ss14 := splitS2[13]
	ss15 := splitS2[14]
	ss16 := splitS2[15]
	ss17 := splitS2[16]
	ss18 := splitS2[17]
	hxH, _ := strconv.Atoi(ss13)
	hxL, _ := strconv.Atoi(ss14)
	hyH, _ := strconv.Atoi(ss15)
	hyL, _ := strconv.Atoi(ss16)
	hzH, _ := strconv.Atoi(ss17)
	hzL, _ := strconv.Atoi(ss18)
	k = 1.0
	hX := float64(hxH<<8 | hxL)
	hY := float64(hyH<<8 | hyL)
	hZ := float64(hzH<<8 | hzL)
	if hX >= k {
		hX -= 2 * k
	}
	if hY >= k {
		hY -= 2 * k
	}
	if hZ >= k {
		hZ -= 2 * k
	}
	klog.V(2).Info("---------hX-----------", hX)
	klog.V(2).Info("---------hY-----------", hY)
	klog.V(2).Info("---------hZ-----------", hZ)
	// angular
	ss19 := splitS2[18]
	ss20 := splitS2[19]
	ss21 := splitS2[20]
	ss22 := splitS2[21]
	ss23 := splitS2[22]
	ss24 := splitS2[23]
	rollH, _ := strconv.Atoi(ss19)
	rollL, _ := strconv.Atoi(ss20)
	pitchH, _ := strconv.Atoi(ss21)
	pitchL, _ := strconv.Atoi(ss22)
	yawH, _ := strconv.Atoi(ss23)
	YawL, _ := strconv.Atoi(ss24)
	k = 180.0
	roll := float64(rollH<<8|rollL) / 32768.0 * k
	pitch := float64(pitchH<<8|pitchL) / 32768.0 * k
	yaw := float64(yawH<<8|YawL) / 32768.0 * k
	if roll >= k {
		roll -= 2 * k
	}
	if pitch >= k {
		pitch -= 2 * k
	}
	if yaw >= k {
		yaw -= 2 * k
	}
	klog.V(2).Info("---------roll-----------", roll)
	klog.V(2).Info("---------pitch-----------", pitch)
	klog.V(2).Info("---------yaw-----------", yaw)

	//element
	ss25 := splitS2[len(splitS2)-8]
	ss26 := splitS2[len(splitS2)-7]
	ss27 := splitS2[len(splitS2)-6]
	ss28 := splitS2[len(splitS2)-5]
	ss29 := splitS2[len(splitS2)-4]
	ss30 := splitS2[len(splitS2)-3]
	ss31 := splitS2[len(splitS2)-2]
	ss32 := splitS2[len(splitS2)-1]
	q0H, _ := strconv.Atoi(ss25)
	q0L, _ := strconv.Atoi(ss26)
	q1H, _ := strconv.Atoi(ss27)
	q1L, _ := strconv.Atoi(ss28)
	q2H, _ := strconv.Atoi(ss29)
	q2L, _ := strconv.Atoi(ss30)
	q3H, _ := strconv.Atoi(ss31)
	q3L, _ := strconv.Atoi(ss32)
	k = 1.0
	q0 := float64(q0H<<8|q0L) / 32768.0
	q1 := float64(q1H<<8|q1L) / 32768.0
	q2 := float64(q2H<<8|q2L) / 32768.0
	q3 := float64(q3H<<8|q3L) / 32768.0
	if q0 >= k {
		q0 -= 2 * k
	}
	if q1 >= k {
		q1 -= 2 * k
	}
	if q2 >= k {
		q2 -= 2 * k
	}
	if q3 >= k {
		q3 -= 2 * k
	}
	klog.V(2).Info("---------q0-----------", q0)
	klog.V(2).Info("---------q1-----------", q1)
	klog.V(2).Info("---------q2-----------", q2)
	klog.V(2).Info("---------q3-----------", q3)
	globals.FBClient.Publish(td.DeviceInstanceName, fmt.Sprintf(`{"node":"%s","__name__":"%s","accX":%f,"accY":%f,"accZ":%f,"wX":%f,"wY":%f,"wZ":%f,"Hx":%f,"Hy":%f,"Hz":%f,"Roll":%f,"Pitch":%f,"Yaw":%f,"Q0":%f,"Q1":%f,"Q2":%f,"Q3":%f,"state":"%s","device":"imu","topicKey":"%s"}`, nodeName, td.DeviceInstanceName, accX, accY, accZ, wX, wY, wZ, hX, hY, hZ, roll, pitch, yaw, q0, q1, q2, q3, td.Client.GetStatus(), globals.TopicKey))
	// construct payload
	var payload []byte
	if strings.Contains(td.Topic, "$hw") {
		if payload, err = common.CreateMessageTwinUpdate(td.Name, td.Type, strconv.Itoa(int(td.Results[0]))); err != nil {
			klog.Error("Create message twin update failed")
			return err
		}
	} else {
		if payload, err = common.CreateMessageData(td.Name, td.Type, strconv.Itoa(int(td.Results[0]))); err != nil {
			klog.Error("Create message data failed")
			return err
		}
	}
	if err = globals.MqttClient.Publish(td.Topic, payload); err != nil {
		klog.Error(err)
	}

	klog.V(2).Infof("Update value: %s, topic: %s", strconv.Itoa(int(td.Results[0])), td.Topic)
	return err
}
