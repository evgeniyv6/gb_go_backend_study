package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

var (
	testFolders = "./testfolder/inner/"
	file1       = "./testfolder/f1.txt"
	innerFile2  = "./testfolder/inner/f2.txt"
	AppFsOs     = afero.NewOsFs()
	AppFsMem    = afero.NewMemMapFs()
	afsOs       = &afero.Afero{Fs: AppFsOs}
	afsMem      = &afero.Afero{Fs: AppFsMem}
)

func init() {
	err := afsOs.MkdirAll(testFolders, 0755)
	if err != nil {
		log.Println(err)
	}

	err = afsMem.MkdirAll(testFolders, 0755)
	if err != nil {
		log.Println(err)
	}

	_, err = afsOs.Create(file1)
	if err != nil {
		log.Println(err)
	}

	_, err = afsMem.Create(file1)
	if err != nil {
		log.Println(err)
	}

	err = afero.WriteFile(afsOs, innerFile2, []byte("test file"), 0755)
	if err != nil {
		log.Println(err)
	}
}

func TestUploadHandler_ServeHTTP(t *testing.T) {
	file, err := afsMem.Open(file1)
	if err != nil {
		t.Errorf("Open file error: %s", err)
	}
	defer file.Close()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		t.Errorf("CreateFormFile err: %s", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		t.Errorf("Open file error: %s", err)
	}
	err = w.Close()
	if err != nil {
		log.Println(err)
	}

	r, _ := http.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Add("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		Host: ts.URL,
		Dir:  testFolders,
	}

	uploadHandler.ServeHTTP(rr, r)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `f1.txt`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestUploadHandler_FilesWalker(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/tmp?ext=txt", &bytes.Buffer{})
	rr := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		Host: ts.URL,
		Dir:  testFolders,
	}

	expected := "testfolder/inner/f1.txt: [f1.txt 0 .txt]\ntestfolder/inner/f2.txt: [f2.txt 9 .txt]"
	uploadHandler.ServeHTTP(rr, r)
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
