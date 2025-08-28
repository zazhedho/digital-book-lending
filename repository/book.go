package repository

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *repoBook) Update(tx *gorm.DB, m models.Book, data interface{}) (int64, error) {
	res := tx.Table(m.TableName()).Where("id = ? AND deleted_at IS NULL", m.ID).Updates(data)
	if res.Error != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.Update; "+res.Error.Error())
		return 0, res.Error
	}

	return res.RowsAffected, nil
}

func (r *repoBook) Delete(m models.Book) (int64, error) {
	result := r.DB.Where("id = ?", m.ID).Delete(&m)
	if result.Error != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.Delete; "+result.Error.Error())
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (r *repoBook) SoftDelete(m models.Book, data interface{}) (int64, error) {
	res := r.DB.Table(m.TableName()).Where("id = ?", m.ID).Updates(data)
	if res.Error != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.SoftDelete; "+res.Error.Error())
		return 0, res.Error
	}

	return res.RowsAffected, nil
}

func (r *repoBook) GetByIsbn(isbn string) (ret models.Book, err error) {
	if err = r.DB.Where("isbn = ?", isbn).First(&ret).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBook.GetByIsbn; "+err.Error())
		return models.Book{}, err
	}

	return ret, nil
}

func (r *repoBook) Fetch(page, limit int, orderBy, orderDir, search string) (ret []models.Book, totalData int64, err error) {
	query := r.DB.Table(models.Book{}.TableName()).Where("deleted_at IS NULL")

	if strings.TrimSpace(search) != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("LOWER(title) LIKE LOWER(?) OR LOWER(author) LIKE LOWER(?) OR LOWER(isbn) LIKE LOWER(?)", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	if orderBy != "" && orderDir != "" {
		validColumns := map[string]bool{
			"title":      true,
			"author":     true,
			"isbn":       true,
			"category":   true,
			"quantity":   true,
			"created_at": true,
			"updated_at": true,
		}

		validDirections := map[string]bool{
			"asc":  true,
			"desc": true,
		}

		if _, ok := validColumns[orderBy]; !ok {
			return nil, 0, fmt.Errorf("invalid orderBy column: %s", orderBy)
		}
		if _, ok := validDirections[orderDir]; !ok {
			return nil, 0, fmt.Errorf("invalid orderDir: %s", orderDir)
		}

		query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
	}

	if limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	if err = query.Find(&ret).Error; err != nil {
		utils.WriteLog(utils.LogLevelError, "sqlBooks.Fetch; "+err.Error())
		return ret, 0, err
	}

	return ret, totalData, nil
}

func (r *repoBook) GetByIdForUpdate(tx *gorm.DB, id string) (ret models.Book, err error) {
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ret, "id = ?", id).Error
	return ret, err
}
