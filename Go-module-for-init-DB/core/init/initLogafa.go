package initial

import (
	"batchLog/core/global"
	"batchLog/core/logafa"
	"fmt"
	"os"
	"path/filepath"
	"time"
)


func logafaInit(){
	env := global.ConfigSetting.Env

	switch env {
	case "dev":
		logafa.CurrentLevel = logafa.DEBUG
	case "prod":
		logafa.CurrentLevel = logafa.INFO
	case "test":
		logafa.CurrentLevel = logafa.WARN
	default:
		logafa.CurrentLevel = logafa.DEBUG
	}

	now := time.Now()

	var err error
	wd, _ := os.Getwd()
	filePath := filepath.Join(wd, "log", now.Format("2006-01-02") + ".log")
	logafa.LogFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("無法打開 log 檔案: %v", err))
	}
}