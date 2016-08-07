// Service Worker Code
// Logan Girvin
// Copyright 2016


function ajax(url, callback) {
    var request = new XMLHttpRequest();
    request.addEventListener("load", callback);
    request.addEventListener("error", console.log);
    request.open("GET", url, true);
    request.send();
    return request;
}

self.addEventListener('install', function(event) {
  self.skipWaiting();
  console.log('Installed', event);
});
self.addEventListener('activate', function(event) {
  console.log('Activated', event);
  
//  var service = self;
//  setInterval(function() {
//      service.registration.showNotification("Service Worker Activiated");
//  }, 5000);
  fetch('http://localhost:8080/register').then(
      function(response) {
          console.log(response);
      }
  ).catch(function(error) {
      console.log(error);
  });
});