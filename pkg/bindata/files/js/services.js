// showProcessList displays/shows a hidden page(div) and fills it with the current process list.
function showProcessList() {
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


// addServiceCheck compliments the functions in golists.js.
// This adds new service check inputs to the table.
function addServiceCheck()
{
     //-- DO NOT RELY ON 'index' FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const index = $('.services-Checks').length;
    const instance = index+1;
    const row = '<tr class="services-Checks" id="services-Checks-'+ index +'">'+
        '<td style="font-size: 22px;">'+ instance +
        '<div class="services-Checks-deletebutton" style="float: right;">'+
        '<button onclick="removeInstance(\'services-Checks\', '+ index +')" type="button" title="Delete this Service Check" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
        '</div>'+
        '</td>'+
        '<td><input type="text" id="Service.'+ index +'.Name" class="client-parameter" data-group="services" data-label="Check '+ instance +' Name" data-original="" value="Service'+ instance +'" style="width: 100%;"></td>'+
        '<td>'+
        '<select id="Service.'+ index +'.Type" class="client-parameter" data-group="services" data-label="Check '+ instance +' Type" value="http" data-original="" style="width: 100%;">'+
            '<option value="process">Process</option>'+
            '<option value="http">HTTP</option>'+
            '<option value="tcp">TCP Port</option>'+
        '</select>'+
        '</td>'+
        '<td><input type="text" id="Service.'+ index +'.Value" class="client-parameter" data-group="services" data-label="Check '+ instance +' Value" data-original="" style="width: 100%;"></td>'+
        '<td><input type="text" id="Service.'+ index +'.Expect" class="client-parameter" data-group="services" data-label="Check '+ instance +' Expect" data-original="" style="width: 100%;"></td>'+
        '<td><input type="text" id="Service.'+ index +'.Interval" class="client-parameter" data-group="services" data-label="Check '+ instance +' Interval" data-original="5m" value="5m" style="width: 100%;"></td>'+
        '<td><input type="text" id="Service.'+ index +'.Timeout" class="client-parameter" data-group="services" data-label="Check '+ instance +' Timeout" data-original="1m" value="1m" style="width: 100%;"></td>'+
        '</tr>';

    // Add the row we just created to the end of the table.
    $('#services-Checks-container').append(row);
    // Hide all delete buttons, and show only the one we just added.
    $('.services-Checks-deletebutton').hide().last().show();
    // Redo tooltips since some got added.
    setTooltips();
    // Hide the "no instances" item that displays if no instances are configured.
    $('#services-Checks-none').hide();
    // Bring up the save changes button.
    findPendingChanges();
}
