package model

import (
	"encoding/json"
	"net/http"
	"strings"
	"fmt"
	"container/list"
)

/**
	TODO: more obfuscating authID reviersible hash
	TODO: hold persistent data in .json files
 */

// =================== Data Types ========================== //

type User struct {
	AuthID string
	FriendlyName string
	Password string
	AssignedChore string        // assigned chore name, otherwise empty string
	Deadline string     // UTC local time for chore deadline if assigned, otherwise empty string
	Shame int
}

func (u *User) isAssigned() bool {
	return u.AssignedChore != INVALID_CHORE
}

func (u *User) unassign() *User {
	u.AssignedChore = INVALID_CHORE
	return u
}

func (u *User) assignChore(choreName string, deadline string) *User {
	u.AssignedChore = choreName
	u.Deadline = deadline
	return u
}

type Chore struct {
	Assignee string     // the friendly name of the user assigned to this chore
	AmtOfShame int
	Active bool
	ReportedTime string // UTC local time for chore deadline
	Description string  // description of chore
	ChoreName string
}

type HttpStatus struct {
	Code        int
	Description string
}

func (e HttpStatus) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Description)
}

var OK = HttpStatus{200, "OK"}

var Users = map[string](*User){"drew:bass": &User{"drew:bass", "drew", "bass", INVALID_CHORE, "", 0}} // key: authId

var Chores = map[string](*Chore){"Dishes": &Chore{INVALID_ASSIGNEE, 3, true, "2016-08-09T02:00:00Z", "Put away clean dishes from the dishwasher and reload the dishwasher with dishes from the sink", "Dishes"}} // key: choreName

var ChoreQ = list.New()

var UsersChan = make(chan map[string](*User), 1)

var ChoresChan = make(chan map[string](*Chore), 1)

var ChoreQChan = make(chan *list.List, 1)

const INVALID_CHORE = ""
const INVALID_ASSIGNEE = ""

// ======================== Public Functions =============== //

// returns JSON object with information about a particular user
func GetUserStatus(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List) interface{} {
		user := users[authID]
		return user
	}, func(){}, authID)
}

func AcceptChore(authID string, choreName string, deadline string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List) HttpStatus {
		if choreExists(choreName, chores) {
			if !users[authID].isAssigned() {
				users[authID].assignChore(choreName, deadline)
				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user is already assigned to a chore: %s", users[authID].AssignedChore)}
		}
		return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist", choreName)}
	}, authID)
}

func GetChoreBoard(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List) interface{} {
		chore1 := Chore{"Bob", 9001, true, "2016-08-03T14:00:00Z", "Take out the trash", "1"}
		chore2 := Chore{"Logan", 2, true, "2016-07-03T14:00:00Z", "Be pretty", "2"}
		chore3 := Chore{"", 500, false, "2016-08-01T14:00:00Z", "Clean the sink", "3"}
		return []Chore{chore1, chore2, chore3}
	}, func(){}, authID)
}

func LoginUser(friendlyName string, password string) ([]byte, HttpStatus){
	authID := constructAuthID(friendlyName, password)
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List) interface{} {
		if passwordCheck(authID, password) {
				// everything checks out, return back the authID and OK status
				return authID
			} else {
				// Invalid password, report
				return []byte{}
			}
	}, func() {
		addUser(User{authID, friendlyName, password, "", "", 0})
	}, authID)
}

func ReportChore(authID string, choreName string, mode string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List) HttpStatus {
		return OK
	}, authID)
}

// ============================== Helpers ===================== //

// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterJson(getMarshalableObject func(map[string](*User), map[string](*Chore), *list.List) interface{},
					optionalFailure func(),
					authID string) ([]byte, HttpStatus) {
	users, chores, choreQ := aqcuireInternals()
	if verifyAuthID(authID, users) {
		bytes, status := marshalAndValidate(getMarshalableObject(users, chores, choreQ))
		releaseInternals(users, chores, choreQ)
		return bytes, status
	} else {
		releaseInternals(users, chores, choreQ)
		optionalFailure()
		return []byte{}, HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterStatus(getStatus func(map[string](*User), map[string](*Chore), *list.List) HttpStatus, authID string) HttpStatus {
	users, chores, choreQ := aqcuireInternals()
	if verifyAuthID(authID, users) {
		status := getStatus(users, chores, choreQ)
		releaseInternals(users, chores, choreQ)
		return status
	} else {
		releaseInternals(users, chores, choreQ)
		return HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

func setNextUser(authID string, friendlyName string) HttpStatus {
	return OK
}

func constructAuthID(friendlyName string, password string) (string) {
	return friendlyName + ":" + password
}

func deconstructAuthID(authID string) (string, string) {
	split := strings.Split(authID, ":")
	return split[0], split[1]
}

func verifyAuthID(authID string, users map[string](*User)) bool {
	_, ok := users[authID]
	return ok
}

func choreExists(choreName string, chores map[string](*Chore)) bool {
	_, ok := chores[choreName]
	return ok
}

func passwordCheck(authID string, password string) bool {
	return true
}

func addUser(u User) {

}

// gets internal data structures synchronously with other goroutine handlers
func aqcuireInternals() (map[string](*User), map[string](*Chore), *list.List) {
	users := <-UsersChan
	chores := <-ChoresChan
	choreQ := <-ChoreQChan
	return users, chores, choreQ
}

// releases internal data structures in the reverse order in which they were acquired
func releaseInternals(users map[string](*User), chores map[string](*Chore), choreQ *list.List) {
	ChoreQChan <- choreQ
	ChoresChan <- chores
	UsersChan <- users
}

func marshalAndValidate(v interface{}) ([]byte, HttpStatus) {
	if json, err := json.Marshal(v); err != nil {
		return []byte{}, HttpStatus{http.StatusInternalServerError, err.Error()}
	} else {
		return json, OK
	}
}