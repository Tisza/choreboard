function getPermission(){alert("This website requires notification permissions."),Notification.requestPermission(function(e){"granted"!=e?getPermission():new Notification("Chore Chart",{body:"thank you for enabling Notifications."})})}function ajax(e,t){if(disconnect){var r=$("prog");return void(r.disabled=!0)}var n=new XMLHttpRequest;return n.addEventListener("load",t),n.addEventListener("error",function(e){if(!disconnect){disconnect=!0,toast("Cannot connect to home");var t=document.createElement("h1");t.id="block warning",t.innerHTML="Cannot connect to home",document.body.appendChild(t)}}),n.open("GET",e,!0),n.send(),n}function $(e){return document.getElementById(e)}function registerUser(e,t){if(prompted)return!1;prompted=!0;var r=newPrompt(!1),n=document.createElement("h1");n.innerHTML=e?e:"Welcome!",r.appendChild(n);var o=document.createElement("input");o.placeholder="Friendly Name",o.autofocus=!0,r.opened=!0;var i,a;return o.addEventListener("change",function(e){if(i){a=o.value,r.kill();var n=throbber();ajax("http://"+BACKEND+"/loginUser?friendlyName="+i+"&password="+a,function(e){if(n.stop(),prompted=!1,200!=e.target.status)403==e.target.status?registerUser("Incorrect secret phrase, try again.",t):registerUser("Server Denied Request, try again.",t);else{var r=JSON.parse(e.target.responseText);authid=r,""==authid?registerUser("Incorrect secret phrase, try again.",t):(document.cookie="authid="+authid,document.cookie="friendlyName="+i,device.postMessage({authid:authid,friendlyName:i}),t())}})}else i=o.value,o.value="",o.placeholder="Secret Phrase"}),r.appendChild(o),!0}function toast(e){var t=document.createElement("paper-toast");document.body.appendChild(t),t.show({text:e,duration:3e3}),setTimeout(function(e){t.parentNode.removeChild(t)},3e3)}function newPrompt(e){var t=document.createElement("paper-dialog");return t.modal=!e,t.opened=!0,document.body.appendChild(t),t.kill=function(){t.close()},t}function error(e,t){if(!errored){errored=!0;var r=newPrompt(!1),n=document.createElement("p"),o=document.createElement("p");r.style.backgroundColor="rgba(255, 0, 0, 0.5)",n.style.fontWeight="900",n.innerHTML=e,o.innerHTML=t,r.appendChild(n),r.appendChild(o)}}function throbber(){var e=$("prog");return e.disabled=!1,throbber.stop=function(){e.disabled=!0},throbber}var BACKEND="choreboard:8080",SERVICEWORKER="worker.js",authid,friendlyName,prompted=!1,errored=!1,disconnect=!1,device;window.addEventListener("load",function(){!1 in window&&(alert("Your browser does not support this website."),window.location="https://www.google.com/chrome/browser/"),!1 in Navigator&&(alert("Your browser does not support this website."),window.location="https://www.google.com/chrome/browser/"),"granted"!=window.Notification.permission&&getPermission(),navigator.serviceWorker.register(SERVICEWORKER,{scope:"./"}).then(function(e){device=e.active,authid&&friendlyName&&device.postMessage({authid:authid,friendlyName:friendlyName})});var e=document.cookie.split(";");e.forEach(function(e){var t=e.split("=");if(2==t.length)switch(t[0].toLowerCase().trim()){case"authid":authid=t[1];break;case"friendlyname":friendlyName=t[1]}}),""!=authid&&""!=friendlyName||(document.cookie="",registerUser("",function(e){}))});