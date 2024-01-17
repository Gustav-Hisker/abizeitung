package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func getTeachers() []string {
	// Fetch HTML content
	resp, err := http.Get("https://www.nepomucenum.de/wir-am-nepo/lehrende/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read HTML content
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Define the regex pattern
	pattern := `<h3[^>]\s*class\s*=\s*["']\s*team-member-name\s*["'][^>]*>(.*?)<\/h3>`
	re := regexp.MustCompile(pattern)

	// Find matches in HTML
	matches := re.FindAllStringSubmatch(string(html), -1)

	// Extract and print matched strings
	ret := []string{}
	for _, match := range matches {
		ret = append(ret, match[1])
	}
	return ret
}

var teachers = getTeachers()

func Teachers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, strings.Join(teachers, ", "))
}

func getJson() map[string]map[string]int {
	data := map[string]map[string]int{}
	fileContent, _ := os.ReadFile("results1.json")
	json.Unmarshal(fileContent, &data)
	return data
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func TeacherRatingUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	print(ps)
}

func main() {
	results1 := getJson()

	for category, ratings := range results1 {
		print(category + ":\t")
		for _, t := range teachers {
			print(strconv.Itoa(ratings[t]) + "\t")
		}
		println()
	}
	println()

	getTeachers()

	router := httprouter.New()
	router.POST("/lehrerranking", TeacherRatingUpload)
	router.GET("/lehrerranking/validate", Index)
	router.GET("/lehrerranking/results", Index)
	router.GET("/lehrerranking/lehrer", Teachers)

	log.Fatal(http.ListenAndServe(":1337", router))
}
