package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	folder       = "/tmp/"
	uploaderPort = ":6091"
	serverPort   = ":6092"
)

type UploadHandler struct {
	Host string
	Dir  string
}

var (
	done  = make(chan bool, 2)
	sugar *zap.SugaredLogger
)

func (u *UploadHandler) FileWalker(w *http.ResponseWriter, ext string) error {
	err := filepath.WalkDir(u.Dir, func(path string, entry os.DirEntry, err error) error {
		info, ierr := entry.Info()
		if ierr != nil {
			sugar.Error(err)
			//return err
		}
		if (!info.IsDir() && ext == "") || (filepath.Ext(path) == "."+ext) {
			fmt.Fprintf(*w, "Path: %s, Name: %s, Size: %d, Ext: %s\n", path, info.Name(), info.Size(), filepath.Ext(path))
		}
		return nil
	})
	if err != nil {
		sugar.Error(err)
		//return err
	}
	return nil
}

func (u *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ext := r.FormValue("ext")
		err := u.FileWalker(&w, ext)
		if err != nil {
			sugar.Error(err)
			http.Error(w, "Couldnot print a list of files", http.StatusInternalServerError)
		}
		fmt.Println(w.Header())
	case http.MethodPost:
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Cannot read file", http.StatusBadRequest)
		}

		defer func() {
			err := file.Close()
			if err != nil {

			}
		}()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Cannot read file", http.StatusBadRequest)
		}

		filePath := u.Dir + fmt.Sprintf("/%s", header.Filename)

		err = ioutil.WriteFile(filePath, data, 0777)

		if err != nil {
			sugar.Error(err)
			http.Error(w, "Cannot save file", http.StatusInternalServerError)
			return
		}
		_, err = fmt.Fprintf(w, "File %s has been successfully uploaded.\nLink: ", header.Filename)
		if err != nil {
			sugar.Error(err)
			return
		}

		err = ioutil.WriteFile(filePath, data, 0777)
		if err != nil {
			sugar.Error(err)
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		fileLink := u.Host + "/" + header.Filename
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
	up := &UploadHandler{
		Host: "http://localhost:6092",
		Dir:  folder,
	}

	http.Handle("/tmp", up)

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
	up := &UploadHandler{
		Dir: folder,
	}
	dirToServe := http.Dir(up.Dir)
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
		done <- true
	}
	wg.Wait()
}
