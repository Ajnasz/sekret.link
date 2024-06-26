package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/parsers"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGetEntryView struct {
	mock.Mock
}

func (m *MockGetEntryView) Render(w http.ResponseWriter, r *http.Request, entry views.EntryReadResponse) {
	m.Called(w, r, entry)
}

func (m *MockGetEntryView) RenderError(w http.ResponseWriter, r *http.Request, err error) {
	m.Called(w, r, err)
}

type GetEntryParserMock struct {
	mock.Mock
}

func (g *GetEntryParserMock) Parse(u *http.Request) (parsers.GetEntryRequestData, error) {
	args := g.Called(u)
	return args.Get(0).(parsers.GetEntryRequestData), args.Error(1)
}

type GetEntryManagerMock struct {
	mock.Mock
}

func (g *GetEntryManagerMock) ReadEntry(ctx context.Context, UUID string, k key.Key) (*services.Entry, error) {
	args := g.Called(ctx, UUID, k)
	return args.Get(0).(*services.Entry), args.Error(1)
}

func TestGetHandle(t *testing.T) {
	viewMock := new(MockGetEntryView)
	parserMock := new(GetEntryParserMock)
	managerMock := new(GetEntryManagerMock)

	handler := NewGetHandler(parserMock, managerMock, viewMock)

	viewMock.On("Render", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	k, err := key.NewGeneratedKey()
	assert.NoError(t, err)
	parserMock.On("Parse", mock.Anything).Return(parsers.GetEntryRequestData{
		UUID:      "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
		KeyString: "12121212aeadf",
		Key:       *k,
	}, nil)

	managerMock.On("ReadEntry", mock.Anything, "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb", *k).
		Return(&services.Entry{
			Data: []byte{18, 18, 18, 18, 174, 173, 15},
			EntryMeta: services.EntryMeta{
				UUID:           "a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb",
				RemainingReads: 0,
				DeleteKey:      "12121212aeadf",
				Created:        time.Now().Add(time.Minute * -1),
				Accessed:       time.Now(),
				Expire:         time.Now().Add(time.Minute * 1),
			},
		}, nil)

	request := httptest.NewRequest("GET", "http://example.com/foo/a6a9d8cc-db7f-11ee-8f4f-3b41146b31eb/12121212aeadf", nil)
	response := httptest.NewRecorder()

	handler.Handle(response, request)
	viewMock.AssertExpectations(t)
}
