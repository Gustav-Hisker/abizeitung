package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func TeacherRatingUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	print(ps)
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/", TeacherRatingUpload)

	log.Fatal(http.ListenAndServe(":8080", router))
}
