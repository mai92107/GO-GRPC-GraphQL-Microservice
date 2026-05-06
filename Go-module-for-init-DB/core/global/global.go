package global

import (
	jsonModal "batchLog/config"

	"gorm.io/gorm"
)

var (
	DB *gorm.DB

	ConfigSetting jsonModal.Config
	MariaDBSetting jsonModal.MariaDbConfig
)


