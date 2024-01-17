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

type Question struct {
	Question string `json:"question"`
	Best     string `json:"best"`
	Worst    string `json:"worst"`
}

// loading functions
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

func getQuestions() map[string]Question {
	data := map[string]Question{}
	fileContent, _ := os.ReadFile("questions.json")
	json.Unmarshal(fileContent, &data)
	return data
}

func getResults() map[string]map[string]int {
	data := map[string]map[string]int{}
	fileContent, _ := os.ReadFile("results1.json")
	json.Unmarshal(fileContent, &data)
	return data
}

// declaring of "consts"
var teachers = getTeachers()
var questions = getQuestions()

// response functions
func Teachers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "[\""+strings.Join(teachers, "\", \"")+"\"]")
}

func Questions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := json.Marshal(questions)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(data))
}

func TeacherRatingUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	print(ps)
}

func NotImplemented(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "This page isn't implemented")
}

func main() {
	results1 := getResults()

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
	router.POST("/lehrer-ranking", TeacherRatingUpload)
	router.GET("/lehrer-ranking/lehrer", Teachers)
	router.GET("/lehrer-ranking/fragen", Questions)

	router.GET("/lehrer-ranking/validate", NotImplemented)
	router.GET("/lehrer-ranking/results", NotImplemented)

	log.Fatal(http.ListenAndServe(":1337", router))
}
