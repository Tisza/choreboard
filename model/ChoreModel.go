package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/Workiva/go-datastructures/queue"
	"time"
	"container/list"
)

/**
TODO: more obfuscating authID reviersible hash
TODO: hold persistent data in .json files
TODO: write unit tests
TODO; write integration tests
*/


// ============================= Constants ============================== //


const INVALID_CHORE = ""
const INVALID_ASSIGNEE = ""
var OK = HttpStatus{200, "OK"}



// =================== Data Types ========================== //

// holds state information of a user
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

func (u *User) isAssignedTo(choreName string) bool {
	return u.AssignedChore == choreName
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

// holds state information of a chore
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

type SummonsQueue struct {
	queue *list.List
}

func initSummonsQueue(args... string) SummonsQueue {
	queue := list.New()
	for arg := range args {
		queue.PushBack(arg)
	}
	return SummonsQueue{queue}
}

// holds information to be returned to the client via HTTP response
type HttpStatus struct {
	Code        int
	Description string
}

func (e HttpStatus) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Description)
}


// Users and Chores don't need to be added to or removed in V1
// ========================== Internal Data Structures =========================== //


// a map of authIDs tp the state of known users
var Users = map[string](*User){"drew:bass": &User{"drew:bass", "drew", "bass", INVALID_CHORE, "", 0, false},
								"logan:huskies": &User{"logan:huskies", "logan", "huskies", INVALID_CHORE, "", 0, false},
								"katie:kittens": &User{"katie:kittens", "katie", "kittens", INVALID_CHORE, "", 0, false}} // key: authId

// a map of chore names to the state of known chores
var Chores = map[string](*Chore){"Dishes": &Chore{INVALID_ASSIGNEE, 3, false, "", "Put away clean dishes from the dishwasher and reload the dishwasher with dishes from the sink", "Dishes"},
									"Trash": &Chore{INVALID_ASSIGNEE, 2, false, "", "Move trash and recycling from the kitchen to their bins outside. If the recycling is full, use the dumpster next to Jak's", "Trash"},
									"Kitchen Counter": &Chore{INVALID_ASSIGNEE, 3, false, "", "Clear objects off of the kitchen counter. Dishes are put into the sink. Wipe down the counter.", "Kitchen Counter"},
									"Kitchen Floor": &Chore{INVALID_ASSIGNEE, 3, false, "", "Pick up items off the floor in the kitchen (including next to the trash!). Mop the floor.", "Kitchen Floor"}} // key: choreName

// a queue to hold chores that have been reported. Chores in this queue need to be done, but have not been assigned
var TodoChoreQ = queue.New(0)

// a queue to provide a summoning order. Users that have been summoned and accept chores will be placed at the back of the queue,
// whereas users that are summoned and decline chores will be placed at the front, so that they will be the next person notified to do work
var SummoningOrder = initSummonsQueue("drew:bass", "logan:huskies", "katie:kittens")

// channels used to synchronize access to other data structures between concurrently operating handlers
var UsersChan = make(chan map[string](*User), 1)
var ChoresChan = make(chan map[string](*Chore), 1)
var TodoChoreQChan = make(chan *queue.Queue, 1)
var SummoningOrderChan = make(chan SummonsQueue, 1)


// ======================== Public Functions =============== //

// returns JSON object with information about a particular user
func GetUserStatus(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) interface{} {
		user := users[authID]
		return user
	}, func(){}, authID)
}

// assigns the first chore in the queue of to-do chores to the summoned user with authID returns 200 OK
// otherwise returns 500 HttpStatus with error message
func AcceptChore(authID string, deadline string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) HttpStatus {

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

// user with authID is declining this chore, so their shame is increased and are marked as no longer summoned
func DeclineChore(authID string) HttpStatus {
	return OK
}

// returns a JSON object with information about all of the chores
func GetChoreBoard(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) interface{} {
		return choreList(chores)
	}, func(){}, authID)
}

// verifies that user with friendlyName and password exists, and returns a custom authID that will be used as an identifier for their session
func LoginUser(friendlyName string, password string) ([]byte, HttpStatus) {
	authID := constructAuthID(friendlyName, password)
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) interface{} {
		if passwordCheck(authID, password) {
				// everything checks out, return back the authID and OK status
				return authID
			} else {
				// Invalid password, report
				return []byte{}
			}
	}, func() {
		// adding users dynamically isn't needed for the demo
		//addUser(User{authID, friendlyName, password, "", "", 0, false})
	}, authID)
}

// reports that a chore needs work and summons the next user in the summons ordering, returns 200 OK. (Note that the user summoned at this time won't necessarily be assigned to this chore)
// otherwise returns 500 HttpStatus with error message
func ReportChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) HttpStatus {
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
	}, authID)
}

// a user with authID has finished choreName, so we unassign them and move them to the back of the summons ordering
func DoneWithChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) HttpStatus {
		if choreExists(choreName, chores) &&
			chores[choreName].isActive() &&
			!choreQContains(choreQ, choreName) {
			user := users[authID]
			if user.isAssignedTo(choreName) {
				// chore has been finished. the chore no longer NeedsWork, and the user assigned is unassigned
				chores[choreName].Assignee = INVALID_ASSIGNEE
				chores[choreName].NeedsWork = false
				chores[choreName].ReportedTime = ""
				user.unassign()
				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user '%s' is not assigned to this chore: %s", user.FriendlyName, user.AssignedChore)}
		} else {
			// In V1, a fixed number of chores exist, so we're not allowing clients to define their own chores on reporting

			return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist, is not assigned, or is in the queue of to-do chores", choreName)}
		}
	}, authID)
}

// ============================== Helpers ===================== //


// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterJson(getMarshalableObject func(map[string](*User), map[string](*Chore), *queue.Queue, SummonsQueue) interface{},
					optionalFailure func(),
					authID string) ([]byte, HttpStatus) {
	users, chores, choreQ, summoningOrder := aqcuireInternals()
	if verifyAuthID(authID, users) {
		bytes, status := marshalAndValidate(getMarshalableObject(users, chores, choreQ, summoningOrder))
		releaseInternals(users, chores, choreQ, summoningOrder)
		return bytes, status
	} else {
		releaseInternals(users, chores, choreQ, summoningOrder)
		optionalFailure()
		return []byte{}, HttpStatus{http.StatusForbidden, "Forbidden: Invalid authID"}
	}
}

// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterStatus(getStatus func(map[string](*User), map[string](*Chore), *queue.Queue, SummonsQueue) HttpStatus, authID string) HttpStatus {
	users, chores, choreQ, summoningOrder := aqcuireInternals()
	if verifyAuthID(authID, users) {
		status := getStatus(users, chores, choreQ, summoningOrder)
		releaseInternals(users, chores, choreQ, summoningOrder)
		return status
	} else {
		releaseInternals(users, chores, choreQ, summoningOrder)
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
	found := false
	for name := range choreNames {
		if name == choreName {
			found = true
		}
		choreQ.Put(name)
	}
	return found
}

func passwordCheck(authID string, password string) bool {
	return true
}

func addUser(u User) {

}

// gets internal data structures synchronously with other goroutine handlers
func aqcuireInternals() (map[string](*User), map[string](*Chore), *queue.Queue, SummonsQueue) {
	users := <-UsersChan
	chores := <-ChoresChan
	choreQ := <-TodoChoreQChan
	summoningOrder := <-SummoningOrderChan
	return users, chores, choreQ, summoningOrder
}

// releases internal data structures in the reverse order in which they were acquired
func releaseInternals(users map[string](*User), chores map[string](*Chore), choreQ *queue.Queue, summoningOrder SummonsQueue) {
	TodoChoreQChan <- choreQ
	ChoresChan <- chores
	UsersChan <- users
	SummoningOrderChan <- summoningOrder
}

func marshalAndValidate(v interface{}) ([]byte, HttpStatus) {
	if json, err := json.Marshal(v); err != nil {
		return []byte{}, HttpStatus{http.StatusInternalServerError, err.Error()}
	} else {
		return json, OK
	}
}
