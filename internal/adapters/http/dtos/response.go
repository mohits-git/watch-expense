package dtos

type BaseResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewResponse(status int, message string, data any) BaseResponse {
	return BaseResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
}
