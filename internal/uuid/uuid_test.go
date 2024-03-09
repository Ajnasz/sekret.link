package uuid

import "testing"

func TestGetUUIDFromPath(t *testing.T) {
	testCases := []struct {
		Name           string
		Value          string
		ExpectedUUID   string
		ExpectedSecret string
	}{
		{
			"simple uuid",
			"/3f356f6c-c8b1-4b48-8243-aa04d07b8873/secret",
			"3f356f6c-c8b1-4b48-8243-aa04d07b8873",
			"secret",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actualUUID, actualSecret, err := GetUUIDAndSecretFromPath(testCase.Value)
			if err != nil {
				t.Fatal(err)
			}
			if testCase.ExpectedUUID != actualUUID {
				t.Errorf("expected: %q, actual: %q", testCase.ExpectedUUID, actualUUID)
			}

			if testCase.ExpectedSecret != actualSecret {
				t.Errorf("expected: %q, actual: %q", testCase.ExpectedSecret, actualSecret)
			}
		})
	}
}
