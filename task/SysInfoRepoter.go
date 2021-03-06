package task

import (
	"encoding/json"
	"errors"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_agent/util"
	"io/ioutil"
	"net"
	"os"
	"path"
	"reflect"
	"strings"
)

const (
	machineTypePhysical  = "physical"
	machineTypeVirtual   = "virtual"
	machineTypeContainer = "container"
)

type SysInfoReporter struct {
	Api *returndata.DynamicApi
}

func (t *SysInfoReporter) Run() error {
	cacheFile := path.Dir(config.B.CredentialFile) + "/" + "sys_info.json"
	cachedParam := &api.SysInfoParam{}
	util.ParseJsonFile(cacheFile, cachedParam)

	params := &api.SysInfoParam{}
	params.Cpu = t.getCpu()
	params.Hostname = t.getHostname()
	params.Ip = t.getIp()
	params.MachineType = t.getMachineType()
	params.Memory = t.getMemory()
	bytes, _ := json.Marshal(params)
	if reflect.DeepEqual(params, cachedParam) {
		return nil
	}

	util.Logger.Info("Calling sys info report api")
	reportResult, err := util.ApiClient.CallApi(t.Api.Name, t.Api.Version, params)
	if err != nil {
		errMsg := "Failed to call sys info report api: " + t.Api.Name + ": " + t.Api.Version + ": " + err.Error()
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	if reportResult.Success == false {
		errMsg := "api return error: " + reportResult.ErrorCode + ": " + reportResult.ErrorMessage
		util.Logger.Error(errMsg)
		return errors.New(errMsg)
	}
	util.Logger.Info("Succeed to call sys info report api.")
	ioutil.WriteFile(cacheFile, bytes, 0600)
	return nil
}

func (t *SysInfoReporter) getCpu() int {
	return 1
}

func (t *SysInfoReporter) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		util.Logger.Error("Failed to get hostname: " + err.Error())
		return ""
	}
	return hostname
}

func (t *SysInfoReporter) getIp() []string {
	var ipv4List []string
	interfaces, err := net.Interfaces()
	if err != nil {
		util.Logger.Error("Failed to get net address list: " + err.Error())
		return ipv4List
	}
	if len(interfaces) < 2 {
		errMsg := "task: amount of net address is less than 2"
		util.Logger.Error(errMsg)
		return ipv4List
	}
	for _, netInterface := range interfaces {
		if netInterface.Flags&net.FlagBroadcast == 0 {
			continue
		}
		addressList, _ := netInterface.Addrs()
		for _, address := range addressList {
			ipv4List = append(ipv4List, strings.Split(address.String(), "/")[0])
			break
		}
	}
	return ipv4List
}

func (t *SysInfoReporter) getMachineType() string {
	return machineTypePhysical
}

func (t *SysInfoReporter) getMemory() int {
	return 1024
}
