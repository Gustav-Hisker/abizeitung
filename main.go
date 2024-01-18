package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// structs and classes

type Question struct {
	Question string `json:"question"`
	Best     string `json:"best"`
	Worst    string `json:"worst"`
}

type RankedTeacher struct {
	name  string `json:"name"`
	rank  int    `json:"rank"`
	score int    `json:"score"`
}

// loading functions

func getTeachers() []string {
	resp, err := http.Get("https://www.nepomucenum.de/wir-am-nepo/lehrende/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	pattern := `<h3[^>]\s*class\s*=\s*["']\s*team-member-name\s*["'][^>]*>(.*?)<\/h3>`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(string(html), -1)

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

func getResults() map[string]map[string]map[string]int {
	data := map[string]map[string]map[string]int{}
	fileContent, _ := os.ReadFile("results1.json")
	json.Unmarshal(fileContent, &data)
	return data
}

func getVoters() map[string]bool {
	data := map[string]bool{}
	fileContent, _ := os.ReadFile("voters.json")
	json.Unmarshal(fileContent, &data)
	return data
}

func saveVoters() {
	data, err := json.MarshalIndent(voters, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("./voter.json", data, 0644)
}

// declaring variables

var teachers = getTeachers()
var questions = getQuestions()
var results = getResults()
var voters = getVoters()

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
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	if r.PostFormValue("name") == "" {
		w.WriteHeader(401)
		return
	}

	for name := range questions {
		println(name + "\tbest:" + r.PostFormValue(name+"-best") + "\tworst:" + r.PostFormValue(name+"-worst"))

	}
}

func Results(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(data))
}

func ResultsOfTeacher(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	teacherName := ps.ByName("teacher")
	res := map[string]map[string]int{}
	for category, categoryResults := range results {
		if categoryResults[teacherName] == nil {
			w.WriteHeader(400)
			return
		}
		res[category] = categoryResults[teacherName]
	}
	data, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(data))
}

func Categories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	res := map[string][]RankedTeacher{}
	for category, categoryResult := range results {
		for teacher, scores := range categoryResult {
			res[category] = append(res[category], RankedTeacher{
				name:  teacher,
				rank:  -1,
				score: scores["b"] - scores["w"],
			})
		}
	}
}

func Category(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

func NotImplemented(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "This page isn't implemented")
}

// main
func main() {
	router := httprouter.New()
	router.POST("/lehrer-ranking", TeacherRatingUpload)

	router.GET("/lehrer-ranking/lehrer", Teachers)
	router.GET("/lehrer-ranking/fragen", Questions)
	router.GET("/lehrer-ranking/ergebnisse", Results)
	router.GET("/lehrer-ranking/ergebnisse/l/:teacher", ResultsOfTeacher)
	router.GET("/lehrer-ranking/ergebnisse/r", Categories)
	router.GET("/lehrer-ranking/ergebnisse/r/:category", Category)

	router.GET("/lehrer-ranking/validate-name", NotImplemented)

	log.Fatal(http.ListenAndServe(":1337", router))
}
