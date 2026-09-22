package service

import (
	"net/http"
	"testing"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

func TestUserServiceRegisterAndLogin(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := NewUserService(repository.NewUserRepository(env.db), env.audit, "test-secret", 24, env.logger)

	req := &dto.RegisterReq{Username: "doctor1", Password: "secret123", RealName: "张医生", Role: constants.RoleDepartment, Department: "心内科"}
	user, err := svc.Register(req, "127.0.0.1")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.Role != constants.RoleDepartment {
		t.Errorf("role = %q", user.Role)
	}

	// 重复用户名。
	if _, err := svc.Register(req, "127.0.0.1"); err == nil {
		t.Error("expected duplicate username error")
	} else if appErr, ok := err.(*util.AppError); ok && appErr.Code != http.StatusConflict {
		t.Errorf("expected conflict, got %d", appErr.Code)
	}

	// 登录成功。
	resp, err := svc.Login(&dto.LoginReq{Username: "doctor1", Password: "secret123"}, "127.0.0.1")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token")
	}

	// 密码错误。
	if _, err := svc.Login(&dto.LoginReq{Username: "doctor1", Password: "wrong"}, "127.0.0.1"); err == nil {
		t.Error("expected login error for wrong password")
	}
}

func TestUserServiceListAndUpdate(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := NewUserService(repository.NewUserRepository(env.db), env.audit, "test-secret", 24, env.logger)
	_ = svc.SeedAdmin()

	user, err := svc.Me(1)
	if err != nil {
		t.Fatalf("Me failed: %v", err)
	}
	if user.Username != "admin" || user.Role != constants.RoleSuperAdmin {
		t.Errorf("seed admin mismatch: %+v", user)
	}

	result, err := svc.List(1, 10, "", "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}

	updated, err := svc.Update(user.ID, &dto.UpdateUserReq{RealName: "超级管理员", Role: constants.RoleSuperAdmin, Status: constants.UserStatusActive}, "admin")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.RealName != "超级管理员" {
		t.Errorf("RealName = %q", updated.RealName)
	}
}
