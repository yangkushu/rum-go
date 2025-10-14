package database

import (
	"gorm.io/gorm"
	"reflect"
	"strings"
)

// autoSelectFields 回调，自动设置 SELECT 字段
func autoSelectFields(db *gorm.DB) {
	if len(db.Statement.Selects) == 0 { // 如果没有自定义SELECT
		if db.Statement.Model != nil {
			columns := generateSelectColumns(db.Statement.Model)
			db.Statement.Selects = columns
		}
	}
}

// generateSelectColumns 生成模型的数据库字段列表
func generateSelectColumns(model interface{}) []string {
	// 获取reflect.Type
	modelType := reflect.TypeOf(model)

	// 如果传入的是指针，解引用获取实际类型
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 如果类型是切片，获取元素的类型
	if modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
		// 如果切片的元素是指针，再次解引用
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
	}

	// 确保此时的modelType是struct
	if modelType.Kind() != reflect.Struct {
		panic("generateSelectColumns: expected struct or slice of structs, got " + modelType.Kind().String())
	}

	var dbFields []string

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			tags := strings.Split(gormTag, ";")
			for _, tag := range tags {
				if strings.HasPrefix(tag, "column:") {
					columnName := strings.TrimPrefix(tag, "column:")
					dbFields = append(dbFields, columnName)
				}
			}
		}
	}
	return dbFields
}
