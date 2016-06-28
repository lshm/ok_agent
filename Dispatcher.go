package main

import (
	//go builtin pkg
	"encoding/json"
	"io/ioutil"
	"os"

	//local pkg
	"github.com/OpsKitchen/ok_agent/model/config"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk"
	"github.com/OpsKitchen/ok_agent/model/api"
	"github.com/OpsKitchen/ok_api_sdk_go/sdk/model"
	"github.com/OpsKitchen/ok_agent/model/api/returndata"
)

type Dispatcher struct {
	ApiClient      *sdk.Client
	ApiParam       *api.RequestParam
	BaseConfigFile string
	Config         *config.Base
	Credential     *config.Credential
	DynamicApiList []returndata.Api
}

func (dispatcher *Dispatcher) Dispatch() {
	dispatcher.parseBaseConfig()
	dispatcher.parseCredentialConfig()
	dispatcher.prepareApiClient()
	dispatcher.prepareApiParam()
	dispatcher.prepareDynamicApiList()

	for _, api := range dispatcher.DynamicApiList {
		debugLogger.Debug(api)
	}
}

func (dispatcher *Dispatcher) parseBaseConfig() {
	var baseConfig *config.Base
	var err error
	var jsonBytes []byte

	debugLogger.Info("base config file: ", dispatcher.BaseConfigFile)
	if _, err := os.Stat(dispatcher.BaseConfigFile); os.IsNotExist(err) {
		debugLogger.Fatal("base config file not found")
	}

	jsonBytes, err = ioutil.ReadFile(dispatcher.BaseConfigFile)
	if err != nil {
		debugLogger.Fatal("base config file not readable")
	}

	err = json.Unmarshal(jsonBytes, &baseConfig)
	if err != nil {
		debugLogger.Fatal("json decode failed: ", err.Error())
	}

	dispatcher.Config = baseConfig
}

func (dispatcher *Dispatcher) parseCredentialConfig() {
	var credentialConfig *config.Credential
	var err error
	var jsonBytes []byte

	debugLogger.Info("credential config file: ", dispatcher.Config.CredentialFile)
	if _, err := os.Stat(dispatcher.Config.CredentialFile); os.IsNotExist(err) {
		debugLogger.Fatal("credential config file not found")
	}

	jsonBytes, err = ioutil.ReadFile(dispatcher.Config.CredentialFile)
	if err != nil {
		debugLogger.Fatal("credential config file not readable")
	}

	err = json.Unmarshal(jsonBytes, &credentialConfig)
	if err != nil {
		debugLogger.Fatal("json decode failed: ", err.Error())
	}

	dispatcher.Credential = credentialConfig
}

func (dispatcher *Dispatcher) prepareApiClient() {
	var client *sdk.Client = sdk.NewClient()
	//inject logger
	sdk.SetDefaultLogger(debugLogger)

	//init config
	client.RequestBuilder.Config.SetAppMarketIdValue("1").SetAppVersionValue(
		dispatcher.Config.AgentVersion).SetGatewayHost(
		dispatcher.Config.GatewayHost).SetDisableSSL(dispatcher.Config.DisableSSL)

	//init credential
	client.RequestBuilder.Credential.SetAppKey(dispatcher.Credential.AppKey).SetSecret(
		dispatcher.Credential.Secret)
	dispatcher.ApiClient = client
}

func (dispatcher *Dispatcher) prepareApiParam() {
	dispatcher.ApiParam = &api.RequestParam{}
	dispatcher.ApiParam.ServerUniqueName = dispatcher.Credential.ServerUniqueName
}

func (dispatcher *Dispatcher) prepareDynamicApiList() {
	var entranceApiResult *model.ApiResult
	var err error

	entranceApiResult, err = dispatcher.ApiClient.CallApi(dispatcher.Config.EntranceApiName,
		dispatcher.Config.EntranceApiVersion, dispatcher.ApiParam)
	if err != nil {
		debugLogger.Fatal("call entrance api failed", err.Error())
	}

	if entranceApiResult.Success == false {
		debugLogger.Fatal("entrance api return error: ", entranceApiResult.ErrorCode, entranceApiResult.ErrorMessage)
	}

	json.Unmarshal(entranceApiResult.DataBytes, &dispatcher.DynamicApiList)
	if len(dispatcher.DynamicApiList) == 0 {
		debugLogger.Fatal("entrance api return empty api list")
	}
}
