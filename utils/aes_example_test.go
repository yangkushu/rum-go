package utils

import "testing"

func TestAesEncrypt(t *testing.T) {
	key := []byte("PZgsF5ZwNgN0NjU7")
	text, err := AesEncrypt("12345678901", key)
	if IsFailed(err) {
		t.Error(err)
	}

	t.Log(text)
}
