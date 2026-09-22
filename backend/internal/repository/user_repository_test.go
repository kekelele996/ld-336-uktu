package repository

import (
	"errors"
	"testing"

	"github.com/medasset/medasset/internal/model"
)

func TestUserRepositoryCreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	u := &model.User{Username: "alice", Password: "hash", RealName: "爱丽丝", Role: "DEVICE_ADMIN", Status: "active"}
	if err := repo.Create(u); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected id assigned")
	}

	found, err := repo.FindByUsername("alice")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if found.RealName != "爱丽丝" {
		t.Errorf("RealName = %q", found.RealName)
	}

	if _, err := repo.FindByUsername("nobody"); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepositoryListAndDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	for _, name := range []string{"u1", "u2", "u3"} {
		_ = repo.Create(&model.User{Username: name, Password: "x", RealName: name, Role: "DEPARTMENT", Status: "active"})
	}
	list, total, err := repo.List(1, 2, "", "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 || len(list) != 2 {
		t.Errorf("total=%d len=%d, want 3/2", total, len(list))
	}
	if err := repo.Delete(list[0].ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := repo.FindByID(list[0].ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
