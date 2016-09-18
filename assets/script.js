// constants and globals

// backend server address
var BACKEND = location.hostname + ":8080";
var SERVICEWORKER = "https://tisza.github.io/choreboard/";

// the authentication id of this user
var authid;
var friendlyName;

// private variable
var prompted = false;
var errored = false;
var disconnect = false;

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
    /*navigator.serviceWorker.register(SERVICEWORKER, 
        {scope: './'}).then(function(registration) {
            device = registration.active;
            if (authid && friendlyName) {
                device.postMessage({authid: authid, friendlyName: friendlyName});
            }
    });*/

    // look up their login info
    var cookies = document.cookie.split(";");
    cookies.forEach(function(value) {
        var parts = value.split("=");
        if (parts.length == 2) {
            switch (parts[0].toLowerCase().trim()) {
                case "authid":
                    authid = parts[1];
                    break;
                case "friendlyname":
                    friendlyName = parts[1];
                    break;
                default:
                    break;
            }
        }
    });

    if (authid == "" || friendlyName == "") {
        document.cookie = "";
        registerUser("", function(e) {});
    }
});

// an ajax request, for url, will call callback on success.
// the XMLHttpRequest object is returned for further event listeners.
function ajax(url, callback) {
    if (disconnect) {
        var prog = $("prog");
        prog.disabled = true;
        return;
    }
    var request = new XMLHttpRequest();
    request.addEventListener("load", callback);
    request.addEventListener("error", function(e) {
        if (!disconnect) {
            disconnect = true;
            toast("Cannot connect to home");
            var header = document.createElement("h1");
            header.id = "block warning";
            header.innerHTML = "Cannot connect to home";
            document.body.appendChild(header);
        }
    })
    request.open("GET", url, true);
    request.send();
    return request;
}

// shortcut for returning id elements
function $(ele) {
    return document.getElementById(ele);
}

// registers a user, either creating a new account or reauthenticating them
// takes a string to display for the first prompt, or empty string for the 
// default prompt, and a callback after a successful registration. 
// Will continue to query backend server and prompt until successful. 
function registerUser(str, callback) {
    if (prompted) 
        return false;
    prompted = true;
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
                    prompted = false;
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
                            document.cookie = "authid=" + authid;
                            document.cookie = "friendlyName=" + friendlyName;
                            //device.postMessage({authid: authid, friendlyName: friendlyName});
                            //callback();
                            window.location = SERVICEWORKER + "?authid=" + authid;
                        }
                    }
                }
            );
        }
    });

    // event listener for user input
    
    // put it all together and display the prompt
    prompt.appendChild(input);
    return true;
}

function toast(message) {
    var t = document.createElement("paper-toast");
    document.body.appendChild(t);
    t.show({text: message, duration: 3000});
    setTimeout(function(e) {
        t.parentNode.removeChild(t);
    }, 3000);
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
    if (errored) {
        return;
    }
    errored = true;
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