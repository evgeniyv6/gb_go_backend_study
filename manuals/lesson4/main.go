package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	ex5()
}

// exrcise 1
func ex1() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ya got the wrong place, pal")
	})
	http.ListenAndServe(":6091", nil)
}

// exercise 2

//type hiHandler struct {
//	subj string
//}
//
//func (h *hiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	fmt.Fprintf(w, "Hello, %s", h.subj)
//}
//
//func ex2() {
//	w := &hiHandler{"World"}
//	r := &hiHandler{"Hello"}
//
//	srv := &http.Server{
//		Addr: ":6091",
//		ReadTimeout: 10 * time.Second,
//		WriteTimeout: 10 * time.Second,
//	}
//
//	http.Handle("/world", w)
//	http.Handle("/hello", r)
//	srv.ListenAndServe()
//}

// exercise 3
//type Handler struct {
//
//}
//
//func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	switch r.Method {
//	case http.MethodGet:
//		name := r.FormValue("name")
//		fmt.Fprintf(w, "Parsed query-param with key \"name\": %s", name)
//	case http.MethodPost:
//		//body, err := ioutil.ReadAll(r.Body)
//		//if err != nil {
//		//	http.Error(w, "Unable to read request body", http.StatusBadRequest)
//		//	return
//		//}
//		var employee Employee
//
//		contentType := r.Header.Get("Content-Type")
//		switch contentType {
//		case "application/json":
//			err := json.NewDecoder(r.Body).Decode(&employee)
//			if err != nil {
//				http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
//				return
//			}
//		default:
//			http.Error(w, "Unknown content type", http.StatusBadRequest)
//			return
//		}
//
//		// err = json.Unmarshal(body, &employee)
//		//err := json.NewDecoder(r.Body).Decode(&employee)
//		//if err != nil {
//		//	http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
//		//	return
//		//}
//
//		defer r.Body.Close()
//		fmt.Fprintf(w, "Parsed request body!!!!: %s %d\n", employee.Name,employee.Age)
//	}
//}

//func ex3() {
//	h := &Handler{}
//	http.Handle("/", h)
//	srv := &http.Server{
//		Addr: ":6091",
//		ReadTimeout: 10 * time.Second,
//		WriteTimeout: 10 * time.Second,
//	}
//	srv.ListenAndServe()
//}

// example 4
//type Employee struct {
//	Name string `json:"name"`
//	Age int `json:"age"`
//	Salary float32 `json:"salary"`
//}

// example5

type UploadHandlerM struct {
	UploadDir string
}

func ex5() {
	uploadHandler := &UploadHandlerM{
		UploadDir: "upload",
	}

	http.Handle("/upload", uploadHandler)
	srv := &http.Server{
		Addr:         ":6091",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srv.ListenAndServe()
}

func (h *UploadHandlerM) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	filePath := h.UploadDir + "/" + header.Filename
	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "File %s has been successfully uploaded", header.Filename)
}
