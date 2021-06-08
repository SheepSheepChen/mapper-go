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
	"github.com/kubeedge/mappers-go/events"
	"math"
	"strconv"
	"strings"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/driver"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"github.com/royeo/dingrobot"
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
	if err != nil {
		for i := 0; i <= 9; i++ {
			if td.Results, err = td.Client.Get(td.RegisterType, td.Address, td.Quantity); err == nil {
				break
			}
			if i == 9 {
				events.EventNotice(nodeName,td.DeviceInstanceName)
				klog.V(2).Infof("设备%v不可用", td.DeviceInstanceName)
				// 添加钉钉机器人提醒
				webhook := "https://oapi.dingtalk.com/robot/send?access_token=e79d127635b34ce6992539a2a2794978136947f4e7f33eaccc5394828d72f570"
				robot := dingrobot.NewRobot(webhook)
				content := fmt.Sprintf("设备%v不可用", td.DeviceInstanceName)
				atMobiles := []string{"18626860751"}
				robot.SendText(content, atMobiles, false)
				return fmt.Errorf("设备不可用")
			}
		}
	}
	s1 := strings.Replace(fmt.Sprintf("%v", td.Results), "[", "", -1)
	s2 := strings.Replace(s1, "]", "", -1)
	splitS2 := strings.Split(s2, " ")

	if strings.Contains(td.DeviceInstanceName, "shutter01") {
		var lux, co2, pressure, temperature, humidity float64
		// 湿度
		ss1 := splitS2[0]
		ss2 := splitS2[1]
		if ss1 != "255" && ss2 != "255" {
			humidity = Hex2Dec(ss1, ss2) / 10
		}
		// 温度
		ss3 := splitS2[2]
		ss4 := splitS2[3]
		if ss3 != "255" && ss4 != "255" {
			temperature = Hex2Dec(ss3, ss4) / 10
		}
		// 光强
		ss5 := splitS2[4]
		ss6 := splitS2[5]
		ss7 := splitS2[6]
		ss8 := splitS2[7]
		if ss5 != "255" && ss6 != "255" && ss7 != "255" && ss8 != "255" {
			lux = Hex2Dec(ss5, ss6, ss7, ss8)
		}
		// CO2
		ss11 := splitS2[14]
		ss12 := splitS2[15]
		if ss11 != "255" && ss12 != "255" {
			co2 = Hex2Dec(ss11, ss12)
		}
		// 大气压强
		ss13 := splitS2[22]
		ss14 := splitS2[23]
		if ss13 != "255" && ss14 != "255" {
			pressure = Hex2Dec(ss13, ss14) / 10
		}
		klog.V(2).Info("---------湿度-----------", humidity)
		klog.V(2).Info("---------温度-----------", temperature)
		klog.V(2).Info("---------光强-----------", lux)
		klog.V(2).Info("---------二氧化碳浓度-----------", co2)
		klog.V(2).Info("---------大气压强-----------", pressure)
		globals.FBClient.Publish(td.DeviceInstanceName, fmt.Sprintf(`{"node":"%s", "__name__":"%s", "humidity":%f, "temperature":%f, "lux":%f, "co2":%f, "pressure":%f, "state":"%s"}`, nodeName, td.DeviceModel, humidity, temperature, lux, co2, pressure, td.Client.GetStatus()))
	} else if strings.Contains(td.DeviceInstanceName, "shutter02") {
		var pm2point5, pm10, noise float64
		// 噪音
		ss9 := splitS2[8]
		ss10 := splitS2[9]
		if ss9 != "255" && ss10 != "255" {
			noise = Hex2Dec(ss9, ss10) / 10
		}
		// PM2.5
		ss15 := splitS2[40]
		ss16 := splitS2[41]
		if ss15 != "255" && ss16 != "255" {
			pm2point5 = Hex2Dec(ss15, ss16)
		}
		// PM 10
		ss17 := splitS2[42]
		ss18 := splitS2[43]
		if ss17 != "255" && ss18 != "255" {
			pm10 = Hex2Dec(ss17, ss18)
		}
		klog.V(2).Info("---------噪音-----------", noise)
		klog.V(2).Info("---------PM2.5-----------", pm2point5)
		klog.V(2).Info("---------PM10-----------", pm10)
		globals.FBClient.Publish(td.DeviceInstanceName, fmt.Sprintf(`{"node":"%s", "__name__":"%s", "nosie":%f, "pm2.5":%f, "pm10":%f, "state":"%s"}`, nodeName, td.DeviceModel, noise, pm2point5, pm10, td.Client.GetStatus()))
	} else if strings.Contains(td.DeviceInstanceName, "snow") {
		var snow float64
		ss1 := strings.Split(s2, " ")[0]
		ss2 := strings.Split(s2, " ")[1]
		if ss1 != "255" && ss2 != "255" {
			snow = Hex2Dec(ss1, ss2)
		}
		klog.V(2).Info("----------雨雪--------------", snow)
		globals.FBClient.Publish(td.DeviceInstanceName, fmt.Sprintf(`{"__name__":"%s","snow":%f,"node":"%s","state":"%s"}`, td.DeviceModel, snow, nodeName, td.Client.GetStatus()))

	}
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

func Hex2Dec(vals ...string) float64 {
	var result float64
	for index, val := range vals {
		floatVal, _ := strconv.ParseFloat(val, 64)
		result += math.Pow(256, float64(len(vals)-index)-1) * floatVal
	}
	return result
}
