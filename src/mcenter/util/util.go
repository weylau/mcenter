package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"mcenter/config"
	"net/http"
	"os"
	"strings"
	"time"
)

func LogOnString(content string) {
	if content != "" {
		content := fmt.Sprintf(" Log - %s", content)
		WriteLog(getLogDir(), content)
	}
}

func LogOnError(err interface{}) {
	if err != nil {
		errinfo := fmt.Sprintf(" ERROR - %s", err)
		WriteLog(getLogDir(), errinfo)
	}
}

func FailOnErr(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

func getLogDir() string {
	fileName := "mcenter" + time.Now().Format("20060102") + ".log"
	folderPath := config.LogDir
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步：先创建文件夹、再修改权限
		os.Mkdir(folderPath, 0777) //0777也可以os.ModePerm
		os.Chmod(folderPath, 0777)
	}
	return folderPath + fileName
}

func WriteLog(name, content string) {
	fd, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	fd_content := strings.Join([]string{fd_time, content, "\n"}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()

}

func BytesToString(b *[]byte) *string {
	s := bytes.NewBuffer(*b)
	r := s.String()
	return &r
}

func NewHttpClient(maxIdleConns, maxIdleConnsPerHost, idleConnTimeout int) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second,
	}

	client := &http.Client{
		Transport: tr,
	}

	return client
}

func CloneToPublishMsg(msg *amqp.Delivery) *amqp.Publishing {
	newMsg := amqp.Publishing{
		Headers: msg.Headers,

		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,

		Body: msg.Body,
	}

	return &newMsg
}

func SetupChannel(mqHost string) (*amqp.Connection, *amqp.Channel, error) {

	conn, err := amqp.Dial(mqHost)
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}

	err = channel.Qos(1, 0, false)
	if err != nil {
		FailOnErr(err, "")
		return nil, nil, err
	}
	return conn, channel, nil
}

func Md5(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	return hex.EncodeToString(w.Sum(nil))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}
