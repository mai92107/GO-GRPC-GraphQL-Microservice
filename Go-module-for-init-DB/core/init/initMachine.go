package initial

import (
	jsonModal "batchLog/config"
	"batchLog/core/global"
	"batchLog/core/logafa"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func loadMachineJson() error{
	fileName := "machine.json"
	// 打開 JSON 檔案
	data, err := loadJsonFile(fileName)
	if err != nil {
		return nil
	}

	var machine jsonModal.Machine
	// 解析 JSON
	err = jsoniter.UnmarshalFromString(data, &machine)
	if err != nil {
		return fmt.Errorf("❌ 解析 JSON 失敗: %s, error: %v",fileName, err)
	}
	// 需要maria有需要使用再載入全域變數
	if machine.MariaDB.InUse{
		global.MariaDBSetting = machine.MariaDB
	}
	return nil
}

func machineInit(){
	mariadb := global.MariaDBSetting

	if ( mariadb.InUse ){
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",mariadb.User,mariadb.Password,mariadb.Host,mariadb.Port,mariadb.Name)

		// logafa.Debug("資料庫連線 DNS: %s",dsn)
		// 連接資料庫
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			logafa.Error("❌ 無法連接資料庫:", err)
		}
	
		logafa.Debug("✅ 資料庫連接成功")
		global.DB = db
	}
}