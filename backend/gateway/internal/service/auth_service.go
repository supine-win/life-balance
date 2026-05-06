// Package service 认证服务
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/life-recorder/gateway/internal/config"
	"github.com/life-recorder/gateway/internal/model"
	"github.com/life-recorder/gateway/internal/pkg/hash"
	"github.com/life-recorder/gateway/internal/pkg/jwt"
	pkguuid "github.com/life-recorder/gateway/internal/pkg/uuid"
	"github.com/life-recorder/gateway/internal/repository"
)

// AuthService 认证服务
type AuthService struct {
	userRepo  *repository.UserRepo
	roleRepo  *repository.RoleRepo
	jwtConfig *config.JWTConfig
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo *repository.UserRepo, roleRepo *repository.RoleRepo, jwtConfig *config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		jwtConfig: jwtConfig,
	}
}

// RegisterInput 注册输入
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Nickname string `json:"nickname"`
}

// LoginInput 登录输入
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResult 认证结果
type AuthResult struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
}

// Register 注册
func (s *AuthService) Register(input *RegisterInput) (*AuthResult, error) {
	// 检查用户名和邮箱唯一性
	existing, _ := s.userRepo.FindByUsername(input.Username)
	if existing != nil {
		return nil, errors.New("用户名或邮箱已存在")
	}

	// 哈希密码
	passwordHash, err := hash.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username: input.Username,
		Email:    input.Email,
		Password: passwordHash,
		Nickname: input.Nickname,
		Status:   1,
	}
	if user.Nickname == "" {
		user.Nickname = input.Username
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配默认角色（user）
	roles, _ := s.roleRepo.List()
	for _, role := range roles {
		if role.Name == "user" {
			s.roleRepo.AssignRoles(user.ID, []string{role.ID})
			break
		}
	}

	// 重新加载用户（包含角色）
	user, _ = s.userRepo.FindByID(user.ID)

	// 生成令牌
	return s.generateTokens(user)
}

// Login 登录
func (s *AuthService) Login(input *LoginInput) (*AuthResult, error) {
	user, err := s.userRepo.FindByUsername(input.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	if !hash.CheckPassword(input.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	s.userRepo.Update(user)

	return s.generateTokens(user)
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(refreshToken string) (*AuthResult, error) {
	claims, err := jwt.ParseToken(refreshToken, s.jwtConfig.Secret)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	return s.generateTokens(user)
}

// GetCurrentUser 获取当前用户
func (s *AuthService) GetCurrentUser(userID string) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// generateTokens 生成令牌对
func (s *AuthService) generateTokens(user *model.User) (*AuthResult, error) {
	var roleNames []string
	var permNames []string
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
		for _, p := range r.Permissions {
			permNames = append(permNames, p.Name)
		}
	}

	accessToken, err := jwt.GenerateAccessToken(
		user.ID, user.Username, roleNames, permNames,
		s.jwtConfig.Secret, s.jwtConfig.AccessTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := jwt.GenerateRefreshToken(
		user.ID, user.Username,
		s.jwtConfig.Secret, s.jwtConfig.RefreshTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtConfig.AccessTTL,
	}, nil
}

// SetupInit 系统初始化
func (s *AuthService) SetupInit(input *SetupInitInput) (*AuthResult, error) {
	input.Username = "admin"
	input.Nickname = "管理员"

	result, err := s.Register(&RegisterInput{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
		Nickname: input.Nickname,
	})
	if err != nil {
		return nil, err
	}

	// 分配管理员角色
	roles, _ := s.roleRepo.List()
	var adminRoleIDs []string
	for _, role := range roles {
		if role.Name == "admin" {
			adminRoleIDs = append(adminRoleIDs, role.ID)
		}
	}
	if len(adminRoleIDs) > 0 {
		s.roleRepo.AssignRoles(result.User.ID, adminRoleIDs)
	}

	return result, nil
}

// SetupInitInput 初始化输入
type SetupInitInput struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname"`
}

// SeedDefaultData 种子数据
func SeedDefaultData(db interface {
	Exec(sql string, values ...interface{}) (interface{}, error)
}) error {
	// 创建默认角色
	roles := []model.Role{
		{ID: pkguuid.New(), Name: "admin", Description: "管理员", SortOrder: 0, Status: 1},
		{ID: pkguuid.New(), Name: "user", Description: "普通用户", SortOrder: 1, Status: 1},
		{ID: pkguuid.New(), Name: "viewer", Description: "访客", SortOrder: 2, Status: 1},
	}

	// 创建默认权限
	permissions := []model.Permission{
		{ID: pkguuid.New(), Name: "event:create", Resource: "event", Action: "create", Description: "创建事件"},
		{ID: pkguuid.New(), Name: "event:read", Resource: "event", Action: "read", Description: "查看事件"},
		{ID: pkguuid.New(), Name: "event:update", Resource: "event", Action: "update", Description: "更新事件"},
		{ID: pkguuid.New(), Name: "event:delete", Resource: "event", Action: "delete", Description: "删除事件"},
		{ID: pkguuid.New(), Name: "media:upload", Resource: "media", Action: "create", Description: "上传媒体"},
		{ID: pkguuid.New(), Name: "media:read", Resource: "media", Action: "read", Description: "查看媒体"},
		{ID: pkguuid.New(), Name: "media:delete", Resource: "media", Action: "delete", Description: "删除媒体"},
		{ID: pkguuid.New(), Name: "admin:users", Resource: "admin", Action: "manage", Description: "用户管理"},
		{ID: pkguuid.New(), Name: "admin:config", Resource: "admin", Action: "config", Description: "系统配置"},
		{ID: pkguuid.New(), Name: "admin:backup", Resource: "admin", Action: "backup", Description: "数据备份"},
	}

	_ = roles
	_ = permissions
	// TODO: 使用 GORM 批量插入

	return nil
}
