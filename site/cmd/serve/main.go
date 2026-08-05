// Command serve is a tiny static file server for site/dist — replaces
// the old HTML generator's -serve flag now that docsapp is the site.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":3030", "listen address")
	dir := flag.String("dir", "dist", "directory to serve")
	flag.Parse()
	fmt.Println("serving", *dir, "on", *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))))
}
