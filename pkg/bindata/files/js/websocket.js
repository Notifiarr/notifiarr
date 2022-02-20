let websockets = [];

function openWebsocket(caller, fileID)
{
    if (!('WebSocket' in window)) {
        toast('Websocket Error', 'Your browser does not support websockets, log tailing not available.', 'error');
        return
    }

    // Set out controller and some other useful variables from data in the template.
    const ctl = caller.closest('.fileController');
    if (!fileID) {
        fileID = ctl.data('currentid');
    }
    if (!fileID) {
        return; // a file has not been selected, and one wasn't provided.
    }

    const source = ctl.data("kind");
    //-- KILL THE SPECIFIED SOCKET
    if (websockets.source) {
        websockets.source.close();
        if (fileID == ctl.data('currentid')) {
            // Tailing the same file needs a pause, or it doesn't work right.
            setTimeout(function() {
                newWebsocket(ctl, source, fileID);
            }, 1000);
            return;
        }
    }

    newWebsocket(ctl, source, fileID);
}

function newWebsocket(ctl, source, fileID)
{
    const url  = ctl.find('#fileinfo-'+ fileID).data('url');
    const box  = ctl.find('.file-content');
    const sort = ctl.find('.fileSortDirection').data('sort');
    let lineCounter = 0;
    let ws = new WebSocket(location.origin.replace(/^http/, 'ws') + url +'?source='+ source +'&fileId='+ fileID);
    websockets.source = ws;

    let sortMsg = '';
    if (sort != "tails") {
        sortMsg = ' (<b>in reverse</b>)';
    }

    // Manipulate meta assets.
    ctl.data('currentid', fileID); // set the current active file id.
    ctl.find('.tailControl').show(); // show the file tail controls.
    ctl.find('.fileinfo-table, .file-control').hide(); // hide non-tail file controls.
    // show the file info div (right side panel) for the file being tailed.
    ctl.find('.file-actions, .fileTablesList, #fileinfo-'+ fileID).show(); // show file actions and file info box.
    $('#'+ source +'TailFile').html(ctl.find('#fileinfo-'+ fileID).data('name')+sortMsg); // set the file name + msg.
    box.html(''); // empty the box out for fresh content.

    //-- HANDLE SOCKET TRANSACTIONS

    ws.onopen = function() {
        toast('Websocket Connected', 'Websocket connection established.', 'success');
    };

    ws.onmessage = function(incoming) {
        lineCounter++;
        let tailLines = $('#'+ source +'TailFileLinesCount').val();
        if (tailLines <= 0) {
            tailLines = 1;
        }

        if (sort == "tails") {
            box.append('<div class="lineContent"><span class="line-number">'+ lineCounter +'</span>'+ incoming.data +'</div>');
            // Truncate box to N lines.
            $.each($('#'+ source +'FileContainer > div > div'), function(counter) {
                if (counter >= tailLines) {
                    $('#'+ source +'FileContainer > div > div:first').remove();
                }
            });
            // Scroll to bottom if auto scroll is enabled.
            if (ctl.find('.tailAutoScroll').prop('checked')) {
                box.scrollTop(box.prop("scrollHeight"));
            }
        } else {
            box.prepend('<div class="lineContent"><span class="line-number">'+ lineCounter +'</span>'+ incoming.data +'</div>');
            // Truncate box to N lines. Not fully tested yet.
            $.each($('#'+ source +'FileContainer > div > div'), function(counter) {
                if (counter >= tailLines) {
                    $('#'+ source +'FileContainer > div > div:last').remove();
                }
            });
            // Scroll to bottom if auto scroll is enabled.
            if (ctl.find('.tailAutoScroll').prop('checked')) {
                box.scrollTop(-1);
            }
        }
    };

    ws.onerror = function(data) {
        toast('Websocket Error', 'Error connecting to the client websocket, details in console.', 'error');
        console.log('Websocket connection error');
        console.log(data);
        ctl.find('.tailControl').hide();
        delete websockets.source;
    };

    ws.onclose = function(data) {
        toast('Websocket Closed', 'The client websocket has been closed, details in console.', 'info');
        console.log('Websocket connection closed: '+ websocketCodes[data.code]);
        console.log(data);
        delete websockets.source;
    };
}

function destroyWebsocket(source)
{
    if (websockets.source) {
        try {
            websockets.source.close();
        } catch {}
    }
    delete websockets.source;
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
