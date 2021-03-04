package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Products []Product

type ProductHandler struct {
	sync.Mutex
	products Products
}

func (ph *ProductHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ph.get(w, r)
	case "POST":
		ph.post(w, r)
	case "PUT", "PATCH":
		ph.put(w, r)
	case "DELETE":
		ph.delete(w, r)
	default:
		ResponseWithError(w, http.StatusMethodNotAllowed, "invalid method")
	}
}
func ResponseWithError(w http.ResponseWriter, code int, msg string) {
	ResponseWithJson(w, code, map[string]string{"error": msg})
}

func (ph *ProductHandler) get(w http.ResponseWriter, r *http.Request) {
	defer ph.Unlock()
	ph.Lock()
	id, err := IdFromURL(r)
	if err != nil {
		ResponseWithJson(w, http.StatusOK, ph.products)
		return
	}
	if id >= len(ph.products) || id < 0 {
		ResponseWithError(w, http.StatusNotFound, "doesn't exist")
		return
	}
	ResponseWithJson(w, http.StatusOK, ph.products[id])
	return
}
func (ph *ProductHandler) post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, err.Error())
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		ResponseWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json required")
		return
	}
	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		ResponseWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer ph.Unlock()
	ph.Lock()
	ph.products = append(ph.products, product)
	ResponseWithJson(w, http.StatusCreated, product)
	return
}
func (ph *ProductHandler) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := IdFromURL(r)
	if err != nil {
		ResponseWithError(w, http.StatusNotFound, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, err.Error())
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		ResponseWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json required")
		return
	}
	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		ResponseWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer ph.Unlock()
	ph.Lock()
	if id >= len(ph.products) || id < 0 {
		ResponseWithError(w, http.StatusNotFound, "doesn't exist")
		return
	}
	if product.Name != "" {
		ph.products[id].Name = product.Name
	}
	if product.Price != 0.0 {
		ph.products[id].Price = product.Price
	}
	ResponseWithJson(w, http.StatusOK, ph.products[id])
	return
}

func (ph *ProductHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := IdFromURL(r)
	if err != nil {
		ResponseWithError(w, http.StatusNotFound, "doesn't exist")
		return
	}
	defer ph.Unlock()
	ph.Lock()
	if id >= len(ph.products) || id < 0 {
		ResponseWithError(w, http.StatusNotFound, "doesn't exist")
		return
	}
	if id < len(ph.products)-1 {
		ph.products[len(ph.products)-1], ph.products[id] = ph.products[id], ph.products[len(ph.products)-1]
	}
	ph.products = ph.products[:len(ph.products)-1]
	ResponseWithJson(w, http.StatusNoContent, "")
}

func IdFromURL(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		return 0, errors.New("not found")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, errors.New("not id")
	}
	return id, nil
}

func ResponseWithJson(w http.ResponseWriter, code int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		products: Products{
			Product{"Shoes", 25.00},
			Product{"Short", 10.00},
			Product{"Cam", 40.00},
			Product{"Mouse", 30.00},
			Product{"WebCam", 20.00},
		},
	}
}

func main() {
	port := ":8080"
	ph := NewProductHandler()
	http.Handle("/products", ph)
	http.Handle("/products/", ph)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello word \n")
	})
	fmt.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
