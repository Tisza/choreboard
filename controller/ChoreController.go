package main

import (
	"choreboard/model"
	"fmt"
	"net/http"
	"strconv"
	"github.com/emirpasic/gods/sets/hashset"
)

const PORT string = "8080"
const HOST = ":" + PORT

var USER_STATUS_PARAMS = []string{"authID"}
var ACCEPT_CHORE_PARAMS = []string{"authID", "deadline"}
var DECLINE_CHORE_PARAMS = []string{"authID"}
var CHORE_BOARD_PARAMS = []string{"authID"}
var SCORE_BOARD_PARAMS = []string{"authID"}
var LOGIN_USER_PARAMS = []string{"friendlyName", "password"}
var REPORT_CHORE_PARAMS = []string{"authID", "choreName"}
var DONE_WITH_CHORE_PARAMS = []string{"authID", "choreName"}

func main() {

	model.InititalizeDataStructures()

	http.HandleFunc("/userStatus", badRequestFilter(handleUserStatus, USER_STATUS_PARAMS))
	http.HandleFunc("/acceptChore", badRequestFilter(handleAcceptChore, ACCEPT_CHORE_PARAMS))
	http.HandleFunc("/declineChore", badRequestFilter(handleDeclineChore, DECLINE_CHORE_PARAMS))
	http.HandleFunc("/choreBoard", badRequestFilter(handleChoreBoard, CHORE_BOARD_PARAMS))
	http.HandleFunc("/scoreBoard", badRequestFilter(handleScoreBoard, SCORE_BOARD_PARAMS))
	http.HandleFunc("/loginUser", badRequestFilter(handleLoginUser, LOGIN_USER_PARAMS))
	http.HandleFunc("/reportChore", badRequestFilter(handleReportChore, REPORT_CHORE_PARAMS))
	http.HandleFunc("/doneWithChore", badRequestFilter(handleDoneWithChore, DONE_WITH_CHORE_PARAMS))

	fmt.Println("About to ListenAndServe on " + HOST)
	err := http.ListenAndServe(HOST, nil)
	fmt.Printf("%v", err)
}

//=============================== Handlers ===============================//

func handleUserStatus(w http.ResponseWriter, r *http.Request) {
	json, status := model.GetUserStatus(r.Form[USER_STATUS_PARAMS[0]][0])
	handleJson(w, json, status)
}

func handleAcceptChore(w http.ResponseWriter, r *http.Request) {
	json, status := model.AcceptChore(r.Form[ACCEPT_CHORE_PARAMS[0]][0], r.Form[ACCEPT_CHORE_PARAMS[1]][0])
	handleJson(w, json, status)
}

func handleDeclineChore(w http.ResponseWriter, r *http.Request) {
	status := model.DeclineChore(r.Form[DECLINE_CHORE_PARAMS[0]][0])
	handleStatus(w, r, status)
}

func handleChoreBoard(w http.ResponseWriter, r *http.Request) {
	json, status := model.GetChoreBoard(r.Form[CHORE_BOARD_PARAMS[0]][0])
	handleJson(w, json, status)
}

func handleScoreBoard(w http.ResponseWriter, r *http.Request) {
	json, status := model.GetScoreBoard(r.Form[SCORE_BOARD_PARAMS[0]][0])
	handleJson(w, json, status)
}

func handleLoginUser(w http.ResponseWriter, r *http.Request) {
	json, status := model.LoginUser(r.Form[LOGIN_USER_PARAMS[0]][0], r.Form[LOGIN_USER_PARAMS[1]][0])
	handleJson(w, json, status)
}

func handleReportChore(w http.ResponseWriter, r *http.Request) {
	status := model.ReportChore(r.Form[REPORT_CHORE_PARAMS[0]][0], r.Form[REPORT_CHORE_PARAMS[1]][0])
	handleStatus(w, r, status)
}

func handleDoneWithChore(w http.ResponseWriter, r *http.Request) {
	status := model.DoneWithChore(r.Form[DONE_WITH_CHORE_PARAMS[0]][0], r.Form[DONE_WITH_CHORE_PARAMS[1]][0])
	handleStatus(w, r, status)
}

//=============================== Helpers ===========================//

func badRequestFilter(next http.HandlerFunc, expectedParams []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		if !isBadRequest(w, r, expectedParams) {
			next.ServeHTTP(w, r)
		}
	})
}

func handleJson(w http.ResponseWriter, json []byte, status model.HttpStatus) {
	if status.Code != 200 {
		http.Error(w, strconv.Itoa(status.Code) + ": " + status.Description, status.Code)
	} else {
		fmt.Fprintf(w, "%s", json)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request, status model.HttpStatus) {
	if status.Code != 200 {
		http.Error(w, strconv.Itoa(status.Code) + ": " + status.Description, status.Code)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func printRequestTraits(r *http.Request) {
	fmt.Printf("URL: %v\n", r.URL)
	fmt.Printf("Body: %v\n", r.Body)
	fmt.Printf("Header: %v\n", r.Header)
	fmt.Printf("Method: %v\n", r.Method)
	fmt.Printf("Remote Address: %v\n", r.RemoteAddr)
	r.ParseForm()
	fmt.Println("Form: ", r.Form)
	fmt.Println()
}

func getParams(r *http.Request) *hashset.Set {
	params := hashset.New()
	for param := range r.Form {
		//fmt.Println(param)
		params.Add(param)
	}
	return params
}

// returns true if request has valid parameters
// otherwise sends back a 400 Bad Request response and returns false
// NOTE: r.Form is ready to be examined after running this function
func isBadRequest(w http.ResponseWriter, r *http.Request, expectedParams []string) bool {
	r.ParseForm()
	requestParams := getParams(r)

	if !eq(requestParams, expectedParams) {
		// uh oh, we got a 400 Bad Request over 'ere
		w.WriteHeader(http.StatusBadRequest)
		return true
	} else {
		return false
	}
}

func eq(a *hashset.Set, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}
	if int(a.Size()) != len(b) {
		return false
	}
	for _, val := range b {
		if !a.Contains(val) {
			return false
		}
	}
	return true
}