package qcsdk

import (
	"os"
	"testing"
)

var api *Api

func init() {
	ak := os.Getenv("ACCESS_KEY_ID")
	sk := os.Getenv("SECRET_KEY")
	region := os.Getenv("ZONE")
	if ak == "" || sk == "" || region == "" {
		panic("environment variable ACCESS_KEY_ID, SECRET_KEY and ZONE must be provided")
	}
	api = NewApi(ak, sk, region)
	api.SetDebug(true)
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func assertStringEqual(t *testing.T, str1, str2 string) {
	if str1 != str2 {
		t.Fatalf("expected %s, got %s", str1, str2)
	}
}
