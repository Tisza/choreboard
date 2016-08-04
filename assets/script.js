(function() {
    window.addEventListener("load", function() {
        if (!"Notification" in window) {
            alert("Your browser does not support this website.");
            window.location = "";
        }
        if (window.Notification.permission != "granted") {
            getPermission();
        }
        if (!"serviceWorker" in Navigator) {
            alert("Your browser does not support this website.");
            window.location = "";
        }
        navigator.serviceWorker.register('worker.js', {scope: './'}).then(
            function(registration) {
                console.log("Registration success.");
            }
        ).catch(function(error) {
            console.log("Error: ");
            console.log(error);
        });
    });

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
})();