package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"container/list"
	"errors"
	"os"
	"io/ioutil"
	"hash/fnv"
)

/**
TODO: write unit tests
TODO; write integration tests
*/


// ============================= Constants ============================== //


const INVALID_CHORE = ""
const INVALID_ASSIGNEE = ""
var OK = HttpStatus{200, "OK"}

const USERS_FILENAME = "users.json"
const CHORES_FILENAME = "chores.json"
const CHOREQ_FILENAME = "todoChoreQ.json"
const SUMMONING_ORDER_FILENAME = "summoningOrder.json"



// =================== Data Types ========================== //

// holds state information of a user
type User struct {
	FriendlyName string
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

func (u *User) acceptChore(choreName string, deadline string) {
	u.AssignedChore = choreName
	u.Deadline = deadline
	u.Summoned = false
}

func (u *User) declineChore(choreShame int) {
	u.Shame += choreShame
	u.Summoned = false
}

// holds state information of a chore
type Chore struct {
	Assignee     string // the friendly name of the user assigned to this chore
	AmtOfShame   int
	NeedsWork       bool
	ReportedTime string // UTC local time for chore deadline
	Description  string // description of chore
}

func (c *Chore) isActive() bool {
	return c.Assignee != INVALID_ASSIGNEE
}

func (c *Chore) assignUser(userName string, reportTime string) {
	c.Assignee = userName
	c.ReportedTime = reportTime
}

func initSummonsOrder(args ...string) *list.List {
	queue := list.New()
	for _, arg := range args {
		queue.PushBack(arg)
	}
	return queue
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
var Users = map[string](*User){"201849424": &User{"drew", INVALID_CHORE, "", 0, false},
								"2897112466": &User{"logan", INVALID_CHORE, "", 0, false},
								"2825122131": &User{"katie", INVALID_CHORE, "", 0, false},
								"3292495916": &User{"steve", INVALID_CHORE, "", 0, false},
								"657641543": &User{"tim", INVALID_CHORE, "", 0, false},
								"1044239700": &User{"alan", INVALID_CHORE, "", 0, false},
								} // key: authId

// a map of chore names to the state of known chores
var Chores = map[string](*Chore){"Dishes": &Chore{INVALID_ASSIGNEE, 3, false, "", "Put away clean dishes from the dishwasher and reload the dishwasher with dishes from the sink"},
									"Trash": &Chore{INVALID_ASSIGNEE, 2, false, "", "Move trash/recycling/compost from the kitchen to their bins outside. If the recycling is full, use the dumpster next to Jak's"},
									"Kitchen Counter": &Chore{INVALID_ASSIGNEE, 3, false, "", "Clear objects off of the kitchen counter. Dishes are put into the sink. Wipe down the counter."},
									"Kitchen Floor": &Chore{INVALID_ASSIGNEE, 3, false, "", "Pick up items off the floor in the kitchen (including next to the trash!). Sweep and Mop the floor."},
									"Dining Room Pickup": &Chore{INVALID_ASSIGNEE, 3, false, "", "Coordinate with mates to clear out personal items from the common space. Throw away large trash."},
									"Dining Room Floor": &Chore{INVALID_ASSIGNEE, 3, false, "", "Sweep and mop the dining room floor"},
									"Whiteboard": &Chore{INVALID_ASSIGNEE, 3, false, "", "Clean the whiteboard with 'whiteboard cleaner'. If 'Save' is written, take a picture of the whiteboard before cleaning"},
									"Stairs": &Chore{INVALID_ASSIGNEE, 3, false, "", "Sweep the stairs"},
									"Upstairs Bathroom Fixtures": &Chore{INVALID_ASSIGNEE, 3, false, "", "Clean toilet, sink, tub"},
									"Downstairs Bathroom Fixtures": &Chore{INVALID_ASSIGNEE, 3, false, "", "Clean toilet, sink, tub"},
									"Upstairs Bathroom Floor/Trash": &Chore{INVALID_ASSIGNEE, 3, false, "", "Take out trash and clean the floor of the bathroom"},
									"Downatairs Bathroom Floor/Trash": &Chore{INVALID_ASSIGNEE, 3, false, "", "Take out trash and clean the floor of the bathroom"},
									"Living Room Pickup": &Chore{INVALID_ASSIGNEE, 3, false, "", "Coordinate with mates to clear out personal items from the common space.  Throw away large trash."},
									"Living Room Floor": &Chore{INVALID_ASSIGNEE, 3, false, "", "Sweep/vacuum the living room floor"},
									"Upstairs Hallway": &Chore{INVALID_ASSIGNEE, 3, false, "", "Sweep and mop the upstairs hallway"},
									"Downstairs Floor": &Chore{INVALID_ASSIGNEE, 3, false, "", "Sweep/vacuum the downstairs common floor"},
									"Downstairs Recycling": &Chore{INVALID_ASSIGNEE, 3, false, "", "Take the recycling pile downstairs directly outside"},
								} // key: choreName

// a queue to hold chores that have been reported. Chores in this queue need to be done, but have not been assigned
var TodoChoreQ = list.New()

// a queue to provide a summoning order. Users that have been summoned and accept chores will be placed at the back of the queue,
// whereas users that are summoned and decline chores will be placed at the front, so that they will be the next person notified to do work
var SummoningOrder = initSummonsOrder("201849424", "2897112466", "2825122131", "3292495916", "657641543", "1044239700")

// channels used to synchronize access to other data structures between concurrently operating handlers
var UsersChan = make(chan map[string](*User), 1)
var ChoresChan = make(chan map[string](*Chore), 1)
var TodoChoreQChan = make(chan *list.List, 1)
var SummoningOrderChan = make(chan *list.List, 1)


// ======================== Public Functions =============== //

// returns JSON object with information about a particular user
func GetUserStatus(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) interface{} {
		return users[authID]
	}, func(){}, authID)
}

// assigns the first chore in the queue of to-do chores to the summoned user with authID, and returns 200 OK
// otherwise returns 500 HttpStatus with error message
func AcceptChore(authID string, deadline string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) HttpStatus {

		if choreQ.Len() == 0 {
			return HttpStatus{500, fmt.Sprint("queue of to-do chores is empty. nothing to accept")}
		}

		tmp := choreQ.Front()
		choreName := tmp.Value.(string)
		user := users[authID]

		if choreExists(choreName, chores) {
			if !user.isAssigned() {
				user.acceptChore(choreName, deadline)
				chores[choreName].Assignee = user.FriendlyName
				choreQ.Remove(tmp)
				
				// mutation occurred, write changes to disk
				writeToJsonStorageFile(USERS_FILENAME, users)
				writeToJsonStorageFile(CHORES_FILENAME, chores)
				writeToJsonStorageFile(CHOREQ_FILENAME, listToSlice(choreQ))

				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user is already assigned to a chore: %s", users[authID].AssignedChore)}
		}
		return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist or is already active", choreName)}
	}, authID)
}

// user with authID is declining this chore, so their shame is increased and are marked as no longer summoned
func DeclineChore(authID string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) HttpStatus {

		if choreQ.Len() == 0 {
			return HttpStatus{500, fmt.Sprint("queue of to-do chores is empty. nothing to accept")}
		}

		tmp := choreQ.Front()
		choreName := tmp.Value.(string)
		user := users[authID]

		if choreExists(choreName, chores) {
			if !user.isAssigned() && user.Summoned {
				nextUserAuthID, err := nextUserToSummon(summoningOrder)
				if err == nil {
					nextUser := users[nextUserAuthID]
					user.declineChore(chores[choreName].AmtOfShame)

					nextUser.Summoned = true
					removeUserFromSummoningOrder(summoningOrder, nextUserAuthID)

					summoningOrder.PushFront(authID)

					// mutation occurred, write changes to disk
					writeToJsonStorageFile(USERS_FILENAME, users)
					writeToJsonStorageFile(SUMMONING_ORDER_FILENAME, listToSlice(summoningOrder))

					return OK
				}
				return HttpStatus{500, fmt.Sprintf("SummonOrderError: %s", err.Error())}
			}
			return HttpStatus{500, fmt.Sprintf("user is already assigned to a chore or has not been summoned\nAssigned Chore: %s\nSummoned: %v", user.AssignedChore, user.Summoned)}
		}
		return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist or is already active", choreName)}
	}, authID)
}

// returns a JSON object with information about all of the chores
func GetChoreBoard(authID string) ([]byte, HttpStatus) {
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) interface{} {
		return chores
	}, func(){}, authID)
}

// verifies that user with friendlyName and password exists, and returns a custom authID that will be used as an identifier for their session
func LoginUser(friendlyName string, password string) ([]byte, HttpStatus) {
	authID := constructAuthID(friendlyName, password)
	return authFilterJson(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) interface{} {
		// everything checks out, return back the authID and OK status
		return authID
	}, func() {
		// adding users dynamically isn't needed for the demo
		//addUser(User{authID, friendlyName, password, "", "", 0, false})
	}, authID)
}

// reports that a chore needs work and summons the next user in the summons ordering, returns 200 OK. (Note that the user summoned at this time won't necessarily be assigned to this chore)
// otherwise returns 500 HttpStatus with error message
func ReportChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) HttpStatus {
		if choreExists(choreName, chores) &&
			!chores[choreName].isActive() &&
			!choreQContains(choreQ, choreName) {
			nextUserAuthID, err := nextUserToSummon(summoningOrder)
			if err == nil {
				nextUser := users[nextUserAuthID]
				// chore exists and is inactive, so we'll put it into the queue of to-do chores, and update appropriate objects
				choreQ.PushBack(choreName)
				nextUser.Summoned = true
				removeUserFromSummoningOrder(summoningOrder, nextUserAuthID)
				chores[choreName].NeedsWork = true
				chores[choreName].ReportedTime = time.Now().Local().Format(time.RFC1123)

				// mutation occurred, write changes to disk
				writeToJsonStorageFile(USERS_FILENAME, users)
				writeToJsonStorageFile(CHORES_FILENAME, chores)
				writeToJsonStorageFile(CHOREQ_FILENAME, listToSlice(choreQ))
				writeToJsonStorageFile(SUMMONING_ORDER_FILENAME, listToSlice(summoningOrder))
				
				return OK
			}
			return HttpStatus{500, fmt.Sprintf("SummonOrderError: %s", err.Error())}
		} else {
			// In V1, a fixed number of chores exist, so we're not allowing clients to define their own chores on reporting

			return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist, is already assigned, or is already in the queue of to-do chores", choreName)}
		}
	}, authID)
}

// a user with authID has finished choreName, so we unassign them and move them to the back of the summons ordering
func DoneWithChore(authID string, choreName string) HttpStatus {
	return authFilterStatus(func(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) HttpStatus {
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
				summoningOrder.PushBack(authID)

				// mutation occurred, write changes to disk
				writeToJsonStorageFile(USERS_FILENAME, users)
				writeToJsonStorageFile(CHORES_FILENAME, chores)
				writeToJsonStorageFile(SUMMONING_ORDER_FILENAME, listToSlice(summoningOrder))

				return OK
			}
			return HttpStatus{500, fmt.Sprintf("user '%s' is not assigned to this chore: %s", user.FriendlyName, user.AssignedChore)}
		} else {
			// In V1, a fixed number of chores exist, so we're not allowing clients to define their own chores on reporting

			return HttpStatus{500, fmt.Sprintf("chore '%s' does not exist, is not assigned, or is in the queue of to-do chores", choreName)}
		}
	}, authID)
}

// sets up storage files if none exist, read in data structures, and assign them to their appropriate channels
func InititalizeDataStructures() {

	// initialize data structures from disk
	var users map[string](*User)
	var chores map[string](*Chore)
	var todoChoreQ *list.List
	var summoningOrder *list.List

	var tmpChoreQ []string
	var tmpOrder []string


	// ensure files for data structures exist, and create them if they don't
	if _, err := os.Stat(USERS_FILENAME); os.IsNotExist(err) {
		initializeStorageFile(USERS_FILENAME, Users)
	}
	if _, err := os.Stat(CHORES_FILENAME); os.IsNotExist(err) {
		initializeStorageFile(CHORES_FILENAME, Chores)
	}
	if _, err := os.Stat(CHOREQ_FILENAME); os.IsNotExist(err) {
		initializeStorageFile(CHOREQ_FILENAME, listToSlice(TodoChoreQ))
	}
	if _, err := os.Stat(SUMMONING_ORDER_FILENAME); os.IsNotExist(err) {
		initializeStorageFile(SUMMONING_ORDER_FILENAME, listToSlice(SummoningOrder))
	}

	// read in structures from storage
	user_bytes, _ := ioutil.ReadFile(USERS_FILENAME)
	json.Unmarshal(user_bytes, &users)

	chores_bytes, _ := ioutil.ReadFile(CHORES_FILENAME)
	json.Unmarshal(chores_bytes, &chores)

	todoChoreQ_bytes, _ := ioutil.ReadFile(CHOREQ_FILENAME)
	json.Unmarshal(todoChoreQ_bytes, &tmpChoreQ)

	summoningOrder_bytes, _ := ioutil.ReadFile(SUMMONING_ORDER_FILENAME)
	json.Unmarshal(summoningOrder_bytes, &tmpOrder)

	todoChoreQ = sliceToList(tmpChoreQ)
	summoningOrder = sliceToList(tmpOrder)

	// assign data structures to appropriate channels
	UsersChan <- users
	ChoresChan <- chores
	TodoChoreQChan <- todoChoreQ
	SummoningOrderChan <- summoningOrder
}

// ============================== Helpers ===================== //


// MMMMMMMMMM
// synchronously acquires and releases internal data structures outside of given function
func authFilterJson(getMarshalableObject func(map[string](*User), map[string](*Chore), *list.List, *list.List) interface{},
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
func authFilterStatus(getStatus func(map[string](*User), map[string](*Chore), *list.List, *list.List) HttpStatus, authID string) HttpStatus {
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

func nextUserToSummon(summoningOrder *list.List) (string, error) {
	if summoningOrder.Len() == 0 {
		return "", errors.New("All users are summoned or assigned to a task!")
	} else {
		authID := summoningOrder.Front().Value.(string)
		return authID, nil
	}
}

func removeUserFromSummoningOrder(summoningOrder *list.List, authID string) {
	for elem := summoningOrder.Front(); true; elem = elem.Next() {
		if elem.Value == authID {
			summoningOrder.Remove(elem)
			return
		}
	}
}

func constructAuthID(friendlyName string, password string) string {
	return fmt.Sprint(hash(friendlyName + password))
}

func verifyAuthID(authID string, users map[string](*User)) bool {
	_, ok := users[authID]
	return ok
}

func choreExists(choreName string, chores map[string](*Chore)) bool {
	_, ok := chores[choreName]
	return ok
}

//func choreList(chores map[string](*Chore)) []Chore {
//	cl := make([]Chore, 0)
//	for choreName := range chores {
//		cl = append(cl, *chores[choreName])
//	}
//	return cl
//}

func choreQContains(choreQ *list.List, choreName interface{}) bool {
	for currChoreName := choreQ.Front(); currChoreName != nil; currChoreName = currChoreName.Next() {
		if (currChoreName.Value == choreName) {
			return true
		}
	}
	return false
}

// gets internal data structures synchronously with other goroutine handlers
func aqcuireInternals() (map[string](*User), map[string](*Chore), *list.List, *list.List) {
	users := <-UsersChan
	chores := <-ChoresChan
	choreQ := <-TodoChoreQChan
	summoningOrder := <-SummoningOrderChan
	return users, chores, choreQ, summoningOrder
}

// releases internal data structures in the reverse order in which they were acquired
func releaseInternals(users map[string](*User), chores map[string](*Chore), choreQ *list.List, summoningOrder *list.List) {
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

func writeToJsonStorageFile(filename string, v interface{}) {
	file, _ := os.Open(filename)
	marshalAndWrite(file, v)

}

func initializeStorageFile(filename string, v interface{}) {
	file, _ := os.Create(filename)
	marshalAndWrite(file, v)
}

func marshalAndWrite(file *os.File, v interface{}) {
	json_bytes, _ := json.Marshal(v)
	if err := ioutil.WriteFile(file.Name(), json_bytes, 777); err != nil {
		fmt.Printf("Error initializing storage file: %v\n", err.Error())
	}
}

func listToSlice(list *list.List) []string {
	arr := make([]string, 0)
	for elem := list.Front(); elem != nil; elem = elem.Next() {
		s, _ := elem.Value.(string)
		arr = append(arr, s)
	}
	return arr
}

func sliceToList(arr []string) *list.List {
	list := list.New()
	for _, s := range arr {
		list.PushBack(s)
	}
	return list
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}