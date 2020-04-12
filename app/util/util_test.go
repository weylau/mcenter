package util

import (
	"encoding/base64"
	"fmt"
	"mcenter"
	"testing"
)

func TestCrypt(t *testing.T) {

	key := "1234567"
	has := mcenter.Md5(key)[16:]

	fmt.Println("has:", has)
	content := "{\"test\":\"test\"}"
	encrypt, err := mcenter.AesEncrypt([]byte(content), []byte(has))
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("base64 str : %s\n", base64.StdEncoding.EncodeToString(encrypt))
	decrypt, err := mcenter.AesDecrypt(encrypt, []byte(has))

	decryptContent := string(decrypt[:])

	if decryptContent != content {
		t.Error("解密失败:", decryptContent)
		return
	}

	fmt.Println("解密成功")

}
