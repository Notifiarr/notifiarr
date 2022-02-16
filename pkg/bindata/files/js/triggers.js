// triggerAction is called from a submenu in the nav bar.
// triggers pretty simple. This only starts an action and is almost always successful.
function triggerAction(action) {
    $.ajax({
        type: 'GET',
        url: 'trigger/'+action,
        success: function (data){
            toast('Trigger Sent', data, 'success');
        },
        error: function (response, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Trigger Error', error+': '+response.responseText, 'error', 10000);
            }
        }
    });
}
