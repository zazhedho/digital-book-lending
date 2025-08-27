package repository

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"

	"gorm.io/gorm"
)

func NewBookRepo(db *gorm.DB) interfaces.Book {
	return &repoBook{
		DB: db,
	}
}

type repoBook struct {
	DB *gorm.DB
}

func (r *repoBook) Store(m models.Book) error {
	if err := r.DB.Create(&m).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.Store; "+err.Error())
		return err
	}

	return nil
}

func (r *repoBook) Update(m models.Book, data interface{}) (int64, error) {
	res := r.DB.Table(m.TableName()).Where("id = ?", m.ID).Updates(data)
	if res.Error != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.Update; "+res.Error.Error())
		return 0, res.Error
	}

	return res.RowsAffected, nil
}

func (r *repoBook) Delete(m models.Book) (int64, error) {
	result := r.DB.Delete(&m)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *repoBook) GetByIsbn(isbn string) (ret models.Book, err error) {
	if err = r.DB.Where("isbn = ?", isbn).First(&ret).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.GetByIsbn; "+err.Error())
		return models.Book{}, err
	}

	return ret, nil
}
