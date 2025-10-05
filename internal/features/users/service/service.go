package service

import (
	"context"

	"template-golang/internal/db/model"
	"template-golang/internal/features/base"
	"template-golang/internal/features/users/dto"
	"template-golang/pkg/apperror"
	"template-golang/pkg/helper"
	"template-golang/pkg/pagination"
	"gorm.io/gorm"
)

type Service struct {
	*base.BaseService
}

func NewService(baseService *base.BaseService) *Service {
	return &Service{
		BaseService: baseService,
	}
}

func (s *Service) HandleMe(ctx context.Context) (any, error) {
	var user model.User
	userID := ctx.Value("user_id").(string)
	if err := s.DB().First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) HandleList(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := s.DB().Find(&users).Error
	if err != nil {
		return []model.User{}, err
	}
	return users, nil
}

func (s *Service) HandleRegister(ctx context.Context, req dto.CreateUserRequest) (model.User, error) {
	userAny, err := s.InTx(ctx, func(tx *gorm.DB) (any, error) {

		user := model.User{
			Name:  req.Name,
			Email: req.Email,
			Role:  model.RoleAdmin,
		}



		if err := tx.Create(&user).Error; err != nil {
			return model.User{}, err
		}

		return user, nil
	})
	if err != nil {
		return model.User{}, apperror.New("users", "failed to register user", 400, err, req.Name)
	}
	return userAny.(model.User), nil
}

func (s *Service) HandleLogin(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error) {
	var user model.User
	err := s.DB().First(&user, "email = ?", req.Email).Error
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if err = helper.CompareHashAndPassword(user.Password, req.Password); err != nil {
		return dto.LoginResponse{}, err
	}
	tokenString, err := helper.GenerateJwtToken(user)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	return dto.LoginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		Token: tokenString,
	}, nil
}

func (s *Service) HandleIndex(ctx context.Context, page, perPage int32) (pagination.PaginationResponse[model.User], error) {
	offset := (page - 1) * perPage
	var users []model.User
	err := s.DB().Limit(int(perPage)).Offset(int(offset)).Find(&users).Error
	if err != nil {
		return pagination.PaginationResponse[model.User]{}, err
	}
	var total int64
	err = s.DB().Model(&model.User{}).Count(&total).Error
	if err != nil {
		return pagination.PaginationResponse[model.User]{}, err
	}
	var response []model.User
	for _, user := range users {
		response = append(response, user)
	}
	return pagination.Paginate(
		response,
		int(total),
		int(page),
		int(perPage),
		"/api/users",
	), nil
}

func (s *Service) HandleShow(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := s.DB().First(&user, "id = ?", id).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *Service) HandleCreate(ctx context.Context, req dto.CreateUserRequest) (model.User, error) {
	userAny, err := s.InTx(ctx, func(tx *gorm.DB) (any, error) {
		user := model.User{
			Name:  req.Name,
			Email: req.Email,
		}
		if err := tx.Create(&user).Error; err != nil {
			return model.User{}, err
		}
		return user, nil
	})
	if err != nil {
		return model.User{}, apperror.New("users", "failed to create user", 400, err, req.Name)
	}
	return userAny.(model.User), nil
}



func (s *Service) HandleGetUserByToken(ctx context.Context) (model.User, error) {
	userID, err := helper.GetUserIDFromToken(ctx.Value("token").(string))
	if err != nil {
		return model.User{}, err
	}
	var user model.User
	err = s.DB().First(&user, "id = ?", userID).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *Service) HandleUpdate(ctx context.Context, id string, req dto.UpdateUserRequest) (model.User, error) {
	userAny, err := s.InTx(ctx, func(tx *gorm.DB) (any, error) {
		var existingUser model.User
		if err := tx.First(&existingUser, "id = ?", id).Error; err != nil {
			return model.User{}, err
		}
		if req.Name != nil {
			existingUser.Name = *req.Name
		}
		if req.Email != nil {
			existingUser.Email = *req.Email
		}
		if err := tx.Save(&existingUser).Error; err != nil {
			return model.User{}, err
		}
		return existingUser, nil
	})
	if err != nil {
		return model.User{}, apperror.New("users", "failed to update user", 400, err, id)
	}
	return userAny.(model.User), nil
}

func (s *Service) HandleDelete(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := s.DB().First(&user, "id = ?", id).Error
	if err != nil {
		return model.User{}, err
	}
	err = s.DB().Delete(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
