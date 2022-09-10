package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func CheckError(err error) {
	if err != nil && err != io.EOF {
		fmt.Printf("Error: %s", err.Error())
		fmt.Println()
		os.Exit(1)
	}
}

func ComputeMd5(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Open Error")
		return "Unable to open file", err
	}
	fileMd5 := md5.New()
	if _, err := io.Copy(fileMd5, file); err != nil {
		return "Unable to hash File", err
	}
	return hex.EncodeToString(fileMd5.Sum(nil)), nil
}