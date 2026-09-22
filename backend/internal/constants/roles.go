package constants

// 用户角色枚举：贯穿模型、DTO、RBAC 中间件、前端路由守卫与按钮显隐。
const (
	RoleSuperAdmin  = "SUPER_ADMIN"  // 系统管理员
	RoleDeviceAdmin = "DEVICE_ADMIN" // 设备科
	RoleDean        = "DEAN"         // 院长
	RoleDepartment  = "DEPARTMENT"   // 科室
	RoleEngineer    = "ENGINEER"     // 维修工程师
)

// ValidRoles 用于参数校验。
var ValidRoles = map[string]bool{
	RoleSuperAdmin:  true,
	RoleDeviceAdmin: true,
	RoleDean:        true,
	RoleDepartment:  true,
	RoleEngineer:    true,
}

// RoleText 角色展示文本。
func RoleText(role string) string {
	switch role {
	case RoleSuperAdmin:
		return "系统管理员"
	case RoleDeviceAdmin:
		return "设备科"
	case RoleDean:
		return "院长"
	case RoleDepartment:
		return "科室"
	case RoleEngineer:
		return "工程师"
	default:
		return role
	}
}
