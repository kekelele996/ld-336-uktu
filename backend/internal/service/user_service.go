package service

import (
	"fmt"
	"errors"
	"log/slog"
	"net/http"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

// UserService 用户与认证服务。
type UserService struct {
	repo    *repository.UserRepository
	audit   *AuditService
	cfg     *serviceConfig
	log     *slog.Logger
}

type serviceConfig struct {
	JWTSecret string
	JWTExpireHours int
}

func NewUserService(repo *repository.UserRepository, audit *AuditService, jwtSecret string, expireHours int, log *slog.Logger) *UserService {
	return &UserService{repo: repo, audit: audit, cfg: &serviceConfig{JWTSecret: jwtSecret, JWTExpireHours: expireHours}, log: log}
}

// Register 注册用户（默认科室角色）。
func (s *UserService) Register(req *dto.RegisterReq, ip string) (*model.User, error) {
	role := req.Role
	if role == "" {
		role = constants.RoleDepartment
	}
	if !constants.ValidRoles[role] {
		return nil, util.NewAppError(http.StatusBadRequest, "角色不合法: role="+role, nil)
	}
	taken, err := s.repo.IsUsernameTaken(req.Username)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if taken {
		return nil, util.NewAppError(http.StatusConflict, constants.MsgDuplicateUser, nil)
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "密码加密失败", err)
	}
	user := &model.User{
		Username:   req.Username,
		Password:   hash,
		RealName:   req.RealName,
		Role:       role,
		Department: req.Department,
		Phone:      req.Phone,
		Email:      req.Email,
		Status:     constants.UserStatusActive,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "创建用户失败: username="+req.Username, err)
	}
	s.log.Info(fmt.Sprintf(constants.LogUserRegistered, user.Username, user.RealName, user.Role))
	s.audit.Record(user.ID, user.Username, "REGISTER", "user", util.Uint64String(user.ID), "用户注册", ip, "")
	return user, nil
}

// Login 登录并签发 JWT。
func (s *UserService) Login(req *dto.LoginReq, ip string) (*dto.LoginResp, error) {
	user, err := s.repo.FindByUsername(req.Username)
	if errors.Is(err, repository.ErrNotFound) {
		s.log.Info(fmt.Sprintf(constants.LogUserLoginFailed, req.Username, "用户不存在"))
		return nil, util.NewAppError(http.StatusUnauthorized, constants.MsgWrongPassword, nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if !util.CheckPassword(user.Password, req.Password) {
		s.log.Info(fmt.Sprintf(constants.LogUserLoginFailed, req.Username, "密码错误"))
		return nil, util.NewAppError(http.StatusUnauthorized, constants.MsgWrongPassword, nil)
	}
	if user.Status != constants.UserStatusActive {
		return nil, util.NewAppError(http.StatusForbidden, constants.MsgUserDisabled, nil)
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, util.HoursDuration(s.cfg.JWTExpireHours), user.ID, user.Username, user.Role)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "生成令牌失败", err)
	}
	s.log.Info(fmt.Sprintf(constants.LogUserLogin, user.Username, user.Role))
	s.audit.Record(user.ID, user.Username, "LOGIN", "auth", util.Uint64String(user.ID), "用户登录", ip, "")
	return &dto.LoginResp{Token: token, User: user}, nil
}

// Me 查询当前用户。
func (s *UserService) Me(userID uint) (*model.User, error) {
	user, err := s.repo.FindByID(userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, constants.MsgNotFound, nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	return user, nil
}

// List 分页查询用户。
func (s *UserService) List(page, pageSize int, role, keyword string) (*util.PageResult, error) {
	list, total, err := s.repo.List(page, pageSize, role, keyword)
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	items := make([]dto.UserListItem, 0, len(list))
	for i := range list {
		items = append(items, dto.ToUserListItem(&list[i]))
	}
	return &util.PageResult{List: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// Update 更新用户信息。
func (s *UserService) Update(id uint, req *dto.UpdateUserReq, operator string) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, util.NewAppError(http.StatusNotFound, "用户不存在: user_id="+util.Uint64String(id), nil)
	}
	if err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	user.RealName = req.RealName
	user.Role = req.Role
	user.Department = req.Department
	user.Phone = req.Phone
	user.Email = req.Email
	if req.Status != "" {
		user.Status = req.Status
	}
	if err := s.repo.Update(user); err != nil {
		return nil, util.NewAppError(http.StatusInternalServerError, "更新用户失败: user_id="+util.Uint64String(id), err)
	}
	s.log.Info(fmt.Sprintf(constants.LogUserUpdated, user.ID, user.Username, user.Role))
	s.audit.Record(user.ID, user.Username, "UPDATE", "user", util.Uint64String(user.ID), "更新用户信息", operator, "")
	return user, nil
}

// Delete 删除用户。
func (s *UserService) Delete(id uint, operator string) error {
	user, err := s.repo.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return util.NewAppError(http.StatusNotFound, constants.MsgNotFound, nil)
	}
	if err != nil {
		return util.NewAppError(http.StatusInternalServerError, constants.MsgInternalError, err)
	}
	if err := s.repo.Delete(id); err != nil {
		return util.NewAppError(http.StatusInternalServerError, "删除用户失败: user_id="+util.Uint64String(id), err)
	}
	s.log.Info(fmt.Sprintf(constants.LogUserDeleted, user.ID, user.Username, operator))
	s.audit.Record(0, operator, "DELETE", "user", util.Uint64String(user.ID), "删除用户: "+user.Username, operator, "")
	return nil
}

// SeedAdmin 初始化默认管理员（无则创建）。
func (s *UserService) SeedAdmin() error {
	_, err := s.repo.FindByUsername("admin")
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := util.HashPassword("admin123")
	if err != nil {
		return err
	}
	admin := &model.User{
		Username: "admin",
		Password: hash,
		RealName: "系统管理员",
		Role:     constants.RoleSuperAdmin,
		Status:   constants.UserStatusActive,
	}
	if err := s.repo.Create(admin); err != nil {
		return err
	}
	s.log.Info(fmt.Sprintf(constants.LogUserRegistered, admin.Username, admin.RealName, admin.Role))
	return nil
}
