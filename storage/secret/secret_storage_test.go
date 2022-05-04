package secret

import (
	"context"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/encrypter/dummy"
	"github.com/Ajnasz/sekret.link/storage/postgresql"
	"github.com/Ajnasz/sekret.link/testhelper"
	"github.com/Ajnasz/sekret.link/uuid"
)

func TestSecretStorage(t *testing.T) {

	testData := "Lorem ipusm dolor sit amet"
	psqlStorage := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	storage := NewSecretStorage(
		psqlStorage,
		dummy.NewEncrypter(),
	)

	t.Cleanup(func() {
		storage.Close()
	})
	UUID := uuid.NewUUIDString()
	ctx := context.Background()
	err := storage.Write(ctx, UUID, []byte(testData), time.Second*10, 1)

	if err != nil {
		t.Fatal(err)
	}

	data, err := storage.Read(ctx, UUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(data.Data)

	if actual != testData {
		t.Errorf("Expected %q, actual %q", testData, actual)
	}
}
