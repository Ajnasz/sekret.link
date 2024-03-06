package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/stretchr/testify/mock"
)

type MockGetEntryView struct {
	mock.Mock
}

func (m *MockGetEntryView) RenderReadEntry(w http.ResponseWriter, r *http.Request, entry *services.Entry, key string) {
	m.Called(w, r, entry, key)
}

func (m *MockGetEntryView) RenderReadEntryError(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

func TestGetHandle(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	viewMock := new(MockGetEntryView)
	handler := NewGetHandler(db, viewMock)

	// viewMock.On("RenderReadEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	viewMock.On("RenderReadEntryError", mock.Anything, mock.Anything, mock.Anything).Return()

	request := httptest.NewRequest("GET", "http://example.com/foo", nil)
	response := httptest.NewRecorder()

	handler.Handle(response, request)

	viewMock.AssertExpectations(t)
}
