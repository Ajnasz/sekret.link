package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/stretchr/testify/mock"
)

type MockDeleteEntryManager struct {
	mock.Mock
}

func (m *MockDeleteEntryManager) DeleteEntry(ctx context.Context, UUID string, deleteKey string) error {
	args := m.Called(ctx, UUID, deleteKey)
	return args.Error(0)
}

type MockDeleteEntryView struct {
	mock.Mock
}

func (m *MockDeleteEntryView) RenderDeleteEntry(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
}

func (m *MockDeleteEntryView) RenderDeleteEntryError(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

func Test_DeleteHandle(t *testing.T) {
	entryManager := new(MockDeleteEntryManager)
	view := new(MockDeleteEntryView)

	entryManager.On("DeleteEntry", mock.Anything, "40e7d7d6-db0d-11ee-b9ee-1340bdbad9b2", "delete-key").Return(nil)
	view.On("RenderDeleteEntry", mock.Anything, mock.Anything).Return()

	handler := NewDeleteHandler(entryManager, view)

	request := httptest.NewRequest("DELETE", "http://example.com/40e7d7d6-db0d-11ee-b9ee-1340bdbad9b2/key/delete-key", nil)
	response := httptest.NewRecorder()

	handler.Handle(response, request)

	entryManager.AssertExpectations(t)
	view.AssertExpectations(t)
}

// It should bad request when the uuid in the URL is invalid
func Test_DeleteHandle_InvalidUUID(t *testing.T) {
	entryManager := new(MockDeleteEntryManager)
	view := new(MockDeleteEntryView)

	view.On("RenderDeleteEntryError", mock.Anything, mock.Anything, mock.MatchedBy(func(err error) bool {
		return errors.Is(err, parsers.ErrInvalidUUID)
	})).Return()

	handler := NewDeleteHandler(entryManager, view)

	request := httptest.NewRequest("DELETE", "http://example.com/foo/gggggggg-db0d-11ee-b9ee-1340bdbad9b2/key/delete-key", nil)
	response := httptest.NewRecorder()

	handler.Handle(response, request)

	view.AssertExpectations(t)
}
