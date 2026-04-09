package dto

type ServiceError struct {
    Code    string
    Message string
    Field   string
}

func (e *ServiceError) Error() string {
    return e.Message
}

func NewServiceError(code string, message string) *ServiceError {
    return &ServiceError{Code: code, Message: message}
}

func NewFieldError(code string, message string, field string) *ServiceError {
    return &ServiceError{Code: code, Message: message, Field: field}
}