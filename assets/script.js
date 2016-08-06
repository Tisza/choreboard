(function() {
    // constants and globals

    // backend server address
    var BACKEND = location.host + ":8080";
    
    // the authentication id of this user
    var authid;

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
        if (authid == "") {
            // new user
            registerUser("", populateChoreChart);
        } else {
            populateChoreChart();
        }
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
        request.open("GET", url, true);
        request.send();
        return request;
    }

    // shortcut for returning id elements
    function $(ele) {
        return document.getElementById("board");
    }

    // re-writes the main wrapper to display the current chore-chart.
    function populateChoreChart() {
        var board = $("board");
        board.innerHTML = "";
        var throbber = document.createElement("div");
        throbber.id = "throbber";
        board.appendChild(throbber);
    }

    // registers a user, either creating a new account or reauthenticating them
    // takes a string to display for the first prompt, or empty string for the 
    // default prompt, and a callback after a successful registration. 
    // Will continue to query backend server and prompt until successful. 
    function registerUser(str, callback) {
        // create the prompts
        var nono = document.createElement("div");
        nono.id = "nonosquare";
        var prompt = document.createElement("div");
        prompt.id = "prompt";
        var text = document.createElement("p");
        text.innerHTML = (str? str : "Enter a friendly name:");
        prompt.appendChild(text);
        var input = document.createElement("input");
        input.type = "text";

        // variables to hold the user information
        var friendlyName;
        var password;

        // event listener for user input
        input.addEventListener("change", function(e) {
            // first or second prompt splitting
            if (!friendlyName) {
                friendlyName = input.value;
                // change the prompt for the second round input and
                // clear the input box
                input.value = "";
                text.innerHTML = "secret phrase:";
            } else {
                password = input.value;
                // remove the prompt and query the backend
                nono.parentNode.removeChild(nono);
                
                ajax("http://" + BACKEND + "/loginUser?friendlyName=" + 
                    friendlyName + "&password=" + password,
                    // callback on backend response
                    function(e) {
                        // if we're not successful, try again
                        if (e.target.status != 200) {
                            // a wrong password or just can't connect.
                            if (e.target.status == 403) {
                                registerUser("Incorrect secret phrase, try again." +
                                    " Friendly Name:", callback);
                            } else {
                                registerUser("Server Denied Request, try again." + 
                                    " Friendly Name:", callback);
                            }
                        } else {
                            // we reach the server, parse and save the response.
                            console.log(e.target.responseText);
                            var res = JSON.parse(e.target.responseText);
                            authid = res.authID;
                            document.cookie = authid;
                            callback();
                        }
                    }
                ).addEventListener("error", function() {
                    // also prompt if we couldn't connect to the backend.
                    registerUser("Couldn't reach server, try again. Friendly Name:",
                        callback);
                });
            }
        });

        // put it all together and display the prompt
        prompt.appendChild(input);
        nono.appendChild(prompt);
        document.body.insertBefore(nono, null);
        input.focus();
    }

})();