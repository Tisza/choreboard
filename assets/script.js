(function() {
    // constants and globals

    // backend server address
    var BACKEND = location.host + ":8080";
    
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

        // link the navigation
        $("navcb").addEventListener("click", function(e) {
            populateChoreChart();
        });

        // look up their login info
        authid = document.cookie;
        
        if (authid == "") {
            // new user
            registerUser("", displaySwitch);
        } else {
            displaySwitch();
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
        return document.getElementById(ele);
    }

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
                chores.forEach(function(value, index) {
                    // make the dom object for it and appendChild
                    var item = document.createElement("div");
                    item.chore = value; // saved for later
                    item.classList.add("item");
                    var title = document.createElement("h2");
                    title.innerHTML = value.ChoreName;
                    item.appendChild(title);
                    var icon = document.createElement("p");
                    icon.classList.add("img");
                    icon.innerHTML = value.ChoreName.substring(0, 1);
                    // random colors are my favorite
                    var r = Math.round(Math.random() * 123);
                    var g = Math.round(Math.random() * 123);
                    var b = Math.round(Math.random() * 123);
                    icon.style.backgroundColor = "rgb(" + r + ", " + g + ", " + b + ")";
                    item.appendChild(icon);
                    var text = document.createElement("div");
                    item.appendChild(text);
                    var desc = document.createElement("p");
                    desc.innerHTML = value.Description;
                    desc.classList.add("desc");
                    text.appendChild(desc);
                    var who = document.createElement("p");
                    who.classList.add("who");
                    if (value.NeedsWork) {
                        item.classList.add("active");
                    }
                    if (value.Assignee) {
                        who.innerHTML = value.Assignee;
                    } else {
                        who.innerHTML = "Not assigned.";
                    }
                    text.appendChild(who);
                    board.appendChild(item);
                    value.dom = item;
                    item.addEventListener("click", function(e) {
                        promptChore(value, e);
                    });
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

    // event listener for choreboard items to flip/flop the chore value
    // REQUIRES: chore.dom to be set to the dom element it is firing for.
    function promptChore(chore, event) {
        var prompt = newPrompt(true);
        var text = document.createElement("p");
        text.style.fontWeight = "900";
        var button = document.createElement("div");
        button.classList.add("promptAction");
        // which of the two to prompts to chore.
        if (chore.NeedsWork) {
            text.innerHTML = "Did you " + chore.ChoreName + "?";
            button.innerHTML = "I did " + chore.ChoreName;
            // sign the chore as inactive.
            button.addEventListener("click", function() {
                prompt.kill();
                ajax("http://" + BACKEND + "/reportChore?authID=" + 
                    authid + "&choreName=" + chore.ChoreName + "&mode=false", 
                    function(e) {
                        if (e.target.status == 200) {
                            chore.dom.classList.remove("active");
                            chore.NeedsWork = false;
                        } else {
                            error(e.target.status, e.target.statusText);
                        }
                    });
            });
        } else {
            text.innerHTML = "Does " + chore.ChoreName + " need to be done?";
            button.innerHTML = chore.ChoreName + " needs to be done.";
            // sign the chore as active
            button.addEventListener("click", function() {
                prompt.kill();
                ajax("http://" + BACKEND + "/reportChore?authID=" + 
                    authid + "&choreName=" + chore.ChoreName + "&mode=true", 
                    function(e) {
                        if (e.target.status == 200) {
                            chore.dom.classList.add("active");
                            chore.NeedsWork = true;
                        } else {
                            error(e.target.status, e.target.statusText);
                        }
                    });
            });
        }
        prompt.appendChild(text);
        prompt.appendChild(button);
    }

    // registers a user, either creating a new account or reauthenticating them
    // takes a string to display for the first prompt, or empty string for the 
    // default prompt, and a callback after a successful registration. 
    // Will continue to query backend server and prompt until successful. 
    function registerUser(str, callback) {
        // create the prompts
        var prompt = newPrompt(false);
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
                prompt.kill();
                
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
                            var res = JSON.parse(e.target.responseText);
                            authid = res;
                            if (authid == "") {
                                registerUser("Incorrect secret phrase, try again." + 
                                   " Friendly Name:", callback);
                            } else {
                                document.cookie = authid;
                                callback();
                            }
                        }
                    }
                );
            }
        });

        // put it all together and display the prompt
        prompt.appendChild(input);
        input.focus();
    }

    // creates and returns a prompt box for customization.
    // use prompt.kill() to remove the prompt box.
    function newPrompt(exitable) {
        var prompt = document.createElement("div");
        var nono = document.createElement("div");
        if (exitable) {
            var ex = document.createElement("div");
            ex.id = "exitBox";
            prompt.appendChild(ex);
            ex.addEventListener("click", function() {
                nono.parentNode.removeChild(nono);
            });
            nono.addEventListener("click", function(e) {
                if (e.target.id == "nonosquare") {
                    nono.parentNode.removeChild(nono);
                }
            });
        }
        prompt.id = "prompt";
        nono.id = "nonosquare";
        nono.appendChild(prompt);
        document.body.appendChild(nono);
        prompt.kill = function() {
            nono.parentNode.removeChild(nono);
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

    // function for setInterval on throbbers to alter its color
    // the param is the throbber to animate
    // REQUIRES throbber t has field interval set to the 
    // setInterval this is called at.
    function throbberHelper(t) {
        var time = t.tic;
        if (!time) {
            time = 0;
        }
        time = time + 0.05;
        var v = Math.round(20 * (Math.sin(time))) + 30;
        var h = Math.round(120 + 60 * Math.cos(time / 2));
        t.style.backgroundColor = "hsl(" + h + ", 50%, " + v + "%)";
        t.tic = time;
    }

    // sets up a throbber and returns it for removal
    function throbber() {
        var board = $("board");
        board.innerHTML = "";
        var throbber = document.createElement("div");
        throbber.id = "throbber";
        board.appendChild(throbber);
        throbber.interval = setInterval(tickUpdate, 100, throbber, function() {
            throbberHelper(throbber);
        });

        throbber.stop = function() {
            clearInterval(throbber.interval);
            throbber.parentNode.removeChild(throbber);
        }

        return throbber;
    }

})();
