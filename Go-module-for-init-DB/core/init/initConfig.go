package initial

import (
	jsonModal "batchLog/config"
	"batchLog/core/global"

	jsoniter "github.com/json-iterator/go"
)


func loadConfigJson()error{
	fileName := "config.json"
	// 打開 JSON 檔案
	data, err := loadJsonFile(fileName)
	if err != nil {
		return nil
	}

	var config jsonModal.Config
	// 解析 JSON
	err = jsoniter.UnmarshalFromString(data, &config)
	if err != nil {
		return err
	}
	global.ConfigSetting = config
	return nil
}

