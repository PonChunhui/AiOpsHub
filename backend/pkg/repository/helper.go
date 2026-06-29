package repository

import "gorm.io/gorm"

func Paginate(db *gorm.DB, page, pageSize int) *gorm.DB {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return db.Limit(pageSize).Offset(offset)
}

func Count(db *gorm.DB, model interface{}) (int64, error) {
	var total int64
	err := db.Model(model).Count(&total).Error
	return total, err
}
