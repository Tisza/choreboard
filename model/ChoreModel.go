package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/Workiva/go-datastructures/queue"
	"time"
)

/**
TODO: more obfuscating authID reviersible hash
TODO: hold persistent data in .json files
TODO: write unit tests
TODO; write integration tests
*/

// =================== Data Types ========================== //

type User struct {
	AuthID       string
	FriendlyName string
	Password string
	AssignedChore string        // assigned chore name, otherwise empty string
	Deadline string     // UTC local time for chore deadline if assigned, otherwise empty string
	Shame int
	Summoned bool
}

func (u *User) isAssigned() bool {
	return u.AssignedChore != INVALID_CHORE
}

func (u *User) unassign() {
	u.AssignedChore = INVALID_CHORE
	u.Deadline = ""
}

func (u *User) assignChore(choreName string, deadline string) {
	u.AssignedChore = choreName
	u.Deadline = deadline
	u.Summoned = false
}

type Chore struct {
	Assignee     string // the friendly name of the user assigned to this chore
	AmtOfShame   int
	NeedsWork       bool
	ReportedTime string // UTC local time for chore deadline
	Description  string // description of chore
	ChoreName    string
}

func (c *Chore) isActive() bool {
	return c.Assignee != INVALID_ASSIGNEE
}

func (c *Chore) assignUser(userName string, reportTime string) {
	c.Assignee = userName
	c.ReportedTime = reportTime
}

type HttpStatus struct {
	Code        int
	Description string
}

func (e HttpStatus) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Description)
}

var OK = HttpStatus{200, "OK"}

var Users = map[string](*User){"drew:bass": &User{"drew:bass", "drew", "bass", INVALID_CHORE, "", 0, false}} // key: authId

var Chores = map[string](*Chore){"Dishes": &Chore{INVALID_ASSIGNEE, 3, false, "", "Put away clean dishes from the dishwasher and reload the dishwasher with dishes from the sink", "Dishes"}} // key: choreName

var TodoChoreQ = queue.New(0)

var UsersChan = make(chan map[string](*User), 1)

var ChoresChan = make(chan map[string](*Chore), 1)

var TodoChoreQChan = make(chan *queue.Queue, 1)

const INVALID_CHORE = ""
const INVALID_ASSIGNEE = ""

// ======================== Public Functions =============== //

// returns JSON object with information about a particular user
func GetUserStatus(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) interface{} {
		user := users[authID]
		return user
	}, func(){}, authID)
}

func AcceptChore(authID string, deadline string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) HttpStatus {

		if choreQ.Empty() {
			return HttpStatus{500, fmt.Sprint("queue of to-do chores is empty. nothing to accept")}
		}

		tmp, _ := choreQ.Peek()
		choreName := tmp.(string)
		user := users[authID]

		if choreExists(choreName, chores) {
			if !user.isAssigned() {
				user.assignChore(choreName, deadline)
				chores[choreName].Assignee = user.FriendlyName
				choreQ.Get(1)  // throw it on the GROUND

				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user is already assigned to a chore: %s", users[authID].AssignedChore)}
		}
		return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist or is already active", choreName)}
	}, authID)
}

func DeclineChore(authID string) HttpStatus {
	return OK
}

func GetChoreBoard(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) interface{} {
		return choreList(chores)
	}, func(){}, authID)
}

func LoginUser(friendlyName string, password string) ([]byte, HttpStatus) {
	authID := constructAuthID(friendlyName, password)
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) interface{} {
		if passwordCheck(authID, password) {
				// everything checks out, return back the authID and OK status
				return authID
			} else {
				// Invalid password, report
				return []byte{}
			}
	}, func() {
		addUser(User{authID, friendlyName, password, "", "", 0, false})
	}, authID)
}

func ReportChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) HttpStatus {
		if choreExists(choreName, chores) &&
			!chores[choreName].isActive() &&
			!choreQContains(choreQ, choreName) {
			user := users[authID]
			if !user.isAssigned() {
				// chore exists and is inactive, so we'll put it into the queue of to-do chores, and update appropriate objects
				choreQ.Put(choreName)
				user.Summoned = true
				chores[choreName].NeedsWork = true
				chores[choreName].ReportedTime = time.Now().Local().Format(time.RFC1123)
				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user '%s' is already assigned to a chore: %s", user.FriendlyName, user.AssignedChore)}
		} else {
			// In V1, a fixed number of chores exist, so we're not allowing clients to define their own chores on reporting

			return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist, is already assigned, or is already in the queue of to-do chores", choreName)}
		}
		return OK
	}, authID)
}

func DoneWithChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) HttpStatus {
		if choreExists(choreName, chores) &&
			chores[choreName].isActive() &&
			!choreQContains(choreQ, choreName) {
			user := users[authID]
			if user.isAssigned() {
				// chore has been finished. the chore no longer NeedsWork, and the user assigned is unassigned
				chores[choreName].Assignee = INVALID_ASSIGNEE
				chores[choreName].NeedsWork = false
				chores[choreName].ReportedTime = ""
				user.unassign()
				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user '%s' is not assigned to a chore cannot be done with this one: %s", user.FriendlyName, user.AssignedChore)}
		} else {
			// In V1, a fixed number of chores exist, so we're not allowing clients to define their own chores on reporting

			return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist, is not assigned, or is in the queue of to-do chores", choreName)}
		}
		return OK
	}, authID)
}

// ============================== Helpers ===================== //


// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterJson(getMarshalableObject func(map[string](*User), map[string](*Chore), *queue.Queue) interface{},
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
func authFilterStatus(getStatus func(map[string](*User), map[string](*Chore), *queue.Queue) HttpStatus, authID string) HttpStatus {
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

func constructAuthID(friendlyName string, password string) string {
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

func choreList(chores map[string](*Chore)) []Chore {
	cl := make([]Chore, 0)
	for choreName := range chores {
		cl = append(cl, *chores[choreName])
	}
	return cl
}

func choreQContains(choreQ *queue.Queue, choreName interface{}) bool {
	choreNames, _ := choreQ.Get(choreQ.Len())
	for name := range choreNames {
		if name == choreName {
			return true
		}
	}
	return false
}

func passwordCheck(authID string, password string) bool {
	return true
}

func addUser(u User) {

}

// gets internal data structures synchronously with other goroutine handlers
func aqcuireInternals() (map[string](*User), map[string](*Chore), *queue.Queue) {
	users := <-UsersChan
	chores := <-ChoresChan
	choreQ := <-TodoChoreQChan
	return users, chores, choreQ
}

// releases internal data structures in the reverse order in which they were acquired
func releaseInternals(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue) {
	TodoChoreQChan <- choreQ
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
