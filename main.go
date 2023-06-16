package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func UUpdateResponse(r *http.Response) error {
	fmt.Println("U1", r.Header.Values("Access-Control-Allow-Origin"))
	r.Header.Set("Access-Control-Allow-Origin", "*")
	fmt.Println("u2", r.Header.Values("Access-Control-Allow-Origin"))
	return nil
}

func main() {
	var (
		httpAddr string
	)
	flag.StringVar(&httpAddr, "http", "localhost:8083", "The http `address` and port of the service")
	flag.Parse()

	http.HandleFunc("/", HandlerProxy)

	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		panic(err)
	}
}

func HandlerProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/token") {
		getToken(w, r)
		return
	}

	token := r.URL.Query().Get("token")
	r.Header.Set("CDN-AUTH-TOKEN", token)

	query := r.URL.Query()
	query.Del("token")
	r.URL.RawQuery = query.Encode()

	baseUrl := "https://api.cdnvideo.ru/app/statistic/v3"

	if strings.HasPrefix(r.URL.String(), "/accounts") {
		baseUrl = "https://api.cdnvideo.ru/app/inventory/v1"
	}

	u, _ := url.Parse(baseUrl)
	fmt.Println("header", r.Header.Values("Access-Control-Allow-Origin"))

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = func(r *http.Response) error {
		// Заголовок задается потому что при запросе cdn1 не добавляет заголовок, а cdn2 добавляет заголовок Access-Control-Allow-Origin
		r.Header.Set("Access-Control-Allow-Origin", "*")
		return nil
	}
	proxy.ServeHTTP(w, r)
}

func getToken(w http.ResponseWriter, r *http.Request) {
	const hardBodyLimit = 1024

	c := struct {
		Username string
		Password string
	}{}

	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, hardBodyLimit)).Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := url.Values{
		"username": {c.Username},
		"password": {c.Password},
	}

	resp, err := http.PostForm("https://api.cdnvideo.ru/app/oauth/v1/token/", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	defer resp.Body.Close()

	m := struct {
		Token    string      `json:"token"`
		Lifetime json.Number `json:"lifetime"`
	}{}

	err = json.NewDecoder(http.MaxBytesReader(w, resp.Body, hardBodyLimit)).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}
