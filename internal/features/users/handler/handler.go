package handler

import (
	"github.com/gofiber/fiber/v2"
	"template-golang/internal/features/users/dto"
	"template-golang/internal/features/users/service"
	"template-golang/pkg/middleware"
	"template-golang/pkg/response"
	"template-golang/pkg/validator"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(r fiber.Router) {
	router := r.Group("/users")
	router.Post("/login", h.Login)
	router.Get("/me", middleware.AuthMiddleware(&[]string{}),h.GetMe)
	router.Post("/", middleware.AuthMiddleware(&[]string{"superadmin"}),h.Store)
	router.Get("/", h.ListUsers)
	router.Get("/:id", h.GetUser)
	router.Put("/:id", middleware.AuthMiddleware(&[]string{"superadmin"}),h.UpdateUser)
	router.Delete("/:id", middleware.AuthMiddleware(&[]string{"superadmin"}),h.DeleteUser)
}

// @Summary Get current user
// @Description Retrieve the authenticated user's own profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /api/v1/users/me [get]
func (h *Handler) GetMe(ctx *fiber.Ctx) error {
	data, err := h.svc.HandleMe(ctx.Context())
	if err != nil {
		return response.Error(ctx, "Failed to fetch user", err)
	}
	return response.Success(ctx, data)
}

// @Summary User login
// @Description Login for admin and superadmin users
// @Tags Users
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Router /api/v1/users/login [post]
func (h *Handler) Login(ctx *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(ctx, "Failed to parse request body", err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	data, err := h.svc.HandleLogin(ctx.Context(), req)
	if err != nil {
		return response.Error(ctx, "Invalid credentials", err)
	}

	return response.Success(ctx, data)
}

// @Summary Store new user
// @Description Store a new admin
// @Tags Users
// @Accept json
// @Produce json
// @Param body body dto.CreateUserRequest true "User registration data"
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /api/v1/users/ [post]
func (h *Handler) Store(ctx *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(ctx, "Failed to parse request body", err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	data, err := h.svc.HandleRegister(ctx.Context(), req)
	if err != nil {
		return response.Error(ctx, "Failed to register user", err)
	}

	return response.Success(ctx, data)
}

// @Summary List users
// @Description Get paginated list of users
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param perPage query int false "Items per page"
// @Success 200 {object} dto.SwaggerPaginationResponse
// @Router /api/v1/users [get]
func (h *Handler) ListUsers(ctx *fiber.Ctx) error {

	data, err := h.svc.HandleList(ctx.Context())
	if err != nil {
		return response.Error(ctx, "Failed to fetch users", err)
	}

	return response.Success(ctx, data)
}

// @Summary Get user details
// @Description Get details of a specific user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	data, err := h.svc.HandleShow(ctx.Context(), id)
	if err != nil {
		return response.Error(ctx, "Failed to fetch user", err)
	}

	return response.Success(ctx, data)
}

// @Summary Update user
// @Description Update an existing user's details
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body dto.UpdateUserRequest true "User update data"
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var req dto.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(ctx, "Failed to parse request body", err)
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	data, err := h.svc.HandleUpdate(ctx.Context(), id, req)
	if err != nil {
		return response.Error(ctx, "Failed to update user", err)
	}

	return response.Success(ctx, data)
}

// @Summary Delete user
// @Description Delete an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	user, err := h.svc.HandleDelete(ctx.Context(), id)
	if err != nil {
		return response.Error(ctx, "Failed to delete user", err)
	}

	return response.Success(ctx, user)
}
