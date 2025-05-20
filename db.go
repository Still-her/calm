package calm

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type GormModel struct {
	Id int64 `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
}

func OpenSDB(dsn string, config *gorm.Config, maxIdleConns, maxOpenConns int, models ...interface{}) (sdb *gorm.DB, err error) {
	if config == nil {
		config = &gorm.Config{}
	}
	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix:   "a_",
			SingularTable: false,
		}
	}
	if sdb, err = gorm.Open(mysql.Open(dsn), config); err != nil {
		return
	}
	if sqlSDB, err := sdb.DB(); err == nil {
		sqlSDB.SetMaxIdleConns(maxIdleConns)
		sqlSDB.SetMaxOpenConns(maxOpenConns)
	} else {

	}

	if err = sdb.AutoMigrate(models...); nil != err {
	}
	return
}

func OpenMDB(dsn string, config *gorm.Config, maxIdleConns, maxOpenConns int, models ...interface{}) (mdb *gorm.DB, err error) {
	if config == nil {
		config = &gorm.Config{}
	}

	config.NamingStrategy = schema.NamingStrategy{
		TablePrefix:   "",
		SingularTable: true,
	}

	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		}
	}

	if mdb, err = gorm.Open(postgres.Open(dsn), config); err != nil {
		return
	}

	if sqlMDB, err := mdb.DB(); err == nil {
		sqlMDB.SetMaxIdleConns(maxIdleConns)
		sqlMDB.SetMaxOpenConns(maxOpenConns)
	} else {

	}

	if err = mdb.AutoMigrate(models...); nil != err {
	}
	return
}
