let websockets = [];

function openWebsocket(caller, url, source, fileID, socket, fileName)
{
    //-- KILL THE SPECIFIED SOCKET
    if (websockets.socket) {
        websockets.socket.close();
    }

    if ('WebSocket' in window) {
        $('#tailFile').html(fileName);
        let tailLines = $('#tailFileLinesCount').val();

        if (tailLines <= 0) {
            tailLines = 1;
        }

        let ws = new WebSocket(location.origin.replace(/^http/, 'ws') + url +'?source='+ source +'&fileId='+ fileID);
        websockets.socket = ws;

        const ctl = caller.closest('.fileController');
        const box = ctl.find('.file-content');
        const sort = ctl.find('.fileSortDirection').data('sort');
        let lineCounter = 0;

        ctl.find('.tailControl').show();
        // show the file info div (right side panel) for the file being tailed.
        ctl.find('.fileinfo-table, .file-control').hide()
        ctl.find('.file-actions, #fileinfo-'+ fileID).show()
        box.html('');

        ws.onopen = function() {
            toast('Websocket Connected', 'Websocket connection established.', 'success');
        };

        ws.onmessage = function(incoming) {
            lineCounter++;

            if (sort == "tails") {
                box.append('<div class="lineContent"><span class="line-number">'+ lineCounter +'</span>'+ incoming.data +'</div>');
                if (ctl.find('.tailAutoScroll').prop('checked')) {
                    box.parent().scrollTop(box.parent().prop("scrollHeight"));
                }

                // Truncate box to 100 lines. Not fully tested yet.
                $.each($('#logFileContainer > pre > div'), function(counter) {
                    if (counter == 99) {
                        $('#logFileContainer > pre > div:first').remove();
                    }
                });
            } else {
                box.append('<div class="lineContent"><span class="line-number">'+ lineCounter +'</span>'+ incoming.data +'</div>');
                box.prepend(incoming.data);
                if (ctl.find('.tailAutoScroll').prop('checked')) {
                    box.parent().scrollTop(-1);
                }

                // Truncate box to 100 lines. Not fully tested yet.
                $.each($('#logFileContainer > pre > div'), function(counter) {
                    if (counter == 99) {
                        $('#logFileContainer > pre > div').remove();
                    }
                });
            }
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
