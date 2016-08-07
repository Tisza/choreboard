package model

import (
	"encoding/json"
	"net/http"
	"strings"
	"fmt"
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
	Chore string        // assigned chore name, otherwise empty string
	Deadline string     // UTC local time for chore deadline if assigned, otherwise empty string
	Shame int
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

// ====================== End of Data Types ================ //

// ======================== Public Functions =============== //

// returns JSON object with information about a particular user
func GetUserStatus(authID string) ([]byte, HttpStatus) {
	if verifyAuthID(authID) {
		person := User{authID, "Bob", "chickens", "", "", 0}
		return marshalAndValidate(person)
	} else {
		return []byte{}, HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

func SetUserChore(authID string, choreName string, accept bool) HttpStatus {
	return OK
}

func GetChoreBoard(authID string) ([]byte, HttpStatus) {
	if verifyAuthID(authID) {
		chore1 := Chore{"Bob", 9001, true, "2016-08-03T14:00:00Z", "Take out the trash", "1"}
		chore2 := Chore{"Logan", 2, true, "2016-07-03T14:00:00Z", "Be pretty", "2"}
		chore3 := Chore{"", 500, false, "2016-08-01T14:00:00Z", "Clean the sink", "3"}
		chores := []Chore{chore1, chore2, chore3}
		return marshalAndValidate(chores)
	} else {
		return []byte{}, HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

func LoginUser(friendlyName string, password string) ([]byte, HttpStatus){
	authID := constructAuthID(friendlyName, password)
	if verifyAuthID(authID) {
		// User is already known
		// check password
		if passwordCheck(authID, password) {
			// everything checks out, return back the authID and OK status
			return marshalAndValidate(authID)
		} else {
			// Invalid password, report
			return []byte{}, HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
		}
	} else {
		// New user
		addUser(User{authID, friendlyName, password, "", "", 0})
		return marshalAndValidate(authID)
	}
}

func ReportChore(authID string, choreName string, mode string) HttpStatus {
	if verifyAuthID(authID) {

		return OK

	} else {
		return HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

// ========================= End of Public Functions ========== //

// ============================== Helpers ===================== //

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

func verifyAuthID(authID string) bool {
	return true
}

func passwordCheck(authID string, password string) bool {
	return true
}

func addUser(u User) {

}

func marshalAndValidate(v interface{}) ([]byte, HttpStatus) {
	if json, err := json.Marshal(v); err != nil {
		return []byte{}, HttpStatus{http.StatusInternalServerError, err.Error()}
	} else {
		return json, OK
	}
}

// ======================== End of Helpers ================== //