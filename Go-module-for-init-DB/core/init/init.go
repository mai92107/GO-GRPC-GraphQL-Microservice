package initial

import (
	"batchLog/core/logafa"
	"fmt"
	"os"
	"path/filepath"
)

func InitAll(){
	loadEnvFromJSON()
	logafaInit()
	machineInit()
	httpInit()
}

func loadEnvFromJSON(){
	err := loadConfigJson()
	if err != nil{
		logafa.Error(" 讀取設定 json 發生異常, error: %v",err)
	}

	err = loadMachineJson()
	if err != nil{
		logafa.Error(" 讀取機器 json 發生異常, error: %v",err)
	}
}


func loadJsonFile(fileName string) (string, error) {
	wd, _ := os.Getwd()
	configFile := "config"
	filePath := filepath.Join(wd, configFile, fileName)
	// 讀取檔案內容為 []byte
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("❌ 無法開啟 JSON 檔案: %s, error: %v", filePath, err)
	}
	return string(content), nil
}