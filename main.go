package main

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	http.HandleFunc("/", handlerProxy)

	if err := http.ListenAndServe(":8083", nil); err != nil {
		panic(err)
	}
}

func handlerProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == http.MethodOptions {
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
	query.Add("account", "cajivivi77")
	r.URL.RawQuery = query.Encode()

	u, _ := url.Parse("https://api.cdnvideo.ru/app/statistic/v3")
	proxy := httputil.NewSingleHostReverseProxy(u)
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
