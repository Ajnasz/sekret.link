package storage

import (
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/testhelper"
	"github.com/Ajnasz/sekret.link/uuid"
)

func TestSecretStorage(t *testing.T) {

	testData := "Lorem ipusm dolor sit amet"
	connection := ConnectToPostgresql(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})
	psqlStorage := PostgresCleanableStorage{connection}
	storage := &CleanableSecretStorage{
		NewSecretStorage(
			psqlStorage,
			NewDummyEncrypter(),
		),
		psqlStorage,
	}
	// TODO defer storage.Close()

	UUID := uuid.NewUUIDString()
	err := storage.Create(UUID, []byte(testData), time.Second*10, 1)

	if err != nil {
		t.Fatal(err)
	}

	data, err := storage.GetAndDelete(UUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(data.Data)

	if actual != testData {
		t.Errorf("Expected %q, actual %q", testData, actual)
	}
}
