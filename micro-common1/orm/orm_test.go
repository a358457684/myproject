package orm

import (
	"common/log"
	"gorm.io/gorm"
	"testing"
	"time"
)

type TestUser struct {
	gorm.Model
	username string
}

type Book struct {
	ID     string `gorm:"primarykey"`
	Name   string
	Size   int
	Create int
	Update int
}

func (Book) TableName() string {
	return "test"
}

func (e *Book) BeforeCreate(tx *gorm.DB) (err error) {
	e.Create = time.Now().Second()
	e.Update = time.Now().Second()
	return
}

func (e *Book) BeforeUpdate(tx *gorm.DB) (err error) {
	e.Update = time.Now().Second()
	return
}

func TestBookDbWrite(t *testing.T) {
	book := &Book{
		ID:   "a",
		Name: "沟通qwer",
		Size: 54,
	}
	DB.AutoMigrate(book)
	DB.Create(book)
	var data []Book
	DB.Find(&data)
	log.Infoln(data)
}

func TestBookDbRead(t *testing.T) {
	book := &Book{}
	DB.First(book, "id", "a")
	log.Info(book)
}

func TestDb(t *testing.T) {
	user := &TestUser{
		username: "username",
	}
	DB.AutoMigrate(&TestUser{})
	create := DB.Create(user)
	log.Info(create.Name())

	var firstUser TestUser
	DB.First(&firstUser)
	log.Info(firstUser)
}
