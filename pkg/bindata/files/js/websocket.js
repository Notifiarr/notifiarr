// TODO: how to stop the websocket? when someone clicks a different file, etc.
function openWebsocket(caller, url)
{
    if ('WebSocket' in window) {
        let ws = new WebSocket(location.origin.replace(/^http/, 'ws')  + url);
        const ctl = caller.closest('.fileController');
        const box = ctl.find('.file-content');
        const sort = ctl.find('.fileSortDirection').data('sort');

        ctl.find('.tailControl').show();
        box.html('');

        ws.onopen = function() {}; //-- USED TO SEND MESSAGE TO THE CLIENT

        ws.onmessage = function(incoming) {
            if (sort == "tails") {
                box.append(incoming.data);
                if (ctl.find('.tailAutoScroll').prop('checked')) {
                    box.parent().scrollTop(box.parent().prop("scrollHeight"));
                }
            } else {
                box.prepend(incoming.data);
                if (ctl.find('.tailAutoScroll').prop('checked')) {
                    box.parent().scrollTop(-1);
                }
            }
            // TODO: Truncate the box to an input byte count. Lets hard code to to 20000 for now.
        };

        ws.onerror = function(data) {
            toast('Websocket Error', 'Error connecting to the client websocket, details in console.', 'error');
            console.log('Websocket connection error');
            console.log(data);
            ctl.find('.tailControl').hide();
        };

        ws.onclose = function(data) {
            toast('Websocket Closed', 'The client websocket has been closed, details in console.', 'info');
            console.log('Websocket connection closed: '+ websocketCodes[data.code]);
            console.log(data);
            ctl.find('.tailControl').hide();
        };
    } else {
        toast('Websocket Error', 'Your browser does not support websockets, log tailing not available.', 'error');
    }
}

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
