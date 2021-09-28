package main

import (
	"log"
	"net"
	"net/http"
	"os"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func wrapHandlerWithLogging(wrappedHandler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)
		log.Printf("--> [Request] Method:%s Path:%s IP:%s", req.Method, req.URL.Path, ip)

		lrw := NewLoggingResponseWriter(w)
		wrappedHandler.ServeHTTP(lrw, req)

		statusCode := lrw.statusCode
		log.Printf("<-- [Response] %d %s", statusCode, http.StatusText(statusCode))
	})
}

func test0(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing test0")
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.Header().Add("VERSION", os.Getenv("VERSION"))

	w.Write([]byte("test0"))
}

func healthz(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing healthz")
	w.Write([]byte("healthz"))

}

func main() {
	log.Println("NOW VERSION:", os.Getenv("VERSION"))

	http.HandleFunc("/", wrapHandlerWithLogging(test0))
	http.HandleFunc("/healthz", wrapHandlerWithLogging(healthz))
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
