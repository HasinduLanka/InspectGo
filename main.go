package main

import (
	"log"
	"net/http"

	"github.com/HasinduLanka/InspectGo/api"
)

// This source file faciliates a simple web server that serves static files and handles requests to the /api/hello endpoint.
// This will not be executed on Vercel.

func main() {

	multiplexer := &http.ServeMux{}

	// serve static files
	multiplexer.Handle("/", http.FileServer(http.Dir("./frontend/public")))

	multiplexer.HandleFunc("/api/hello", api.HelloEndpoint)
	multiplexer.HandleFunc("/api/inspect", api.InspectEndpoint)

	log.Println("Listening on port 20000. Visit http://localhost:20000 if you're running this locally.")

	// Blocks until the program is terminated
	serveErr := http.ListenAndServe(":20000", multiplexer)

	// serveErr is always non nil
	log.Println(serveErr.Error())

}
