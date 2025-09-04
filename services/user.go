package services

import (
	"digital-book-lending/models"
	"digital-book-lending/repository"
	"digital-book-lending/utils"
	"digital-book-lending/utils/request"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

func (s *UserService) RegisterUser(req request.Register) (models.Users, error) {
	userRepo := repository.NewUserRepo(s.DB)

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Users{}, err
	}

	user := models.Users{
		Id:        utils.CreateUUID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPwd),
		Role:      utils.RoleMember,
		CreatedAt: time.Now(),
	}

	if err = userRepo.Store(user); err != nil {
		return models.Users{}, err
	}

	return user, nil
}

func (s *UserService) LoginUser(req request.Login, logId string) (string, error) {
	userRepo := repository.NewUserRepo(s.DB)

	user, err := userRepo.GetByEmail(req.Email)
	if err != nil {
		return "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", err
	}

	token, err := utils.GenerateJwt(&user, logId)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) LogoutUser(token string) error {
	blacklistRepo := repository.NewBlacklistRepo(s.DB)

	blacklist := models.Blacklist{
		ID:        utils.CreateUUID(),
		Token:     token,
		CreatedAt: time.Now(),
	}

	if err := blacklistRepo.Store(blacklist); err != nil {
		return err
	}

	return nil
}
