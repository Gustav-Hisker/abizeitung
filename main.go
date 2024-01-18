package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"slices"
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
	Name  string `json:"name"`
	Rank  int    `json:"rank"`
	Score int    `json:"score"`
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
	os.WriteFile("./voters.json", data, 0644)
}

func saveResults() {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		panic(err)
	}
	os.WriteFile("./results1.json", data, 0644)
}

// declaring variables

var teachers = getTeachers()
var questions = getQuestions()
var results = getResults()
var voters = getVoters()

// helper functions
func writeAsJson(w http.ResponseWriter, stuff any) {
	data, err := json.Marshal(stuff)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(data))
}

func fillResults() {
	for questionName := range questions {
		if results[questionName] == nil {
			results[questionName] = map[string]map[string]int{}
		}
		for _, teacher := range teachers {
			if results[questionName][teacher] == nil {
				results[questionName][teacher] = map[string]int{"b": 0, "w": 0}
			}
		}
	}
}

// response functions

func Teachers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "[\""+strings.Join(teachers, "\", \"")+"\"]")
}

func Questions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAsJson(w, questions)
}

func TeacherRatingUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(400)
		return
	}

	name := r.PostFormValue("name")
	if !voters[strings.ToLower(name)] {
		w.WriteHeader(401)
		return
	}

	for questionName := range questions {
		best := r.PostFormValue(questionName + "-best")
		worst := r.PostFormValue(questionName + "-worst")
		println(questionName + "\tbest:" + best + "\tworst:" + worst)
	}
}

func Results(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeAsJson(w, results)
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
	writeAsJson(w, res)
}

func Categories(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	res := map[string][]RankedTeacher{}
	for category, categoryResult := range results {
		for teacher, scores := range categoryResult {
			res[category] = append(res[category], RankedTeacher{
				Name:  teacher,
				Rank:  -1,
				Score: scores["b"] - scores["w"],
			})
		}
		slices.SortFunc(res[category],
			func(a, b RankedTeacher) int {
				return b.Score - a.Score
			},
		)
		for i := 0; i < len(res[category]); i++ {
			if i > 0 && res[category][i].Score == res[category][i-1].Score {
				res[category][i].Rank = res[category][i-1].Rank
			} else {
				res[category][i].Rank = i + 1
			}
		}
	}
	writeAsJson(w, res)
}

func Category(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	res := []RankedTeacher{}

	category := ps.ByName("category")
	if results[category] == nil {
		w.WriteHeader(400)
		return
	}

	for teacher, scores := range results[category] {
		res = append(res, RankedTeacher{
			Name:  teacher,
			Rank:  -1,
			Score: scores["b"] - scores["w"],
		})
	}
	slices.SortFunc(res,
		func(a, b RankedTeacher) int {
			return b.Score - a.Score
		},
	)
	for i := 0; i < len(res); i++ {
		if i > 0 && res[i].Score == res[i-1].Score {
			res[i].Rank = res[i-1].Rank
		} else {
			res[i].Rank = i + 1
		}
	}
	writeAsJson(w, res)
}

func ValidateName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := ps.ByName("name")
	fmt.Fprint(w, voters[strings.ToLower(name)])
}

func NotImplemented(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "This page isn't implemented")
}

func MiddleCORS(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter,
		r *http.Request, ps httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next(w, r, ps)
	}
}

// main
func main() {
	fillResults()
	saveResults()

	router := httprouter.New()
	router.POST("/lehrer-ranking", MiddleCORS(TeacherRatingUpload))

	router.GET("/lehrer-ranking/lehrer", MiddleCORS(Teachers))
	router.GET("/lehrer-ranking/fragen", MiddleCORS(Questions))
	router.GET("/lehrer-ranking/ergebnisse", MiddleCORS(Results))
	router.GET("/lehrer-ranking/ergebnisse/l/:teacher", MiddleCORS(ResultsOfTeacher))
	router.GET("/lehrer-ranking/ergebnisse/k", MiddleCORS(Categories))
	router.GET("/lehrer-ranking/ergebnisse/k/:category", MiddleCORS(Category))
	router.GET("/lehrer-ranking/validate-name/:name", MiddleCORS(ValidateName))

	log.Fatal(http.ListenAndServe(":1337", router))
}
