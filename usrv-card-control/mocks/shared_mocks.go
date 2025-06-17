// Package mocks holds mocked interfaces for testing.
package mocks

import (
	"testing"

	mocks "bitbucket.org/kushki/usrv-card-control/mocks/core"
	"github.com/stretchr/testify/mock"
)

// GetMockLogger mock logger for global use.
func GetMockLogger(t *testing.T) *mocks.KushkiLogger {
	t.Helper()
	mockLogger := &mocks.KushkiLogger{}
	mockLogger.On("PrintStructureAsJSON", mock.Anything, mock.Anything).Return()
	mockLogger.On("Trace", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warning", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	return mockLogger
}
