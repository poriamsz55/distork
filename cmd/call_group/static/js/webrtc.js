const configuration = {
    iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
    ]
};

let localStream;
let peerConnections = {};

async function createPeerConnection(targetUser, initiator = false) {
    if (peerConnections[targetUser]) {
        console.log(`Peer connection already exists for ${targetUser}`);
        return peerConnections[targetUser];
    }

    console.log(`Creating new peer connection for ${targetUser} (initiator: ${initiator})`);
    const peerConnection = new RTCPeerConnection(configuration);
    peerConnections[targetUser] = peerConnection;

    // Add local stream
    if (localStream) {
        console.log('Adding local stream tracks to peer connection');
        localStream.getTracks().forEach(track => {
            peerConnection.addTrack(track, localStream);
        });
    }

    // Handle ICE candidates
    peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
            console.log('Sending ICE candidate');
            sendSignal(targetUser, {
                type: 'candidate',
                candidate: event.candidate
            });
        }
    };

    // Log state changes
    peerConnection.oniceconnectionstatechange = () => {
        console.log(`ICE Connection State: ${peerConnection.iceConnectionState}`);
    };

    peerConnection.onsignalingstatechange = () => {
        console.log(`Signaling State: ${peerConnection.signalingState}`);
    };

    peerConnection.onconnectionstatechange = () => {
        console.log(`Connection State: ${peerConnection.connectionState}`);
    };

    // Handle incoming stream
    peerConnection.ontrack = (event) => {
        console.log(`Received track from ${targetUser}`);
        const stream = event.streams[0];
        
        // Remove any existing audio element for this user
        const existingAudio = document.getElementById(`audio-${targetUser}`);
        if (existingAudio) {
            existingAudio.remove();
        }

        // Create new audio element
        const audioElement = new Audio();
        audioElement.id = `audio-${targetUser}`;
        audioElement.srcObject = stream;
        audioElement.autoplay = true;
        audioElement.controls = true; // Add controls for debugging
        document.body.appendChild(audioElement);
        
        // Ensure audio plays
        audioElement.play().catch(error => {
            console.error('Error playing audio:', error);
        });
    };

    if (initiator) {
        try {
            console.log('Creating and sending offer');
            const offer = await peerConnection.createOffer({
                offerToReceiveAudio: true
            });
            await peerConnection.setLocalDescription(offer);
            sendSignal(targetUser, {
                type: 'offer',
                offer: offer
            });
        } catch (error) {
            console.error('Error creating offer:', error);
        }
    }

    return peerConnection;
}

async function handleSignalingMessage(message) {
    const signal = JSON.parse(message.signal);
    const sender = message.username;

    console.log(`Handling signal: ${signal.type} from ${sender}`);

    try {
        let peerConnection = peerConnections[sender];

        switch (signal.type) {
            case 'offer':
                if (!peerConnection) {
                    peerConnection = await createPeerConnection(sender, false);
                }
                
                if (peerConnection.signalingState !== "stable") {
                    console.log('Ignoring offer in non-stable state');
                    return;
                }

                await peerConnection.setRemoteDescription(new RTCSessionDescription(signal.offer));
                const answer = await peerConnection.createAnswer();
                await peerConnection.setLocalDescription(answer);
                
                sendSignal(sender, {
                    type: 'answer',
                    answer: answer
                });
                break;

            case 'answer':
                if (peerConnection && peerConnection.signalingState !== "stable") {
                    await peerConnection.setRemoteDescription(new RTCSessionDescription(signal.answer));
                }
                break;

            case 'candidate':
                if (peerConnection) {
                    try {
                        await peerConnection.addIceCandidate(new RTCIceCandidate(signal.candidate));
                    } catch (e) {
                        console.error('Error adding received ice candidate:', e);
                    }
                }
                break;
        }
    } catch (error) {
        console.error('Error handling signal:', error);
    }
}

function closeConnection(userId) {
    const peerConnection = peerConnections[userId];
    if (peerConnection) {
        peerConnection.close();
        delete peerConnections[userId];
    }

    const audioElement = document.getElementById(`audio-${userId}`);
    if (audioElement) {
        audioElement.remove();
    }
}

function closeAllConnections() {
    Object.keys(peerConnections).forEach(userId => {
        closeConnection(userId);
    });
}