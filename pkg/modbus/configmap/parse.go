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

package configmap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/kubeedge/mappers-go/pkg/common"
	"github.com/kubeedge/mappers-go/pkg/modbus/globals"
	"k8s.io/klog"
)

// Parse parse the configmap.
func Parse(path string,
	devices map[string]*globals.ModbusDev,
	dms map[string]common.DeviceModel,
	protocols map[string]common.Protocol) error {
	var deviceProfile common.DeviceProfile

	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonFile, &deviceProfile); err != nil {
		return err
	}
	for i := 0; i <= len(deviceProfile.DeviceInstances)-1; i++ {
		for j := 0; j <= len(deviceProfile.Protocols)-1; j++ {
			var protocolFound bool
			if deviceProfile.Protocols[j].Protocol != globals.Modbus {
				continue
			}
			if deviceProfile.DeviceInstances[i].ProtocolName == deviceProfile.Protocols[j].Name {
				protocolFound = true
				deviceProfile.DeviceInstances[i].PProtocol = deviceProfile.Protocols[j]
				break
			}
			if !protocolFound && j == len(deviceProfile.Protocols)-1 {
				return fmt.Errorf("Protocol not found")
			}
		}
		for k := 0; k <= len(deviceProfile.DeviceInstances[i].PropertyVisitors)-1; k++ {
			modelName := deviceProfile.DeviceInstances[i].PropertyVisitors[k].ModelName
			propertyName := deviceProfile.DeviceInstances[i].PropertyVisitors[k].PropertyName
			for l := 0; l <= len(deviceProfile.DeviceModels)-1; l++ {
				var deviceModelFound bool
				if modelName == deviceProfile.DeviceModels[l].Name {
					deviceModelFound = true
					for m := 0; m <= len(deviceProfile.DeviceModels[l].Properties)-1; m++ {
						var propertyFound bool
						if propertyName == deviceProfile.DeviceModels[l].Properties[m].Name {
							propertyFound = true
							deviceProfile.DeviceInstances[i].PropertyVisitors[k].PProperty = deviceProfile.DeviceModels[l].Properties[m]
							break
						}
						if !propertyFound && m == len(deviceProfile.DeviceModels[l].Properties)-1 {
							return fmt.Errorf("Property not found")
						}
					}
					break
				}
				if !deviceModelFound && l == len(deviceProfile.DeviceModels)-1 {
					return fmt.Errorf("Device model not found")
				}
			}
		}
		for n := 0; n <= len(deviceProfile.DeviceInstances[i].Twins)-1; n++ {
			propertyName := deviceProfile.DeviceInstances[i].Twins[n].PropertyName
			for o := 0; o <= len(deviceProfile.DeviceInstances[i].PropertyVisitors)-1; o++ {
				var propertyNameFound bool
				if propertyName == deviceProfile.DeviceInstances[i].PropertyVisitors[o].PropertyName {
					propertyNameFound = true
					deviceProfile.DeviceInstances[i].Twins[n].PVisitor = &deviceProfile.DeviceInstances[i].PropertyVisitors[o]
					break
				}
				if !propertyNameFound && o == len(deviceProfile.DeviceInstances[i].PropertyVisitors)-1 {
					return fmt.Errorf("PropertyVisitor not found")
				}
			}
		}

		deviceProfile.DeviceInstances[i].Datas.Properties = deviceProfile.DeviceInstances[i].Properties
		deviceProfile.DeviceInstances[i].Datas.Topic = deviceProfile.DeviceInstances[i].Topic

		for p := 0; p <= len(deviceProfile.DeviceInstances[i].Datas.Properties)-1; p++ {
			propertyName := deviceProfile.DeviceInstances[i].Datas.Properties[p].PropertyName
			for q := 0; q <= len(deviceProfile.DeviceInstances[i].PropertyVisitors)-1; q++ {
				var PropertyNameFound bool
				if propertyName == deviceProfile.DeviceInstances[i].PropertyVisitors[q].PropertyName {
					PropertyNameFound = true
					deviceProfile.DeviceInstances[i].Datas.Properties[p].PVisitor = &deviceProfile.DeviceInstances[i].PropertyVisitors[q]
					break
				}
				if !PropertyNameFound && q == len(deviceProfile.DeviceInstances[i].PropertyVisitors)-1 {
					return fmt.Errorf("PropertyVisitor not found")
				}
			}
		}
		devices[deviceProfile.DeviceInstances[i].ID] = new(globals.ModbusDev)
		devices[deviceProfile.DeviceInstances[i].ID].Instance = deviceProfile.DeviceInstances[i]
		klog.V(4).Info("Instance: ", deviceProfile.DeviceInstances[i].ID, deviceProfile.DeviceInstances[i])
	}
	for _, deviceModel := range deviceProfile.DeviceModels {
		dms[deviceModel.Name] = deviceModel
	}

	for _, protocol := range deviceProfile.Protocols {
		protocols[protocol.Name] = protocol
	}
	return nil

}
