// triggerAction is called from a submenu in the nav bar.
// triggers pretty simple. This only starts an action and is almost always successful.
function triggerAction(action)
{
    $.ajax({
        type: 'GET',
        url: URLBase+'trigger/'+action,
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

function testRegex()
{
    let fields = '';
    $.each($('.regex-parameter'), function() {
        if ($(this).attr('id') !== undefined) {
            fields += '&' + $(this).serialize();
        }
    });

    $.ajax({
        type: 'POST',
        data: fields,
        url: URLBase+'regexTest',
        success: function (data){
            toast('Matched!', data, 'success');
        },
        error: function (response, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 20000);
            } else {
                toast(error, response.responseText, 'error', 15000);
            }
        }
    });
}

// getCmdStats returns the cmdstats template for a specific command.
function getCmdStats(caller, index)
{
    toast('Working', 'Getting stats!', 'success', 2200);
    $.ajax({
        url: URLBase+'cmdstats/'+index,
        success: function (data){
            $('#commandStats'+index).html(data);
            dialog(caller, 'left');
        },
        error: function (response, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 20000);
            } else {
                toast(error, response.responseText, 'error', 15000);
            }
        }
    });
}