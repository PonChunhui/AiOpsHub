package service

type ErrorCode int

const (
	Success ErrorCode = 0

	EntityNotFound      ErrorCode = 1000
	EntityAlreadyExists ErrorCode = 1001

	ToolNotFound          ErrorCode = 2000
	ToolNotEnabled        ErrorCode = 2001
	ToolExecutionFailed   ErrorCode = 2002
	ToolCommandNotAllowed ErrorCode = 2003
	ToolHostNotAllowed    ErrorCode = 2004

	AgentNotFound          ErrorCode = 3000
	AgentNotEnabled        ErrorCode = 3001
	AgentToolBindingFailed ErrorCode = 3002

	InvalidParameter ErrorCode = 4000
	MissingParameter ErrorCode = 4001

	DatabaseError ErrorCode = 5000
	InternalError ErrorCode = 5001
)

type ServiceError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewServiceError(code ErrorCode, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func IsServiceError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

func GetServiceError(err error) *ServiceError {
	if se, ok := err.(*ServiceError); ok {
		return se
	}
	return NewServiceError(InternalError, "内部错误", err)
}
