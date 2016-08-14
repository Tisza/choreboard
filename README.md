# ChoreBoard

Choreboard is a web application used to manage chores in a large-ish household using Go and vanilla Javascript.

ChoreBoard is designed to be run on a private network, where users/housemates are able to report chores that need to be done, and the service will notify other users/housemates that there's work to-do. ChoreBoard will track and report various metrics of users like chores completed, chores skipped, etc, and compare users with each other in an effort to drive competition in the house (to be the least shitty roommate).

## Installation

1. Install [Go](https://golang.org/)
2. Setup a Go workspace ([How to Write Go Code](https://golang.org/doc/code.html))
3. Clone this repository into the src folder of your Go workspace
4. Resolve dependencies
     ```sh
     $ go get github.com/Workiva/go-datastructures/queue
     ```
        
     ```sh
     $ go get github.com/emirpasic/gods/sets/hashset
     ```
5. Start up the service
     ```sh
     $ go run $GOPATH/src/choreboard/controller/ChoreController.go
     ```
        
From there, you should have the service running on `localhost:8080`. 

6. Install your favorite web server (ex. [Nginx](https://www.nginx.com/))
7. Configure your server's root directory to the clone you just made (jank AF)
8. Go to your favorite web browser and go to the server's IP

NOTE: As of now, only certain users are allowed to log into the service, and new users cannot be created.

