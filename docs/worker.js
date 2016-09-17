// Service Worker Code
// Logan Girvin
// Copyright 2016


function ajax(url, callback, error) {
    fetch(url).then(callback).catch(error);
}

var FRONT = "https://tisza.github.io/choreboard/index.html";
var BACK = "http://choreboard:8080";

var seconds = 1000; // multiplier for timeouts

function responseWrapper(authid, friendlyName) {
    var h = new Headers();
    h.append("authid", authid);
    h.append("friendlyname", friendlyName);
    return new Response("", {headers: h});
}

function responseDigest(res) {
    var h = res.headers;
    var authid = h.get("authid");
    var friendlyName = h.get("friendlyname");
    return {"authid" : authid, "friendlyName": friendlyName};
}

self.addEventListener("message", function(e) {
    var data = e.data;
    var cache = self.caches;
    caches.open("choreBoardUserInfo").then(function(cache) {
        cache.put("userInfo", responseWrapper(data.authid, data.friendlyName));
    });
});

self.addEventListener("notificationclick", function(e) {
    clients.openWindow(FRONT);
    e.notification.close();
});

function signIn() {
    self.registration.showNotification("Please sign in", 
        {
            //actions: [{action: "", title: "Log In"}],
            body: "We can't check your status in the background!"
        });
    var timeout = 30 * 60 * seconds;
    setTimeout(statusCheck, timeout);
}

function promptUser() {
    self.showNotification("ChoreBoard", 
    {
        body: "A chore has been assigned to you."
    });
}

function readStatus(e) {
    e.json().then(function(json) {
        if (json.Summoned) {
            promptUser();
        }
    });
}

function error(e) {
    self.registration.showNotification(e.status, 
        {
            body: e.statusText
        });
}

function statusCheck() {
    caches.match("userInfo").then(function(res) {
        var obj = responseDigest(res);
        var authid = obj.authid;
        var friendlyName = obj.friendlyName;
        // make the call.
        if (!authid) {
            signIn();
        } else {
            // have authid
            ajax(BACK + "/userStatus?authID=" + authid, function(e) {
                if (e.status == 200) {
                    readStatus(e);
                } else if (e.status == 403) {
                    signIn();
                } else {
                    error(e);
                }
                var timeout = 60 * seconds;
                setTimeout(statusCheck, timeout);
            }, function(e) { // not connected to the network
                var timeout = 15 * 60 * seconds;
                setTimeout(statusCheck, timeout);
            });
        }
    });
}

self.addEventListener('install', function(event) {
  self.skipWaiting();
  self.registration.showNotification("Choreboard", 
    {
        body: "A new version of choreboard notifier has been installed!"
    });
});

self.addEventListener('activate', function(event) {
    statusCheck();
});