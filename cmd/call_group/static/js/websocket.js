let ws;

function connectWebSocket() {
    ws = new WebSocket('ws://' + window.location.host + '/ws');

    ws.onopen = function() {
        const msg = {
            type: 'join',
            room: room,
            username: username
        };
        ws.send(JSON.stringify(msg));
    };

    ws.onmessage = async function(event) {
        const message = JSON.parse(event.data);
        
        switch(message.type) {
            case 'chat':
                displayMessage(message);
                break;
            case 'user_joined':
                displaySystemMessage(`${message.username} joined the room`);
                console.log("user joined =>")
                console.log(username)
                console.log(message.username)
                if (message.username !== username) {
                    createPeerConnection(message.username, true);
                }
                break;
            case 'user_left':
                displaySystemMessage(`${message.username} left the room`);
                if (peerConnections[message.username]) {
                    peerConnections[message.username].close();
                    delete peerConnections[message.username];
                }
                break;
            case 'user_list':
                const users = message.content ? message.content.split(',') : [];
                console.log("users list =>")
                console.log(users)
                updateUsersList(users);
                users.forEach(user => {
                    if (user !== username && !peerConnections[user]) {
                        createPeerConnection(user, true);
                    }
                });
                break;
            case 'signal':
                handleSignalingMessage(message);
                break;
        }
    };

    ws.onclose = function() {
        displaySystemMessage('Connection lost. Trying to reconnect...');
        setTimeout(connectWebSocket, 1000);
    };
}

function sendSignal(target, data) {
    const signal = {
        type: 'signal',
        target: target,
        username: username,
        signal: JSON.stringify(data)
    };
    ws.send(JSON.stringify(signal));
}

function sendMessage() {
    const messageInput = document.getElementById('message-input');
    const message = messageInput.value.trim();
    if (message && ws) {
        const msg = {
            type: 'chat',
            room: room,
            username: username,
            content: message
        };
        ws.send(JSON.stringify(msg));
        messageInput.value = '';
    }
}