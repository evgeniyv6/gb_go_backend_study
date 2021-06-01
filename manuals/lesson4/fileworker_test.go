package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

type osFsMock struct{}

var (
	testFolders = "./testfolder/inner/"
	file1       = "./testfolder/f1.txt"
	file2       = "./testfolder/f2.txt"
	innerFile   = "./testfolder/inner/f3.txt"
	AppFs       = afero.NewOsFs()
	afs         = &afero.Afero{Fs: AppFs}
	fsTest      = osFsMock{}
)

func init() {
	afs.MkdirAll(testFolders, 0755)
	afs.Create(file1)
	afs.Create(file2)
	afs.Create(innerFile)
	afero.WriteFile(afs, file1, []byte("test file"), 0755)
	afero.WriteFile(afs, file2, []byte("test file"), 0755)
	afero.WriteFile(afs, innerFile, []byte("test file"), 0755)
}

func TestUploadHandler_ServeHTTP(t *testing.T) {
	//file, err := os.Open("/Users/evgeniivakhrushev/Downloads/trash/sw.txt")
	file, err := afs.Open(file1)
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
	io.Copy(part, file)
	w.Close()

	r, _ := http.NewRequest(http.MethodPost, "/tmp", body)
	r2, _ := http.NewRequest(http.MethodGet, "/tmp?ext=txt", body)
	rr2 := httptest.NewRecorder()

	r.Header.Add("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		Host: ts.URL,
		Dir:  "/tmp",
	}
	uploadHandler2 := &UploadHandler{
		Host: ts.URL,
		Dir:  "./testfolder",
	}

	uploadHandler.ServeHTTP(rr, r)
	uploadHandler2.ServeHTTP(rr2, r2)
	fmt.Printf("+>>>>>>>>>> %s", rr2.Body.String())

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `f1.txt`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

//func TestUploadHandler_FileWalker(t *testing.T) {
//
//}
