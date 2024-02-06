package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// Parse the target URL
	targetURL, err := url.Parse("https://go.dev/doc/")
	if err != nil {
		log.Fatalf("Error parsing target URL: %s", err)
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Start a server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set the host header to the target host
		r.Host = targetURL.Host

		// Serve the reverse proxy
		proxy.ServeHTTP(w, r)
	})

	log.Println("Backend server started. Listening on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}



// package main

// import (
//     "fmt"
//     "log"
//     "net/http"
// )

// func main() {
//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         fmt.Fprintf(w, "This is the backend server responding to your request!")
//     })

//     log.Println("Backend server started. Listening on port 8081...")
//     log.Fatal(http.ListenAndServe(":8081", nil))
// }
