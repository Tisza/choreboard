package main

import (
	"fmt"
	"net/http"
	"choreboard/model"
	"github.com/golang-collections/go-datastructures/set"
	"strconv"
)

/**
	TODO: implement handleUserStatus
	TODO: implement handleSignChore
	TODO: implement handleChoreBoard
	TODO: implement handleLoginUser
	TODO: implement handleReportChore
 */

//var HOST_NAME , err = externalIP()
var PORT string = "8080"
var HOST = ":" + PORT

var USER_STATUS_PARAMS = []string{"authID"}
var ACCEPT_CHORE_PARAMS = []string{"authID", "choreName", "deadline"}
var CHORE_BOARD_PARAMS = []string{"authID"}
var LOGIN_USER_PARAMS = []string{"friendlyName", "password"}
var REPORT_CHORE_PARAMS = []string{"authID", "choreName", "mode"}


func main() {

	// TODO: figure out a way to refactor channel initialization to model
	model.UsersChan <- model.Users
	model.ChoresChan <- model.Chores
	model.ChoreQChan <- model.ChoreQ

	http.HandleFunc("/userStatus", badRequestFilter(handleUserStatus, USER_STATUS_PARAMS))
	http.HandleFunc("/acceptChore", badRequestFilter(handleAcceptChore, ACCEPT_CHORE_PARAMS))
	http.HandleFunc("/choreBoard", badRequestFilter(handleChoreBoard, CHORE_BOARD_PARAMS))
	http.HandleFunc("/loginUser", badRequestFilter(handleLoginUser, LOGIN_USER_PARAMS))
	http.HandleFunc("/reportChore", badRequestFilter(handleReportChore, REPORT_CHORE_PARAMS))

	fmt.Println("About to ListenAndServe on " + HOST)
	http.ListenAndServe(HOST, nil)
}

//=============================== Handlers ===============================//


func handleUserStatus(w http.ResponseWriter, r *http.Request) {
	json, status := model.GetUserStatus(r.Form[USER_STATUS_PARAMS[0]][0])
	handleJson(w, json, status)
}

func handleAcceptChore(w http.ResponseWriter, r *http.Request) {
	status := model.AcceptChore(r.Form[ACCEPT_CHORE_PARAMS[0]][0], r.Form[ACCEPT_CHORE_PARAMS[1]][0], r.Form[ACCEPT_CHORE_PARAMS[2]][0])
	handleStatus(w, r, status)
}

func handleChoreBoard(w http.ResponseWriter, r *http.Request) {
	json, status := model.GetChoreBoard(r.Form[CHORE_BOARD_PARAMS[0]][0])
	handleJson(w, json, status)
}

func handleLoginUser(w http.ResponseWriter, r *http.Request) {
	json, status := model.LoginUser(r.Form[LOGIN_USER_PARAMS[0]][0], r.Form[LOGIN_USER_PARAMS[1]][0])
	handleJson(w, json, status)
}

func handleReportChore(w http.ResponseWriter, r *http.Request) {
	status := model.ReportChore(r.Form[REPORT_CHORE_PARAMS[0]][0], r.Form[REPORT_CHORE_PARAMS[1]][0], r.Form[REPORT_CHORE_PARAMS[2]][0])
	handleStatus(w, r, status)
}


//=============================== Helpers ===========================//

func badRequestFilter(next http.HandlerFunc, expectedParams []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBadRequest(w, r, expectedParams) {
			w.Header().Add("Access-Control-Allow-Origin", "*")
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

func getParams(r *http.Request) *set.Set {
	params := set.New()
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
	//fmt.Println("Form: ", r.Form)
	//fmt.Println("Expected Parameters: ", expectedParams)
	requestParams := getParams(r)
	//fmt.Println("Request Parameters: ", requestParams)
	//fmt.Println()

	if !eq(requestParams, expectedParams) {
		// uh oh, we got a 400 Bad Request over 'ere
		//println("bad request")
		//fmt.Printf("%s\n", requestParams)
		//fmt.Printf("%s\n", expectedParams)
		w.WriteHeader(http.StatusBadRequest)
		return true
	} else {
		return false
	}
}

func eq(a *set.Set, b []string) bool {
	if a == nil && b == nil {
		return true;
	}

	if a == nil || b == nil {
		return false;
	}

	if int(a.Len()) != len(b) {
		return false
	}
	for _, val := range b {
		if !a.Exists(val) {
			return false
		}
	}
	return true
}