(function() {
    // constants and globals

    // backend server address
    var BACKEND = location.host + ":8080";
    
    // the authentication id of this user
    var authid;

    // connection error Message
    var ERRCONNECT = "Couldn't contact server. Message Drew or Logan.";


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
        return document.getElementById("board");
    }

    // re-writes the main wrapper to display the current chore-chart.
    function populateChoreChart() {
        var th = throbber();

        ajax("http://" + BACKEND + "/choreBoard?authID=" + authid,
        function(e) {
            // success
            if (e.target.status == 200) {
                // get the object and run through each chore
                var chores = JSON.parse(e.target.responseText);
                var board = $("board");
                chores.chores.forEach(function(value, index) {
                    // make the dom object for it and appendChild
                    var item = document.createElement("div");
                    item.chore = value; // saved for later
                    item.classList.add("item");
                    var title = document.createElement("h2");
                    title.innerHTML = value.choreName;
                    item.appendChild(title);
                    var icon = document.createElement("p");
                    icon.classList.add("img");
                    icon.innerHTML = value.choreName.substring(0, 1);
                    // random colors are my favorite
                    var r = Math.round(Math.random() * 123);
                    var g = Math.round(Math.random() * 123);
                    var b = Math.round(Math.random() * 123);
                    icon.style.backgroundColor = "rgb(" + r + ", " + g + ", " + b + ")";
                    item.appendChild(icon);
                    var text = document.createElement("div");
                    item.appendChild(text);
                    var desc = document.createElement("p");
                    desc.innerHTML = value.description;
                    desc.classList.add("desc");
                    text.appendChild(desc);
                    var who = document.createElement("p");
                    who.classList.add("who");
                    if (value.active) {
                        who.innerHTML = value.assignee;
                        item.classList.add("active");
                    } else {
                        who.innerHTML = "Not assigned.";
                    }
                    text.appendChild(who);
                    board.appendChild(item);
                });
            // incorrect auth id
            } else if (e.target.status == 403) {
                document.cookie = "";
                authid = "";
                registerUser("Please sign in. Friendly Name:", populateChoreChart);
            } else {
                // some other error
                error(e.target.status, e.target.statusText);
            }
            th.stop();
        });
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
                );
            }
        });

        // put it all together and display the prompt
        prompt.appendChild(input);
        nono.appendChild(prompt);
        document.body.insertBefore(nono, null);
        input.focus();
    }

    // presents a non-closable fatal error to the user
    function error(title, body) {
        var prompt = document.createElement("div");
        var nono = document.createElement("div");
        var head = document.createElement("p");
        var text = document.createElement("p");
        prompt.id = "prompt";
        nono.id = "nonosquare";
        prompt.style.backgroundColor = "rgba(255, 0, 0, 0.5)";
        head.style.fontWeight = "900";
        head.innerHTML = title;
        text.innerHTML = body;
        prompt.appendChild(head);
        prompt.appendChild(text);
        nono.appendChild(prompt);
        document.body.appendChild(nono);
    }

    // function for setInterval on throbbers to alter its color
    // the param is the throbber to animate
    // REQUIRES throbber t has field interval set to the 
    // setInterval this is called at.
    function throbberHelper(t) {
        if (t.parentNode == null) {
            clearInterval(t.interval);
        }
        var time = t.tic;
        if (!time) {
            time = 0;
        }
        time = time + 0.05;
        var v = Math.round(20 * (Math.sin(time))) + 30;
        var h = Math.round(120 + 60 * Math.cos(time / 2));
        t.style.backgroundColor = "hsl(" + h + ", 50%, " + v + "%)";
        t.hue = h;
        t.tic = time;
    }

    // sets up a throbber and returns it for removal
    function throbber() {
        var board = $("board");
        board.innerHTML = "";
        var throbber = document.createElement("div");
        throbber.id = "throbber";
        board.appendChild(throbber);
        throbber.interval = setInterval(throbberHelper, 100, throbber);

        throbber.stop = function() {
            clearInterval(throbber.interval);
            throbber.parentNode.removeChild(throbber);
        }

        return throbber;
    }

})();