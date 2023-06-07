package main

import (
	"fmt"
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

func SendJSONError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, `{"error":{"msg":%q}}`, msg)
}

func handlerProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusOK)
		return
	}

	if strings.HasPrefix(r.URL.String(), "/token") {
		//Добавить получение токена
		fmt.Println("Get token")
	}

	token := r.URL.Query().Get("token")
	r.Header.Set("CDN-AUTH-TOKEN", token)
	fmt.Println("Token", token)

	query := r.URL.Query()
	query.Del("token")
	r.URL.RawQuery = query.Encode()

	url, err := url.Parse("https://api.cdnvideo.ru/app/statistic/v3" + r.URL.String())

	if err != nil {
		fmt.Println("ERR:", err.Error())
		SendJSONError(w, err.Error())
		return
	}

	fmt.Println("URL:", url)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)
}
