package dto

import (
	"time"

	"github.com/medasset/medasset/internal/model"
)

// AuditListItem 审计日志列表项。
type AuditListItem struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Module    string    `json:"module"`
	EntityID  string    `json:"entity_id"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
	RequestID string    `json:"request_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ToAuditListItem 转换审计日志实体为列表项。
func ToAuditListItem(a *model.AuditLog) AuditListItem {
	return AuditListItem{
		ID:        a.ID,
		UserID:    a.UserID,
		Username:  a.Username,
		Action:    a.Action,
		Module:    a.Module,
		EntityID:  a.EntityID,
		Detail:    a.Detail,
		IP:        a.IP,
		RequestID: a.RequestID,
		CreatedAt: a.CreatedAt,
	}
}
