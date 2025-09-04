package services

import (
	"digital-book-lending/interfaces"
	"digital-book-lending/models"
	"digital-book-lending/utils"
	"digital-book-lending/utils/request"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo      interfaces.Users
	blacklistRepo interfaces.Blacklist
}

func NewUserService(userRepo interfaces.Users, blacklistRepo interfaces.Blacklist) *UserService {
	return &UserService{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
	}
}

func (s *UserService) RegisterUser(req request.Register) (models.Users, error) {
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

	if err = s.userRepo.Store(user); err != nil {
		return models.Users{}, err
	}

	return user, nil
}

func (s *UserService) LoginUser(req request.Login, logId string) (string, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
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
	blacklist := models.Blacklist{
		ID:        utils.CreateUUID(),
		Token:     token,
		CreatedAt: time.Now(),
	}

	if err := s.blacklistRepo.Store(blacklist); err != nil {
		return err
	}

	return nil
}
