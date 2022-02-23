// showProcessList displays/shows a hidden page(div) and fills it with the current process list.
function showProcessList()
{
    swapNavigationTemplate('processlist');
    // Start with a spinner because this takes a second or 3.
    $('#process-list-content').html('<h4><i class="fas fa-cog fa-spin"></i> Loading process list...</h4>');

    $.ajax({
        url: 'ps',
        success: function (data){
            const lineCount = data.split(/\n/).length-1; // do not count last newline.
            $('#process-list-msg').html("Displaying "+lineCount+" running processes. Updated: "+ new Date().toLocaleTimeString());
            // Put the data we just downloaded into the content div for the process list.
            $('#process-list-content').text(data);
            // Process List uses line counter. Because why not? They're damn cool.
            updateFileContentCounters();
        },
        error: function (request, status, error) {
            if (error == "") {
                $('#process-list-content').html('<h4>Web Server Error</h4>Notifiarr client appears to be down! Hard refresh recommended.');
            } else {
                $('#process-list-content').html('<h4>'+ error +'</h4>'+ request.responseText);
            }
        },
    });
}

// servicesAction can be used to toggle service checks or initiate service checks.
function servicesAction(action, refresh, refreshDelay = 0) {
    $.ajax({
        url: 'services/'+action,
        success: function (data){
            setTimeout(function() {
                refreshPage(refresh, false);
            }, refreshDelay);
            toast('Yay!', data, 'success')
        },
        error: function (request, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Error', error+': '+request.responseText, 'error', 10000);
            }
        },
    });
}

// addServiceCheck compliments the functions in golists.js.
// This adds new service check inputs to the table.
function addServiceCheck()
{
     //-- DO NOT RELY ON 'index' FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const index = $('.services-Checks').length;
    const instance = index+1;
    const row = '<tr class="bk-success services-Checks" id="services-Checks-'+ index +'">'+
        '<td style="font-size: 22px;"><span '+ (smScreen || mdScreen ? 'style="display:none;" ' : '') +'id="checksIndexLabel'+ index +'" class="mobile-hide tablet-hide">'+ instance +'</span>'+
        '<div class="services-Checks-deletebutton" style="float: right;">'+
        '<button onclick="removeInstance(\'services-Checks\', '+ index +')" type="button" title="Delete this Service Check" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
        '</div>'+
        '</td>'+
        '<td><input type="text" id="Service.'+ index +'.Name" data-app="checks" data-index="'+ index +'" class="client-parameter" data-group="services" data-label="Check '+ instance +' Name" data-original="" value="Service'+ instance +'" style="width: 100%;"></td>'+
        '<td>'+
        '<select id="Service.'+ index +'.Type" data-app="checks" data-index="'+ index +'" class="client-parameter" onChange="checkTypeChange($(this));" data-group="services" data-label="Check '+ instance +' Type" value="http" data-original="" style="width: 100%;">'+
            '<option value="process">Process</option>'+
            '<option value="http">HTTP</option>'+
            '<option value="tcp">TCP Port</option>'+
        '</select>'+
        '</td>'+
        '<td><input type="text" id="Service.'+ index +'.Value" data-app="checks" data-index="'+ index +'" class="client-parameter" data-group="services" data-label="Check '+ instance +' Value" data-original="" style="width: 100%;"></td>'+
        '<td>'+
            '<input type="text" placeholder="200, 302, 404, 500, etc" title="Enter allowed HTTP return codes. Separate with commas." id="Service.'+ index +'.Expect" data-index="'+ index +'" data-app="checks" class="client-parameter serviceHTTPParam" data-group="services" data-label="Check '+ instance +' Expect" data-original="" style="width: 100%;display:none;">'+
            '<input disabled type="text" data-index="'+ index +'" data-app="checks" placeholder="unused" class="client-parameter serviceTCPParam" data-group="services" data-original="" style="width: 100%;display:none;">'+
            '<input type="checkbox" onChange="checkExpectChange($(this));" title="Check this box to send alerts when a matched process restarts." data-index="'+ index +'" data-app="checks" class="serviceProcessParam serviceProcessParamRestart" style="margin-right:4px;">'+
            '<input type="checkbox" onChange="checkExpectChange($(this));" title="Check this box to send alerts when a matched process is found running. Uncommon, and not usable with other options." data-index="'+ index +'" data-app="checks" class="serviceProcessParam serviceProcessParamRunning" style="margin-right:4px;">'+
            '<input type="number" onChange="checkExpectChange($(this));" title="Minimum number of proesses allowed to run." placeholder="Min" data-index="'+ index +'" data-app="checks" class="serviceProcessParam serviceProcessParamMin" value="0" style="margin-right:3px;width:35%;">'+
            '<input type="number" onChange="checkExpectChange($(this));" title="Maximm number of process allowed to run." placeholder="Max" data-index="{{$index}}" data-app="checks" class="serviceProcessParam serviceProcessParamMax" value="0" style="width:35%;">'+
        '</td>'+
        '<td><input type="text" id="Service.'+ index +'.Interval" data-app="checks" data-index="'+ index +'" class="client-parameter" data-group="services" data-label="Check '+ instance +' Interval" data-original="5m" value="5m" style="width: 100%;"></td>'+
        '<td><input type="text" id="Service.'+ index +'.Timeout" data-app="checks" data-index="'+ index +'" class="client-parameter" data-group="services" data-label="Check '+ instance +' Timeout" data-original="1m" value="1m" style="width: 100%;"></td>'+
        '</tr>';

    $('#services-Checks-container').append(row);

    // redo tooltips since some got added.
    setTooltips();

    // hide the "no instances" item that displays if no instances are configured.
    $('#services-Checks-none').hide();

    // bring up the save changes button.
    findPendingChanges();
}

// This fires when a check type is changed. It updates the "expect" for inputs
// to appropriate and easy to use values for each type.
function checkTypeChange(from)
{
    const ctl = from.closest('.services-Checks'); // just this row.
    ctl.find('.serviceProcessParam').hide();
    ctl.find('.serviceHTTPParam').hide();
    ctl.find('.serviceTCPParam').hide();

    switch (from.val()) {
    case "process":
        ctl.find('.serviceProcessParam').show();
        break;
    case "http":
        ctl.find('.serviceHTTPParam').show();
        break;
    case "tcp":
        ctl.find('.serviceTCPParam').show();
        break;
    }
}

// This fires when an expect value (for process type) is changed.
// Some values are incompatible with others, and the whole thing is copied
// into the "real" value, so it can be POSTed properly (it's a string).
function checkExpectChange(from)
{
    const ctl = from.closest('.services-Checks'); // just this row.
    const min = ctl.find('.serviceProcessParamMin').val();
    const max = ctl.find('.serviceProcessParamMax').val();
    const res = ctl.find('.serviceProcessParamRestart').prop('checked');
    const run = ctl.find('.serviceProcessParamRunning').prop('checked');

    // The "running" checkbox does not allow any other arguments, so deal with that here.
    if (run) {
        // Copy "running" into real value that is POSTed.
        ctl.find('.serviceHTTPParam').val('running');
        // Disable incompatible options.
        ctl.find('.serviceProcessParamMin').prop('disabled', true);
        ctl.find('.serviceProcessParamMax').prop('disabled', true);
        ctl.find('.serviceProcessParamRestart').prop('disabled', true);
    } else {
        // Copy the concatenated string (from three sources) into the real value.
        ctl.find('.serviceHTTPParam').val('count:'+ min +':'+ max + (res ? ',restart' : ''));
        // Enable all values.
        ctl.find('.serviceProcessParamMin').prop('disabled', false);
        ctl.find('.serviceProcessParamMax').prop('disabled', false);
        ctl.find('.serviceProcessParamRestart').prop('disabled', false);
    }
}
