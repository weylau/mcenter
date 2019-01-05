package util

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestCrypt(t *testing.T) {

	key := "1234567"
	has := Md5(key)[16:]

	fmt.Println("has:", has)
	content := "{\"code\":1,\"msg\":\"ok\"}"
	encrypt, err := AesEncrypt([]byte(content), []byte(has))
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("base64 str : %s\n", base64.StdEncoding.EncodeToString(encrypt))
	decrypt, err := AesDecrypt(encrypt, []byte(has))

	decryptContent := string(decrypt[:])

	if decryptContent != content {
		t.Error("解密失败:", decryptContent)
		return
	}

	fmt.Println("解密成功")

}
