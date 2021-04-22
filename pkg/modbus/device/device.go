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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/configmap"
	"github.com/kubeedge/mappers-go/pkg/modbus/driver"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"k8s.io/klog"
)

var devices map[string]*globals.ModbusDev
var models map[string]common.DeviceModel
var protocols map[string]common.Protocol
var wg sync.WaitGroup

// setVisitor check if visitory is readonly, if not then set it.
func setVisitor(visitorConfig *configmap.ModbusVisitorConfig, twin *common.Twin, client *driver.ModbusClient) {
	if twin.PVisitor.PProperty.AccessMode == "ReadOnly" {
		klog.V(1).Info("Visit readonly register: ", visitorConfig.Offset)
		return
	}

	klog.V(2).Infof("Convert type: %s, value: %s ", twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	value, err := common.Convert(twin.PVisitor.PProperty.DataType, twin.Desired.Value)
	if err != nil {
		klog.Error(err)
		return
	}

	valueInt, _ := value.(int64)
	_, err = client.Set(visitorConfig.Register, visitorConfig.Offset, uint16(valueInt))
	if err != nil {
		klog.Error(err, visitorConfig)
		return
	}
}

// getDeviceID extract the device ID from Mqtt topic.
func getDeviceID(topic string) (id string) {
	re := regexp.MustCompile(`hw/events/device/(.+)/twin/update/delta`)
	return re.FindStringSubmatch(topic)[1]
}

// onMessage callback function of Mqtt subscribe message.
func onMessage(client mqtt.Client, message mqtt.Message) {
	klog.V(2).Info("Receive message", message.Topic())
	// Get device ID and get device instance
	id := getDeviceID(message.Topic())
	if id == "" {
		klog.Error("Wrong topic")
		return
	}
	klog.V(2).Info("Device id: ", id)

	var dev *globals.ModbusDev
	var ok bool
	if dev, ok = devices[id]; !ok {
		klog.Error("Device not exist")
		return
	}

	// Get twin map key as the propertyName
	var delta common.DeviceTwinDelta
	if err := json.Unmarshal(message.Payload(), &delta); err != nil {
		klog.Error("Unmarshal message failed: ", err)
		return
	}
	for twinName, twinValue := range delta.Delta {
		i := 0
		for i = 0; i < len(dev.Instance.Twins); i++ {
			if twinName == dev.Instance.Twins[i].PropertyName {
				break
			}
		}
		if i == len(dev.Instance.Twins) {
			klog.Error("Twin not found: ", twinName)
			continue
		}
		// Desired value is not changed.
		if dev.Instance.Twins[i].Desired.Value == twinValue {
			continue
		}
		dev.Instance.Twins[i].Desired.Value = twinValue
		var visitorConfig configmap.ModbusVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Error("Unmarshal visitor config failed")
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.ModbusClient)
	}
}

// isRS485Enabled is RS485 feature enabled for RTU.
func isRS485Enabled(customizedValue configmap.CustomizedValue) bool {
	isEnabled := false

	if len(customizedValue) != 0 {
		if value, ok := customizedValue["serialType"]; ok {
			if value == "RS485" {
				isEnabled = true
			}
		}
	}
	return isEnabled
}

// initModbus initialize modbus client
func initModbus(protocolConfig configmap.ModbusProtocolCommonConfig, slaveID int16) (client *driver.ModbusClient, err error) {
	if protocolConfig.COM.SerialPort != "" {
		modbusRTU := driver.ModbusRTU{SlaveID: byte(slaveID),
			SerialName:   protocolConfig.COM.SerialPort,
			BaudRate:     int(protocolConfig.COM.BaudRate),
			DataBits:     int(protocolConfig.COM.DataBits),
			StopBits:     int(protocolConfig.COM.StopBits),
			Parity:       protocolConfig.COM.Parity,
			RS485Enabled: isRS485Enabled(protocolConfig.CustomizedValues),
			Timeout:      5 * time.Second}
		client, _ = driver.NewClient(modbusRTU)
	} else if protocolConfig.TCP.IP != "" {
		modbusTCP := driver.ModbusTCP{
			SlaveID:  byte(slaveID),
			DeviceIP: protocolConfig.TCP.IP,
			TCPPort:  strconv.FormatInt(protocolConfig.TCP.Port, 10),
			Timeout:  5 * time.Second}
		client, _ = driver.NewClient(modbusTCP)
	} else {
		return nil, errors.New("No protocol found")
	}
	return client, nil
}

// initTwin initialize the timer to get twin value.
func initTwin(dev *globals.ModbusDev) {
	for i := 0; i < len(dev.Instance.Twins); i++ {
		var visitorConfig configmap.ModbusVisitorConfig
		if err := json.Unmarshal([]byte(dev.Instance.Twins[i].PVisitor.VisitorConfig), &visitorConfig); err != nil {
			klog.Error(err)
			continue
		}
		setVisitor(&visitorConfig, &dev.Instance.Twins[i], dev.ModbusClient)

		twinData := TwinData{Client: dev.ModbusClient,
			Name:               dev.Instance.Twins[i].PropertyName,
			Type:               dev.Instance.Twins[i].Desired.Metadatas.Type,
			RegisterType:       visitorConfig.Register,
			Address:            visitorConfig.Offset,
			Quantity:           uint16(visitorConfig.Limit),
			Topic:              fmt.Sprintf(common.TopicTwinUpdate, dev.Instance.ID),
			DeviceModel:        dev.Instance.Model,
			DeviceInstanceName: dev.Instance.Name}
		collectCycle := time.Duration(dev.Instance.Twins[i].PVisitor.CollectCycle)
		// If the collect cycle is not set, set it to 1 second.
		if collectCycle == 0 {
			collectCycle = 1 * time.Second
		}
		timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
		wg.Add(1)
		go func() {
			if err := timer.Start(); err != nil {
				wg.Done()
			}
		}()
	}
}

// initData initialize the timer to get data.
func initData(dev *globals.ModbusDev) {
	if dev.Instance.Model == "modbus-rtu-shutter-model" {
		wg.Add(1)
		go func() {
			twinData := TwinData{Client: dev.ModbusClient,
				Name:               "shutterAll",
				Type:               "float",
				RegisterType:       "HoldingRegister",
				Address:            0,
				Quantity:           21,
				Topic:              fmt.Sprintf(common.TopicDataUpdate, dev.Instance.ID),
				DeviceModel:        dev.Instance.Model,
				DeviceInstanceName: dev.Instance.Name,
			}
			collectCycle, _ := time.ParseDuration("1s")
			timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
			if err := timer.Start(); err != nil {
				wg.Done()
			}
		}()
	}
	if dev.Instance.Model == "modbus-rtu-snow-model" {
		wg.Add(1)
		go func() {
			twinData := TwinData{Client: dev.ModbusClient,
				Name:               "snow",
				Type:               "float",
				RegisterType:       "HoldingRegister",
				Address:            22,
				Quantity:           1,
				Topic:              fmt.Sprintf(common.TopicDataUpdate, dev.Instance.ID),
				DeviceModel:        dev.Instance.Model,
				DeviceInstanceName: dev.Instance.Name,
			}
			collectCycle, _ := time.ParseDuration("1s")
			timer := common.Timer{Function: twinData.Run, Duration: collectCycle, Times: 0}
			if err := timer.Start(); err != nil {
				wg.Done()
			}
		}()
	}
	wg.Wait()
}

// initSubscribeMqtt subscribe Mqtt topics.
func initSubscribeMqtt(instanceID string) error {
	topic := fmt.Sprintf(common.TopicTwinUpdateDelta, instanceID)
	klog.V(1).Info("Subscribe topic: ", topic)
	return globals.MqttClient.Subscribe(topic, onMessage)
}

// initGetStatus start timer to get device status and send to eventbus.
func initGetStatus(dev *globals.ModbusDev) {
	getStatus := GetStatus{Client: dev.ModbusClient,
		topic: fmt.Sprintf(common.TopicStateUpdate, dev.Instance.ID)}
	timer := common.Timer{Function: getStatus.Run, Duration: 1 * time.Second, Times: 0}
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer.Start()
	}()
}

// start start the device.
func start(dev *globals.ModbusDev) {
	var protocolConfig configmap.ModbusProtocolCommonConfig
	if err := json.Unmarshal([]byte(dev.Instance.PProtocol.ProtocolCommonConfig), &protocolConfig); err != nil {
		klog.Error(err)
		return
	}

	client, err := initModbus(protocolConfig, dev.Instance.PProtocol.ProtocolConfigs.SlaveID)
	if err != nil {
		klog.Error(err)
		return
	}
	dev.ModbusClient = client

	initTwin(dev)
	initData(dev)

	if err := initSubscribeMqtt(dev.Instance.ID); err != nil {
		klog.Error(err)
		return
	}

	initGetStatus(dev)
}

// DevInit initialize the device datas.
func DevInit(configmapPath string) error {
	devices = make(map[string]*globals.ModbusDev)
	models = make(map[string]common.DeviceModel)
	protocols = make(map[string]common.Protocol)
	return configmap.Parse(configmapPath, devices, models, protocols)
}

// DevStart start all devices.
func DevStart() {
	for id, dev := range devices {
		klog.V(4).Info("Dev: ", id, dev)
		start(dev)
	}
	wg.Wait()
}
