package constants

// 全局错误码集中维护。
const (
	CodeOK               = 0
	CodeBadRequest       = 40000
	CodeUnauthorized     = 40100
	CodeForbidden        = 40300
	CodeNotFound         = 40400
	CodeConflict         = 40900
	CodeValidation       = 42200
	CodeInternalError    = 50000
	CodeInvalidToken     = 40101
	CodeTokenExpired     = 40102
	CodeRateLimited      = 42900
	CodeInvalidStatus    = 40901
	CodeDuplicateUser    = 40902
	CodeDuplicateDevice  = 40903
	CodeDuplicateRequest = 40904
	CodeDuplicateCalib   = 40905
	CodeWrongPassword    = 40103
	CodeUserDisabled     = 40301
	CodeDeviceNotAllowed = 40906
)
