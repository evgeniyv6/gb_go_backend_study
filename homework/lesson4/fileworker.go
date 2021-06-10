package main

import (
	"bytes"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	folder       = "/tmp/"
	uploaderPort = ":6091"
	serverPort   = ":6092"
)

type (
	UploadHandler struct {
		Host string
		Dir  string
	}
	mapInfo map[string][3]string
)

var (
	done  = make(chan bool, 2)
	sugar *zap.SugaredLogger
)

func (m mapInfo) mapPrinter() *bytes.Buffer {
	b := new(bytes.Buffer)
	for k, v := range m {
		_, err := fmt.Fprintf(b, "%s: %s\n", k, v)
		if err != nil {
			sugar.Error(err)
			return nil
		}
	}
	return b
}

func (h *UploadHandler) FilesWalker(ext string) (mapInfo, error) {
	resultMap := make(mapInfo)
	err := filepath.WalkDir(h.Dir, func(path string, entry os.DirEntry, err error) error {
		info, ierr := entry.Info()
		if ierr != nil {
			sugar.Error(err)
			return err
		}
		if (!info.IsDir() && ext == "") || (filepath.Ext(path) == "."+ext) {
			resultMap[path] = [3]string{info.Name(), strconv.FormatInt(info.Size(), 10), filepath.Ext(path)}
		}
		return nil
	})
	if err != nil {
		sugar.Error(err)
		return resultMap, err
	}
	return resultMap, nil
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ext := r.FormValue("ext")
		resMap, err := h.FilesWalker(ext)
		if err != nil {
			sugar.Error(err)
			http.Error(w, "Couldnot print a list of files", http.StatusInternalServerError)
		}
		_, err = fmt.Fprintf(w, resMap.mapPrinter().String())
		if err != nil {
			sugar.Error(err)
		}
	case http.MethodPost:
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Couldnot read file", http.StatusBadRequest)
		}

		defer func() {
			err := file.Close()
			if err != nil {
				sugar.Error(err)
			}
		}()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Couldnot read file", http.StatusBadRequest)
			return
		}

		filePath := h.Dir + fmt.Sprintf("/%s", header.Filename)

		err = ioutil.WriteFile(filePath, data, 0755)

		if err != nil {
			sugar.Error(err)
			http.Error(w, "Couldnot save file", http.StatusInternalServerError)
			return
		}
		_, err = fmt.Fprintf(w, "File %s has been successfully uploaded.\nLink: ", header.Filename)
		if err != nil {
			sugar.Error(err)
			return
		}

		fileLink := h.Host + "/" + header.Filename
		_, err = fmt.Fprintln(w, fileLink)
		if err != nil {
			sugar.Error(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
	}
}

func uploader(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	h := &UploadHandler{
		Host: "http://localhost" + serverPort,
		Dir:  folder,
	}

	http.Handle("/upload", h)

	srv := &http.Server{
		Addr:         uploaderPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		select {
		case d := <-done:
			sugar.Infof("srv get msg from done : %t", d)
			err := srv.Close()
			if err != nil {
				sugar.Error(err)
			}
		}
	}()
	sugar.Infof("Starting uploader.")
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		sugar.Error(err)
	}
}

func server(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	h := &UploadHandler{
		Dir: folder,
	}
	dirToServe := http.Dir(h.Dir)
	fs := &http.Server{
		Addr:         serverPort,
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		select {
		case d := <-done:
			sugar.Infof("fs get msg from done : %t", d)
			err := fs.Close()
			if err != nil {
				sugar.Error(err)
			}
		}
	}()
	sugar.Infof("Starting server.")
	err := fs.ListenAndServe()
	if err != http.ErrServerClosed {
		sugar.Error(err)
	}
}

func main() {
	var (
		wg           sync.WaitGroup
		logger, _    = zap.NewProduction()
		ctx, cancel  = context.WithCancel(context.Background())
		cancelSignal = make(chan os.Signal, 1)
		catchSignal  = func(cancelFunc context.CancelFunc) {
			sig := <-cancelSignal
			sugar.Infof("Received stop signal - %v", sig)
			cancelFunc()
		}
	)
	defer logger.Sync()
	sugar = logger.Sugar()

	signal.Notify(cancelSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go catchSignal(cancel)
	go uploader(&wg)
	go server(&wg)

	select {
	case <-ctx.Done():
		sugar.Infof("Stop signal. Ctx canceled.")
		done <- true
		done <- true // for uploader and server
	}
	wg.Wait()
}
