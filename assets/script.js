// constants and globals

// backend server address
var BACKEND = location.hostname + ":8080";

// the authentication id of this user
var authid;

// connection error Message
var ERRCONNECT = "Couldn't contact server. Message Drew or Logan.";


// main
window.addEventListener("load", function() {
    // browser checks
    if (!"Notification" in window) {
        alert("Your browser does not support this website.");
        window.location = "https://www.google.com/chrome/browser/";
    }
    if (!"serviceWorker" in Navigator) {
        alert("Your browser does not support this website.");
        window.location = "https://www.google.com/chrome/browser/";
    }
    if (window.Notification.permission != "granted") {
        getPermission();
    }
    navigator.serviceWorker.register('worker.js', {scope: './'});

    // look up their login info
    authid = document.cookie;
    
    /*if (authid == "") {
        // new user
        registerUser("", displaySwitch);
    } else {
        displaySwitch();
    }*/
});

// continuously bombards for permission until granted.
function getPermission() {
    alert("This website requires notification permissions.");
    Notification.requestPermission(function(p) {
        if (p != "granted") {
            getPermission();
        } else {
            new Notification("Chore Chart", {body: "thank you for enabling Notifications."});
        }
    });
}

// an ajax request, for url, will call callback on success.
// the XMLHttpRequest object is returned for further event listeners.
function ajax(url, callback) {
    var request = new XMLHttpRequest();
    request.addEventListener("load", callback);
    request.addEventListener("error", function(e) {
        console.log(e);
        error("Connection Error", ERRCONNECT);
    })
    request.open("GET", url, true);
    request.send();
    return request;
}

// shortcut for returning id elements
function $(ele) {
    return document.getElementById(ele);
}
/*
// gets a user's status and decides which ui to display (choreboard or decidechore)
function displaySwitch() {
    var th = throbber();
    ajax("http://" + BACKEND + "/userStatus?authID=" + authid, function(e) {
        th.stop();
        if (e.target.status == 200) {
            var usr = JSON.parse(e.target.responseText);
            if (usr.Summoned) {
                decideChore();
            } else {
                populateChoreChart();
            }
        } else {
            error(e.target.status, e.target.statusText);
        }
    });
}

// clears and loads a display for deciding to or not to do a chore
function decideChore() {
    var form = document.createElement("form");

    // so they know what they're doing
    var label = document.createElement("label");
    label.innerHTML = "deadline.";
    form.appendChild(label);

    // slider input for chore deadline
    var deadline = document.createElement("input");
    deadline.type = "range";
    deadline.min = "1";
    deadline.max = "72";
    deadline.value = "1";
    deadline.id = "timespan";

    // easy to read output for slider
    var timeBox = document.createElement("p");
    timeBox.id = "timedisplay";
    timeBox.innerHTML = "Deadline.";

    // update the read output for the slider
    deadline.addEventListener("input", function(e) {
        var val = deadline.value;
        var now = new Date(Date.now() + 3600000 * val);
        timeBox.innerHTML = now.toLocaleTimeString() + " " + now.toDateString();
    });

    // update the output with the time as well
    timeBox.interval = setInterval(tickUpdate, 1000, timeBox, function() {
        var val = deadline.value;
        var now = new Date(Date.now() + 3600000 * val);
        timeBox.innerHTML = now.toLocaleTimeString() + " " + now.toDateString();
    });

    form.appendChild(deadline);
    form.appendChild(timeBox);

    // accept and decline buttons
    var accept = document.createElement("div");
    accept.innerHTML = "Accept";
    accept.id = "accept";
    accept.addEventListener("click", function(e) {
        var val = deadline.value;
        ajax("http://" + BACKEND + "/acceptChore?authID=" + authid + "&deadline=" +
            new Date(Date.now() + 3600000 * val).toUTCString(),
        function(r) {
            if (r.target.status == 200) {
                populateChoreChart();
            } else {
                error(e.status, e.statusText);
            }
        });
    });

    var deny = document.createElement("div");
    deny.innerHTML = "Decline";
    deny.id = "deny";
    deny.addEventListener("click", function(e) {
        ajax("http://" + BACKEND + "/declineChore?authID=" + authid, 
        function(r) {
            if (r.target.status == 200) {
                populateChoreChart();
            } else {
                error(e.status, e.statusText);
            }
        });
    });

    accept.classList.add("button");
    deny.classList.add("button");
    form.appendChild(accept);
    form.appendChild(deny);

    $("wrapper").appendChild(form);
}

// Interval event that calls callback, stops on element DOM removal
// requires element has interval field set to the interval id
function tickUpdate(element, callback) {
    if (element.parentNode == null) {
        clearInterval(element.interval);
    }
    callback();

}

*/
function activateChore(chore) {
    ajax("http://" + BACKEND + "/reportChore?authID=" + 
        authid + "&choreName=" + chore.ChoreName, 
        function(e) {
            if (e.target.status == 200) {
                console.log("Success.");
            } else {
                error(e.target.status, e.target.statusText);
            }
    });
}

function deactiveChore(chore) {
    ajax("http://" + BACKEND + "/doneWithChore?authID=" + 
        authid + "&choreName=" + chore.ChoreName, 
        function(e) {
            if (e.target.status == 200) {
                console.log("Success.");
            } else {
                error(e.target.status, e.target.statusText);
            }
    });
}

// registers a user, either creating a new account or reauthenticating them
// takes a string to display for the first prompt, or empty string for the 
// default prompt, and a callback after a successful registration. 
// Will continue to query backend server and prompt until successful. 
function registerUser(str, callback) {
    // create the prompts
    var prompt = newPrompt(false);
    var text = document.createElement("h1");
    text.innerHTML = (str? str : "Welcome!");
    prompt.appendChild(text);
    var input = document.createElement("input");
    input.placeholder = "Friendly Name";
    input.autofocus = true;
    prompt.opened = true;

    // variables to hold the user information
    var friendlyName;
    var password;
    
    input.addEventListener("change", function(e) {
        // first or second prompt splitting
        if (!friendlyName) {
            friendlyName = input.value;
            // change the prompt for the second round input and
            // clear the input box
            input.value = "";
            input.placeholder = "Secret Phrase";
        } else {
            password = input.value;
            // remove the prompt and query the backend
            prompt.kill();
            var th = throbber();
            ajax("http://" + BACKEND + "/loginUser?friendlyName=" + 
                friendlyName + "&password=" + password,
                // callback on backend response
                function(e) {
                    // if we're not successful, try again
                    th.stop();
                    if (e.target.status != 200) {
                        // a wrong password or just can't connect.
                        if (e.target.status == 403) {
                            registerUser("Incorrect secret phrase, try again.",
                            callback);
                        } else {
                            registerUser("Server Denied Request, try again.",
                             callback);
                        }
                    } else {
                        // we reach the server, parse and save the response.
                        var res = JSON.parse(e.target.responseText);
                        authid = res;
                        if (authid == "") {
                            registerUser("Incorrect secret phrase, try again.",
                             callback);
                        } else {
                            document.cookie = authid;
                            callback();
                        }
                    }
                }
            );
        }
    });

    // event listener for user input
    
    // put it all together and display the prompt
    prompt.appendChild(input);
}

// creates and returns a prompt box for customization.
// use prompt.kill() to remove the prompt box.
function newPrompt(exitable) {
    var prompt = document.createElement("paper-dialog");
    prompt.modal = !exitable;
    prompt.opened = true;
    document.body.appendChild(prompt);
    prompt.kill = function() {
        prompt.close();
    }
    return prompt;
}
// presents a non-closable fatal error to the user
function error(title, body) {
    var prompt = newPrompt(false);
    var head = document.createElement("p");
    var text = document.createElement("p");
    prompt.style.backgroundColor = "rgba(255, 0, 0, 0.5)";
    head.style.fontWeight = "900";
    head.innerHTML = title;
    text.innerHTML = body;
    prompt.appendChild(head);
    prompt.appendChild(text);
}

// sets up a throbber and returns it for removal
function throbber() {
    var prog = $("prog");
    prog.disabled = false;

    throbber.stop = function() {
        prog.disabled = true;
    }

    return throbber;
}