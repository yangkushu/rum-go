package database

import (
	"fmt"
	"gorm.io/gorm/clause"
	"testing"
	"time"
)

type TestOnConflictTable struct {
	Id        int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar;size:255;unique;not null"`
	Age       int       `gorm:"column:age;type:int;size:5;not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime"`
}

func (TestOnConflictTable) TableName() string {
	return "test_onconflict_table"
}

func TestPostgresOnConflict(t *testing.T) {
	pg, err := NewPostgres(
		&PostgresConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "123456",
			DBName:   "postgres",
		})

	if err != nil {
		panic(fmt.Errorf("failed to connect to postgres: %w", err))
	}

	// Create table
	err = pg.AutoMigrate(&TestOnConflictTable{})
	if err != nil {
		panic(fmt.Errorf("failed to create table: %w", err))
	}

	// Insert data

	pg.Create(&TestOnConflictTable{
		Name: "Alice",
		Age:  20,
	})

	pg.Create(&TestOnConflictTable{
		Name: "Bob",
		Age:  21,
	})

	time.Sleep(time.Second * 10)

	// Insert data with conflict
	pg.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"age"}),
	}).Create(

		[]TestOnConflictTable{
			{
				Name: "Alice",
				Age:  30,
			},
			{
				Name: "Charlie",
				Age:  22,
			},
		},
	)

	pg.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"age", "updated_at"}),
	}).Create(

		[]TestOnConflictTable{
			{
				Name: "Bob",
				Age:  31,
			},
			{
				Name: "David",
				Age:  23,
			},
		},
	)

}
