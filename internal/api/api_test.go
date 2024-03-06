package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockParser struct {
	mock.Mock
}

func (m *MockParser) Parse(r *http.Request) (*parsers.CreateEntryRequestData, error) {
	args := m.Called(r)

	return args.Get(0).(*parsers.CreateEntryRequestData), args.Error(1)
}

type MockEntryManager struct {
	mock.Mock
}

func (m *MockEntryManager) CreateEntry(
	ctx context.Context,
	body []byte,
	maxReads int,
	expiration time.Duration,
) (*services.EntryMeta, error) {
	args := m.Called(ctx, body, maxReads, expiration)

	return args.Get(0).(*services.EntryMeta), args.Error(1)
}

type MockEntryView struct {
	mock.Mock
}

func (m *MockEntryView) RenderEntryCreated(w http.ResponseWriter, r *http.Request, entry *services.EntryMeta, key string) {
	m.Called(w, r, entry, key)
}

func (m *MockEntryView) RenderCreateEntryErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

func Test_CreateEntryHandle(t *testing.T) {
	k, err := key.NewGeneratedKey()

	if err != nil {
		t.Fatal(err)
	}

	data := bytes.NewBufferString("This is a test")

	parser := new(MockParser)
	entryManager := new(MockEntryManager)
	view := new(MockEntryView)

	request := httptest.NewRequest("POST", "http://example.com/foo", data)
	response := httptest.NewRecorder()

	parser.On("Parse", request).Return(&parsers.CreateEntryRequestData{}, nil)
	entryManager.On("CreateEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&services.EntryMeta{}, nil)
	view.On("RenderEntryCreated", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewCreateHandler(10, parser, entryManager, view, k)

	handler.Handle(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	parser.AssertExpectations(t)
	entryManager.AssertExpectations(t)
	view.AssertExpectations(t)
}

// on parser.Parse error, view.RenderCreateEntryErrorResponse should be called
func Test_CreateEntryHandleParserError(t *testing.T) {
	k, err := key.NewGeneratedKey()

	if err != nil {
		t.Fatal(err)
	}

	data := bytes.NewBufferString("This is a test")

	parser := new(MockParser)
	entryManager := new(MockEntryManager)
	view := new(MockEntryView)

	request := httptest.NewRequest("POST", "http://example.com/foo", data)
	response := httptest.NewRecorder()

	parser.On("Parse", request).Return(&parsers.CreateEntryRequestData{}, errors.New("error"))
	view.On("RenderCreateEntryErrorResponse", mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewCreateHandler(10, parser, entryManager, view, k)

	handler.Handle(response, request)

	parser.AssertExpectations(t)
	entryManager.AssertExpectations(t)
	view.AssertExpectations(t)
}

// On entryManager.CreateEntry error, view.RenderCreateEntryErrorResponse should be called
func Test_CreateEntryHandleError(t *testing.T) {
	k, err := key.NewGeneratedKey()

	if err != nil {
		t.Fatal(err)
	}

	data := bytes.NewBufferString("This is a test")

	parser := new(MockParser)
	entryManager := new(MockEntryManager)
	view := new(MockEntryView)

	request := httptest.NewRequest("POST", "http://example.com/foo", data)
	response := httptest.NewRecorder()

	parser.On("Parse", request).Return(&parsers.CreateEntryRequestData{}, nil)
	entryManager.On("CreateEntry", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&services.EntryMeta{}, errors.New("error"))
	view.On("RenderCreateEntryErrorResponse", mock.Anything, mock.Anything, mock.Anything).Return()

	handler := NewCreateHandler(10, parser, entryManager, view, k)

	handler.Handle(response, request)

	parser.AssertExpectations(t)
	entryManager.AssertExpectations(t)
	view.AssertExpectations(t)
}
