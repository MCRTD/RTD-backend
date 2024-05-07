package setting

import "os"

type DatabaseSettingS struct {
	DBType       string
	UserName     string
	Password     string
	Host         string
	DBName       string
	Charset      string
	ParseTime    bool
	MaxIdleConns int
	MaxOpenConns int
}

func GetDatabaseSetting() *DatabaseSettingS {

	databaseSetting := DatabaseSettingS{
		DBType:       "postgres",
		UserName:     os.Getenv("UserName"),
		Password:     os.Getenv("Password"),
		Host:         os.Getenv("Host"),
		DBName:       os.Getenv("DBName"),
		Charset:      "utf8",
		ParseTime:    true,
		MaxIdleConns: 10,
		MaxOpenConns: 30,
	}

	return &databaseSetting
}
