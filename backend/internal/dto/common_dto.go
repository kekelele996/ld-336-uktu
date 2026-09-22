package dto

// IDParam 路径参数 ID。
type IDParam struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// PageQuery 分页查询参数。
type PageQuery struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}
