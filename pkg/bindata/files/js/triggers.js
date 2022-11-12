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
            if (response.status == 0) {
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
        cache: false,
        url: URLBase+'regexTest',
        success: function (data){
            toast('Matched!', data, 'success');
        },
        error: function (response, status, error) {
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 20000);
            } else {
                toast(error!=''?error:'Bad Request', response.responseText, 'error', 15000);
            }
        }
    });
}

// getCmdStats returns the cmdstats template for a specific command.
function getCmdStats(caller, hash)
{
    toast('Working', 'Getting stats!', 'success', 2200);
    $.ajax({
        url: URLBase+'ajax/cmdstats/'+hash,
        success: function (data){
            $('#commandStats'+hash).html(data);
            dialog(caller, 'left');
        },
        error: function (response, status, error) {
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 20000);
            } else {
                toast(error!=''?error:'Bad Request', response.responseText, 'error', 15000);
            }
        }
    });
}

function getCmdArgs(from, hash)
{
    $.ajax({
        url: URLBase+'ajax/cmdargs/'+hash,
        success: function (data){
            $('#commandArgs'+hash).html(data);
            dialog(from, 'left');
        },
        error: function (response, status, error) {
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 20000);
            } else {
                toast(error!=''?error:'Bad Request', response.responseText, 'error', 15000);
            }
        }
    });
}

function runCommand(from, hash)
{
    let fields = '';
    from.closest('table').find('#args').each(function() {
        fields += '&' + $(this).serialize();
    });

    $.ajax({
        type: 'POST',
        url: URLBase+'runCommand/'+hash,
        data: fields,
        success: function (data){
            toast('Command Executed', data, 'success');
        },
        error: function (response, status, error) {
            if (response.responseText === undefined) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Command Execution Error', error+': '+response.responseText, 'error', 15000);
            }
        }
    });
    from.parents('.ui-dialog').find('.ui-dialog-content').dialog('close');
    $('#commandArgs'+hash).html('');
}
