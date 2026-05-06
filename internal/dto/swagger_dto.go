package dto

// ErrorResponse đại diện cho cấu trúc báo lỗi
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
	Field string `json:"field,omitempty"`
}

// SuccessResponse đại diện cho cục data trả về
type SuccessResponse[T any] struct {
	Data T `json:"data"`
}

// PaginatedResponse đại diện cho trả về có phân trang
type PaginatedResponse[T any] struct {
	Status     string        `json:"status"`
	Pagination PaginationDTO `json:"pagination"`
	Data       T             `json:"data"`
}

// MessageResponse đại diện cho trả về chỉ có status và message
type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
