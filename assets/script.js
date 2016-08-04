(function() {
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
        var authid = document.cookie;
        if (authid == "") {
            // new user
            registerUser();
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

    function registerUser() {
        var nono = document.createElement("div");
        nono.id = "nonosquare";
        var prompt = document.createElement("div");
        prompt.id = "prompt";
        var text = document.createElement("p");
        text.innerHTML = "Enter a friendly name:";
        prompt.appendChild(text);
        var input = document.createElement("input");
        input.type = "text";

        var friendlyName;
        var password;

        input.addEventListener("change", function(e) {
            if (!friendlyName) {
                friendlyName = input.value;
                input.value = "";
                text.innerHTML = "secret phrase:";
            } else {
                password = input.value;
                console.log(friendlyName + ", " + password);
                nono.parentNode.removeChild(nono);
            }
        });

        prompt.appendChild(input);
        nono.appendChild(prompt);
        document.body.insertBefore(nono, null);
    }

})();