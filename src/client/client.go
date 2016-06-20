package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"config"

	"github.com/bitly/go-simplejson"
	"logger"

    "okAgent/command"
    "okAgent/file"
    "okAgent/augeas"
    "okAgent/http"
    "flag"
)

/**/
func cmdHttp(jsonStr string, apiName string,detailBool bool) {
	bools := checkJson(jsonStr, apiName)
	if !bools {
		return
	}

	js, _ := simplejson.NewJson([]byte(jsonStr))
	if apiName == config.BaseConfig.BASE_API_NAME {

		arr_new, err := js.Get("data").Array()
		if err != nil {
			logger.Info("decode error: get data array failed!")
			return
		}
		// Cycle apilist 
		for _, v := range arr_new {
			api_name := v.(string)
			jsonStr, err := agentHttp.DoHttpRequest("POST", api_name)
			if err != nil {
                return
			}
			if detailBool {
				fileName := config.BaseConfig.TEMP_JSON_FILE_PATH + config.BaseConfig.TEMP_JSON_EX + api_name 
				jsonMap := map[string]interface{}{"filePath": fileName + ".json", "fileContent": jsonStr, "fileType": "file"}
				//Cycle store temporary json files
				agentFile.DoFile(jsonMap)
			}
			doJsonFile(api_name,jsonStr)
		}
	}

}

func checkJson(jsonStr string, api_name string) bool {
	js, err := simplejson.NewJson([]byte(jsonStr))
	if err != nil {
		logger.Error(api_name+" json format error")
		exitAgent(api_name+" Json format error")
		return false
	}
	result, err := js.Get("success").Bool()
	if err != nil {
		logger.Info("Decode error: Get int failed!")
		exitAgent("Decode error: Get int failed!")
		return true
	}
	if result == false {
		// errorCode, _ := js.Get("errorCode").Int()
		errorMessage, _ := js.Get("errorMessage").String()
		logger.Info("Error Message: " + errorMessage)
		exitAgent("Error Message: " + errorMessage)
		return false
	}
	return true
}

/*read file*/
func readJsonFile(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		exitAgent("Read Json File failed ")
		return ""
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func doJsonFile(api_name string ,jsonStr string) {
	bools := checkJson(jsonStr, api_name)
	if bools {
		js, _ := simplejson.NewJson([]byte(jsonStr))
		arr_new, err := js.Get("data").Array()
		if err != nil {
			return
		}
		
		for _, v := range arr_new {
			cmdMap := v.(map[string]interface{})
			category := cmdMap["category"].(string)
			switch category {
			case "file":
				err := agentFile.DoFile(cmdMap)
				if err != nil {
					exitAgent("Failed to set your file. Please try again ok_agent")
				}
			case "command":
				err := agentCommand.DoCommand(cmdMap)
				if err != nil {
					exitAgent("Failed to execute command, please check your last failed command and try again ok_agent")
				}
			case "confFile":
				err := agentAugeas.DoAugeas(cmdMap)
				if err != nil {
					exitAgent("Failed to set your configuration file and try again ok_agent")
				}
			default:

			}
		}
	}
}


func exitAgent(errMsg string) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println(r)
        }
    }()
    fmt.Println("** Exception error occur. "+errMsg+" **")
	os.Exit(1)
	// panic(fmt.Sprintf("** Exception error occur. "+errMsg+" **"))
}

func logerClientInit() {
	logerCwd := config.BaseConfig.LOGER_CWD
	fileName := config.BaseConfig.LOGER_FILE
	filePath := logerCwd + fileName
	//if fine log no exit, touch a new log file
	jsonMap := map[string]interface{}{"command": "touch " + fileName, "cwd": logerCwd, "unless": "ls " + filePath}
	agentCommand.DoCommand(jsonMap)

	// logger.SetConsole(false)
	logger.SetLevel(logger.INFO)
	logger.SetRollingDaily(logerCwd, fileName)
}

func main() {

	//第一个参数是“参数名”，第二个是“默认值”，第三个是“说明”。返回的是指针
    versionBool := flag.Bool("v", false, "version")
    detailBool := flag.Bool("d", false, "list detail for user")
 
    //正式开始Parse命令行参数
    flag.Parse()
    if *versionBool {
    	fmt.Println("ok_agent version:",config.BaseConfig.AGENT_VERSION)
    } else {

		tempJsonPath := config.BaseConfig.TEMP_JSON_FILE_PATH
		tempJsonEx := config.BaseConfig.TEMP_JSON_EX

		fmt.Println(`Starting ... Please don't interrupt your program when you meet “ completed !!” in the end !`)
		fmt.Println("----------------------")

		logerClientInit()
		//backup repo
		fileMap := map[string]interface{}{"filePath": config.BaseConfig.AGENT_REPO_BACKUP_DIR , "fileType": "dir", "unless": "ls "+config.BaseConfig.AGENT_REPO_BACKUP_DIR}
		agentFile.DoFile(fileMap)
		cmdMap := map[string]interface{}{"command": "mv -f "+config.BaseConfig.SOURCE_REPO_DIR+"/* "+config.BaseConfig.AGENT_REPO_BACKUP_DIR, 
		"onlyIf": "ls "+config.BaseConfig.SOURCE_REPO_DIR+"/*"}
		agentCommand.DoCommand(cmdMap)

		//apiList
		base_api_name := config.BaseConfig.BASE_API_NAME
		jsonStr, err := agentHttp.DoHttpRequest("POST", base_api_name)
		if err != nil {
		    fmt.Println("Http request error: ", err)
		    os.Exit(1)
		}
		if *detailBool {
			jsonMap := map[string]interface{}{"filePath": tempJsonPath + tempJsonEx + base_api_name + ".json", "fileContent": jsonStr, "fileType": "file"}
			agentFile.DoFile(jsonMap)
		}
		cmdHttp(jsonStr, base_api_name , *detailBool)


		fmt.Println("----------------------")
		fmt.Println("Congratulations, your operation is completed !!")
	}


}
