package api

import (
	"context"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGenerateEntryKeyView struct {
	mock.Mock
}

func (m *MockGenerateEntryKeyView) Render(w http.ResponseWriter, r *http.Request, entry views.GenerateEntryKeyResponseData) {
	m.Called(w, r, entry)
}

func (m *MockGenerateEntryKeyView) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

type MockGenerateEntryKeyManager struct {
	mock.Mock
}

func (m *MockGenerateEntryKeyManager) GenerateEntryKey(ctx context.Context, UUID string, key []byte) (*services.EntryKeyData, error) {
	args := m.Called(ctx, UUID, key)
	return args.Get(0).(*services.EntryKeyData), args.Error(2)
}

type MockGenerateEntryKeyParser struct {
	mock.Mock
}

func (g *MockGenerateEntryKeyParser) Parse(u *http.Request) (parsers.GenerateEntryKeyRequestData, error) {
	args := g.Called(u)
	return args.Get(0).(parsers.GenerateEntryKeyRequestData), args.Error(1)
}

func TestGenerateEntryKey_Handle(t *testing.T) {
	viewMock := new(MockGenerateEntryKeyView)
	parserMock := new(MockGenerateEntryKeyParser)
	managerMock := new(MockGenerateEntryKeyManager)

	handler := NewGenerateEntryKeyHandler(parserMock, managerMock, viewMock)

	newKey := []byte{18, 18, 18, 18, 174, 173, 15}
	expire := time.Now().Add(time.Hour * 24)

	viewMock.On("Render", mock.Anything, mock.Anything, views.GenerateEntryKeyResponseData{
		UUID:   "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
		Key:    hex.EncodeToString(newKey),
		Expire: expire,
	}).Return()
	parserMock.On("Parse", mock.Anything).Return(parsers.GenerateEntryKeyRequestData{
		UUID: "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
		Key:  []byte{18, 18, 18, 18, 174, 173, 15},
	}, nil)

	managerMock.On("GenerateEntryKey", mock.Anything, "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb", []byte{18, 18, 18, 18, 174, 173, 15}).Return(&services.EntryKeyData{
		Expire:    expire,
		EntryUUID: "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
		KEK:       newKey,
	}, newKey, nil)

	handler.Handle(nil, nil)
	managerMock.AssertExpectations(t)
	parserMock.AssertExpectations(t)
	viewMock.AssertExpectations(t)
}

func TestGenerateEntryKey_HandleParseError(t *testing.T) {
	viewMock := new(MockGenerateEntryKeyView)
	parserMock := new(MockGenerateEntryKeyParser)
	managerMock := new(MockGenerateEntryKeyManager)

	handler := NewGenerateEntryKeyHandler(parserMock, managerMock, viewMock)

	parserMock.On("Parse", mock.Anything).Return(parsers.GenerateEntryKeyRequestData{}, assert.AnError)

	viewMock.On("RenderError", mock.Anything, mock.Anything, mock.Anything).Return()
	handler.Handle(nil, nil)
	managerMock.AssertExpectations(t)
	parserMock.AssertExpectations(t)
	viewMock.AssertExpectations(t)
}

func TestGenerateEntryKey_HandleManagerError(t *testing.T) {
	viewMock := new(MockGenerateEntryKeyView)
	parserMock := new(MockGenerateEntryKeyParser)
	managerMock := new(MockGenerateEntryKeyManager)

	handler := NewGenerateEntryKeyHandler(parserMock, managerMock, viewMock)

	viewMock.On("RenderError", mock.Anything, mock.Anything, mock.Anything).Return()
	parserMock.On("Parse", mock.Anything).Return(parsers.GenerateEntryKeyRequestData{
		UUID: "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
		Key:  []byte{18, 18, 18, 18, 174, 173, 15},
	}, nil)

	managerMock.On("GenerateEntryKey", mock.Anything, "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb", []byte{18, 18, 18, 18, 174, 173, 15}).Return(&services.EntryKeyData{}, []byte{}, assert.AnError)

	handler.Handle(nil, nil)
	managerMock.AssertExpectations(t)
	parserMock.AssertExpectations(t)
	viewMock.AssertExpectations(t)
}
