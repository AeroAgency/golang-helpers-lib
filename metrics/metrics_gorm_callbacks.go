package metrics

import (
	"gorm.io/gorm"
	"time"
)

func RegisterGormMetricCallbacks(db *gorm.DB, m Metrics) {
	db.Callback().Query().Before("*").Register("count_start", func(db *gorm.DB) {
		start := time.Now()
		db.InstanceSet("start", start)
	})

	db.Callback().Query().After("*").Register("error_handle", func(db *gorm.DB) {
		if db.Statement.Error != nil {
			m.Inc(StorageRequestError.Name, "postgresql")
			return
		}
		m.Inc(StorageRequestSuccessful.Name, "postgresql")
	})

	db.Callback().Query().After("*").Register("count_end", func(db *gorm.DB) {
		start, _ := db.InstanceGet("start")
		elapsed := time.Since(start.(time.Time)).Seconds()
		m.Observe(StorageRequestExecutionTime.Name, elapsed, "postgresql")
	})
}
