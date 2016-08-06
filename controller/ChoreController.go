package main

import (
	"fmt"
	"net/http"
	"net"
	"errors"
	"choreboard/model"
)

/**
	TODO: implement handleUserStatus
	TODO: implement handleSignChore
	TODO: implement handleChoreBoard
	TODO: implement handleLoginUser
	TODO: implement handleReportChore
 */

var HOST_NAME , err = externalIP()
var PORT string = "8080"
var HOST = HOST_NAME + ":" + PORT

var USER_STATUS_PARAMS = []string{"authID"}
var SIGN_CHORE_PARAMS = []string{"authID", "accept"}
var CHORE_BOARD_PARAMS = []string{"authID"}
var LOGIN_USER_PARAMS = []string{"friendlyName", "password"}
var REPORT_CHORE_PARAMS = []string{"authID", "chore", "mode"}


func main() {

	http.HandleFunc("/userStatus", middleware(handleUserStatus, USER_STATUS_PARAMS))
	http.HandleFunc("/signChore", middleware(handleSignChore, SIGN_CHORE_PARAMS))
	http.HandleFunc("/choreBoard", middleware(handleChoreBoard, CHORE_BOARD_PARAMS))
	http.HandleFunc("/loginUser", middleware(handleLoginUser, LOGIN_USER_PARAMS))
	http.HandleFunc("/reportChore", middleware(handleReportChore, REPORT_CHORE_PARAMS))

	fmt.Println("About to ListenAndServe on " + HOST)
	http.ListenAndServe(HOST, nil)
}

//=============================== Handlers ===============================//


func middleware(next http.HandlerFunc, expectedParams []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isBadRequest(w, r, expectedParams) {
			next.ServeHTTP(w, r)
		}
	})
}


func handleUserStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Executed userStatus handler!\n")
	fmt.Fprintf(w, model.Do())
}

func handleSignChore(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Executed signChore handler!\n")

}

func handleChoreBoard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Executed choreBoard handler!\n")

}

func handleLoginUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Executed loginUser handler!\n")

}

func handleReportChore(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Executed reportChore handler!\n")

}

//=============================== End of Handlers ===========================//


//=============================== Helpers ===========================//


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

func getParams(r *http.Request) ([]string) {
	params := make([]string, 0)
	for param := range r.Form {
		//fmt.Println(param)
		params = append(params, param)
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
	if !sliceEq(requestParams, expectedParams) {
		// uh oh, we got a 400 Bad Request over 'ere
		w.WriteHeader(http.StatusBadRequest)
		return true
	} else {
		return false
	}
}

func sliceEq(a, b []string) bool {
	if a == nil && b == nil {
		return true;
	}

	if a == nil || b == nil {
		return false;
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

//=============================== End of Helpers ===========================//
