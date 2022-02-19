let websocketCodes = {
    '1000': 'Normal Closure',
    '1001': 'Going Away',
    '1002': 'Protocol Error',
    '1003': 'Unsupported Data',
    '1004': '(For future)',
    '1005': 'No Status Received',
    '1006': 'Abnormal Closure',
    '1007': 'Invalid frame payload data',
    '1008': 'Policy Violation',
    '1009': 'Message too big',
    '1010': 'Missing Extension',
    '1011': 'Internal Error',
    '1012': 'Service Restart',
    '1013': 'Try Again Later',
    '1014': 'Bad Gateway',
    '1015': 'TLS Handshake'
};

function openWebsocket(url, elm) 
{
    if ('WebSocket' in window) {       
        let ws = new WebSocket('ws://localhost:5454'+ url);
        
        ws.onopen = function() {}; //-- USED TO SEND MESSAGE TO THE CLIENT
        
        ws.onmessage = function(incoming) {
            const clientMessage = incoming.data;
            console.log(clientMessage);
        };

        ws.onerror = function(data) { 
            toast('Websocket Error', 'Error connecting to the client websocket, details in console.', 'error');
            console.log('Websocket connection error');
            console.log(data);
        };

        ws.onclose = function(data) {
            toast('Websocket Closed', 'The client websocket has been closed, details in console.', 'info');
            console.log('Websocket connection closed: '+ websocketCodes[data.code]);
            console.log(data);
        };
    } else {
        toast('Websocket Error', 'Your browser does not support websockets, log tailing not available.', 'error');
    }
}
