package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dllFile := "WinDivert.dll"
	sysFile := "WinDivert64.sys"
	_, err := os.Stat(dllFile)
	_, err1 := os.Stat(sysFile)
	if err == nil && err1 == nil {
		return
	}
	url := "https://reqrypt.org/download/WinDivert-1.3.0-MINGW.zip"
	response, err := http.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		log.Fatal("fail to download DLL")
	}
	defer response.Body.Close()
	zipFile, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(zipFile)
	zipReader, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "amd64/WinDivert.dll") {
			saveFile(file)
		}
		if strings.HasSuffix(file.Name, "amd64/WinDivert64.sys") {
			saveFile(file)
		}
	}
}

func saveFile(file *zip.File) {
	f, err := file.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Base(file.Name), data, 0444)
	if err != nil {
		log.Fatal(err)
	}
}
