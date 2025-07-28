package service

import (
	"errors"
	"smart-analysis/internal/model"
	"smart-analysis/internal/utils"
)

type UserService struct {
	// 这里应该有数据库连接，为简化先用内存存储
	users  map[int]*model.User
	nextID int
}

func NewUserService() *UserService {
	return &UserService{
		users:  make(map[int]*model.User),
		nextID: 1,
	}
}

// Register 用户注册
func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	// 检查用户是否已存在
	for _, user := range s.users {
		if user.Email == req.Email {
			return nil, errors.New("email already exists")
		}
		if user.Username == req.Username {
			return nil, errors.New("username already exists")
		}
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		ID:       s.nextID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	s.users[s.nextID] = user
	s.nextID++

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *model.LoginRequest) (*model.User, error) {
	// 查找用户
	var user *model.User
	for _, u := range s.users {
		if u.Email == req.Email {
			user = u
			break
		}
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int) (*model.User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID int, req *model.UpdateProfileRequest) (*model.User, error) {
	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// 检查用户名是否重复
	if req.Username != "" && req.Username != user.Username {
		for _, u := range s.users {
			if u.Username == req.Username && u.ID != userID {
				return nil, errors.New("username already exists")
			}
		}
		user.Username = req.Username
	}

	// 检查邮箱是否重复
	if req.Email != "" && req.Email != user.Email {
		for _, u := range s.users {
			if u.Email == req.Email && u.ID != userID {
				return nil, errors.New("email already exists")
			}
		}
		user.Email = req.Email
	}

	return user, nil
}
