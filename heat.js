function promptPassword() {
    return prompt("Please enter tonight's meeting password.");
}

// authenticate
function auth() {
    if (!authenticated) {
        let password = promptPassword();

        // synchronous request to /authenticate
        var xhttp = new XMLHttpRequest();
        xhttp.open("POST", "/authenticate", false);
        xhttp.setRequestHeader("Content-Type", "application/json");

        let event = {
            "id": wsID,
            "password": password
        };

        xhttp.send(JSON.stringify(event));

        authenticated = xhttp.status == 200;
    }

    return authenticated;
}

// Ask to create an entry
function create() {
    if (!auth()) {
        return
    }

    // Get values
    let name = document.getElementById("name");
    let type = document.getElementById("type");
    let desc = document.getElementById("description");

    // Check for errors
    if (!name || !type || !desc) {
        return;
    }

    // Create event
    let event = {
        "event": "Create",
        "name": name.value,
        "talk_type": type.value,
        "desc": desc.value,
    };

    name.value = ""
    type.value = ""
    desc.value = ""

    // Send it
    checkAndReset().await;

    websocket.send(JSON.stringify(event));
}

// Ask to hide an entry
function hide(id) {
    if (!auth()) {
        return
    }

    let event = {
        "event": "Hide",
        "id": id,
    };

    // Send it
    checkAndReset().await;
    websocket.send(JSON.stringify(event));
}

var websocket = null;
var wsID = null;
var authenticated = false;

window.onload = function () {
    // Tie pressing enter on the description field to the create button
    let button = document.getElementById("create");
    document.getElementById("description").addEventListener("keydown",
        function (event) {
            if (!event) {
                var event = window.event;
            }
            if (event.keyCode == 13) {
                button.click();
            }
        }, false);

    register();
};

async function checkAndReset() {
    // Check if websocket is running
    if (!websocket) {
        register();
    } else if (websocket.readyState != WebSocket.OPEN && websocket.readyState != WebSocket.CONNECTING) {
        // sync the visible talks on screen
        fetch("/talks")
            .then(function (response) {
                return response.json();
            })
            .then(function (result) {
                // Empty the table
                var rows = document.getElementById('tb').children;

                for (let i = rows.length - 2; i >= 0; i--) {
                    rows[i].remove();
                }

                result.forEach(talk => addTalk(talk));
            });

        // get a new websocket
        console.log("connection closed getting new connection");
        register();
    }
}
// Periodically check that the websocket is open, if not create a new one
setInterval(checkAndReset, 10000);

var ordering = {}
ordering["forum topic"] = 1
ordering["lightning talk"] = 2
ordering["project update"] = 3
ordering["announcement"] = 4
ordering["after meeting slot"] = 5

function register() {
    // Register a websocket connection
    fetch("/register")
        .then(function (response) {
            return response.json();
        })
        .then(function (result) {
            authenticated = result.authenticated;
            if (window.location.protocol[4] == 's') {
                websocket = new WebSocket("wss://" + "0.0.0.0" + "/socket/" + result.id);
            } else {
                websocket = new WebSocket("ws://" + "0.0.0.0" + "/socket/" + result.id);
            }

            websocket.onmessage = function (event) {
                let json = JSON.parse(event.data);

                if (json.event == "Show") {
                    addTalk(json);
                } else if (json.event == "Hide") {
                    // Remove the row with matching id
                    var rows = document.getElementById('tb').children;

                    for (i = 0; i < rows.length - 1; i++) {
                        if (json.id == rows[i].children[0].innerHTML) {
                            rows[i].remove();
                            break;
                        }
                    }
                }
            }
            wsID = result.id;
        })
        .catch(function (error) {
            console.log("Error: " + error);
        });
}

function addTalk(json) {
    var table = document.getElementById('table');
    var rows = document.getElementById('tb').children;

    // Insert the new data into the correct location in the table
    let i = 0
    for (i = 0; i < rows.length - 1; i++) {
        // Order by talk type then by id

        let order = ordering[rows[i].children[2].innerText];
        let id = rows[i].children[0].innerText;

        if (ordering[json.talk_type] < order) {
            break;
        }
    }

    // Building a new event object using _javascript_
    var row = table.insertRow(i + 1);
    row.setAttribute("class", "event");

    var c0 = row.insertCell(0);
    c0.setAttribute("style", "display: none;");
    c0.innerHTML = json.id;

    var c1 = row.insertCell(1);
    c1.setAttribute("class", "name");
    c1.innerHTML = json.name;

    var c2 = row.insertCell(2);
    c2.setAttribute("class", "type");
    c2.innerHTML = json.talk_type;

    var c3 = row.insertCell(3);
    c3.setAttribute("class", "desc");
    c3.innerHTML = json.description;

    var c4 = row.insertCell(4);
    c4.setAttribute("class", "actions");
    c4.innerHTML = '<button onclick="hide(' + json.id + ')"> x </button>';

}

// Setup the nav bar
fetch('https://dubsdot.cslabs.clarkson.edu/cosi-nav.json')
    .then(res => res.json())
    .then(json => {
        let links = json.links;
        let linkList = document.querySelector('cosi-nav');
        links.forEach(link => {
            let newLink = document.createElement('li');
            newLink.innerHTML = `<a href="${link.url}">${link.name}</a>`;
            linkList.appendChild(newLink);
        });
    });
