
function pingTunnels() {
    $("#tunnel-ping-spinner").show();
    $.ajax({
        type: 'GET',
        url: URLBase+'tunnel/ping',
        success: function (data){
            toast('Tunnel Ping', "Find the response times in the primary tunnel list.", 'success');
            $("#tunnel-ping-spinner").hide();
            const obj = JSON.parse(data)
            for(var idx in obj) {
                // loop data and update each tunnel with the ping time.
                if (obj.hasOwnProperty(idx)){
                    $('#tunnel-ping'+idx).html(obj[idx]);
                }
            }
        },
        error: function (response, status, error) {
            $("#tunnel-ping-spinner").hide();
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Ping Error', error+': '+response.responseText, 'error', 10000);
            }
        }
    });
}

function saveTunnels() {
    $("#tunnel-save-spinner").show();
    $.ajax({
        type: 'POST',
        url: URLBase+'tunnel/save',
        data: $(".tunnel-param").serialize(),
        success: function (data){
            toast('Tunnels Saved', data, 'success');
            $("#tunnel-save-spinner").hide();
            refreshPage('tunnel', false);
        },
        error: function (response, status, error) {
            $("#tunnel-save-spinner").hide();
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Tunnel Save Error', error+': '+response.responseText, 'error', 10000);
            }
        }
    });
}
