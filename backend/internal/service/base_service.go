package service

import (
	"fmt"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
)

type BaseService struct{}

func (s *BaseService) LogInfo(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

func (s *BaseService) LogError(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

func (s *BaseService) LogDebug(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}

func (s *BaseService) HandleError(err error, message string) *ServiceError {
	s.LogError("%s: %v", message, err)
	return GetServiceError(err)
}
