let username;
let room;

async function joinRoom() {
    username = document.getElementById('username-input').value.trim();
    room = document.getElementById('room-input').value.trim();

    if (!username || !room) {
        alert('Please enter both username and room name');
        return;
    }

    try {
        // Request audio with specific constraints
        localStream = await navigator.mediaDevices.getUserMedia({
            audio: {
                echoCancellation: true,
                noiseSuppression: true,
                autoGainControl: true
            }
        });

        // Create a local audio element for testing
        const localAudio = new Audio();
        localAudio.id = 'local-audio';
        localAudio.muted = true; // Mute local audio to prevent feedback
        localAudio.srcObject = localStream;
        document.body.appendChild(localAudio);

        document.getElementById('voice-container').style.display = 'block';
        connectWebSocket();
        document.getElementById('login-container').style.display = 'none';
        document.getElementById('chat-container').style.display = 'block';
    } catch (error) {
        alert('Error accessing microphone: ' + error.message);
        console.error('Microphone error:', error);
    }
}


function displayMessage(message) {
    const messagesDiv = document.getElementById('messages');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message';
    messageDiv.textContent = `${message.username}: ${message.content}`;
    messagesDiv.appendChild(messageDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function displaySystemMessage(text) {
    const messagesDiv = document.getElementById('messages');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message';
    messageDiv.style.fontStyle = 'italic';
    messageDiv.textContent = text;
    messagesDiv.appendChild(messageDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function updateUsersList(users = []) {
    const usersContent = document.getElementById('users-content');
    usersContent.innerHTML = '';
    
    users.forEach(user => {
        if (user !== username) {
            const userDiv = document.createElement('div');
            userDiv.className = 'user-item';
            userDiv.innerHTML = `
                ${user}
                <div class="user-status">Connected</div>
            `;
            usersContent.appendChild(userDiv);
        }
    });
}

function endCall() {
    if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
    }

    closeAllConnections();

    document.querySelectorAll('audio').forEach(audio => {
        if (audio.id !== 'local-audio') {
            audio.remove();
        }
    });

    if (ws) {
        ws.close();
    }

    document.getElementById('voice-container').style.display = 'none';
    document.getElementById('chat-container').style.display = 'none';
    document.getElementById('login-container').style.display = 'block';
    
    document.getElementById('username-input').value = '';
    document.getElementById('room-input').value = '';
}

// Event Listeners
document.getElementById('message-input').addEventListener('keypress', function(event) {
    if (event.key === 'Enter') {
        sendMessage();
    }
});