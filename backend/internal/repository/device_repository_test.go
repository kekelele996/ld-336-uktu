package repository

import (
	"errors"
	"testing"

	"github.com/medasset/medasset/internal/model"
)

func TestDeviceRepositoryCRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewDeviceRepository(db)

	d := &model.Device{AssetCode: "MA-0001", Name: "CT机", Status: "in_storage", Department: "放射科", Category: "影像设备"}
	if err := repo.Create(d); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	found, err := repo.FindByAssetCode("MA-0001")
	if err != nil {
		t.Fatalf("FindByAssetCode failed: %v", err)
	}
	if found.Name != "CT机" {
		t.Errorf("Name = %q", found.Name)
	}
	list, total, err := repo.List(1, 10, "放射科", "", "", "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("total=%d len=%d", total, len(list))
	}
	if err := repo.UpdateStatusTx(db, d.ID, "in_use"); err != nil {
		t.Fatalf("UpdateStatusTx failed: %v", err)
	}
	after, _ := repo.FindByID(d.ID)
	if after.Status != "in_use" {
		t.Errorf("status = %q, want in_use", after.Status)
	}
	if err := repo.Delete(d.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := repo.FindByID(d.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeviceRepositoryCountAndGroup(t *testing.T) {
	db := newTestDB(t)
	repo := NewDeviceRepository(db)
	_ = repo.Create(&model.Device{AssetCode: "A1", Name: "x", Status: "in_use", Department: "心内科", Manufacturer: "GE"})
	_ = repo.Create(&model.Device{AssetCode: "A2", Name: "y", Status: "in_use", Department: "心内科", Manufacturer: "GE"})
	_ = repo.Create(&model.Device{AssetCode: "A3", Name: "z", Status: "in_storage", Department: "放射科", Manufacturer: "SIEMENS"})
	n, err := repo.Count("in_use")
	if err != nil || n != 2 {
		t.Errorf("Count(in_use) = %d, err=%v", n, err)
	}
	dist, err := repo.GroupCount("department")
	if err != nil {
		t.Fatalf("GroupCount failed: %v", err)
	}
	if dist["心内科"] != 2 || dist["放射科"] != 1 {
		t.Errorf("department dist = %v", dist)
	}
}
