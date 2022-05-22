// Remove an instance from a list of applications.
// Example: Click delete button on one of the Radarrs.
// Works for Snapshot and Download Clients too.
function removeInstance(name, index)
{
    // service table has to be handled specially.
    serviceTable.row($('#'+ name +'-'+ index)).remove();
    $('#'+ name +'-'+ index).fadeOut(1000);
    setTimeout(function() {
        serviceTable.draw();
        $('#'+ name +'-'+ index).remove();
        // if all instances are deleted, show the "no instances" item.
        if (!$('.'+ name).length) {
            $('#'+ name +'-none').show();
        }
        // mark this instance as deleted (to bring up save changes button).
        $('.'+ name + index +'-deleted').val(true);

        //-- FIX THE INDEXING
        reindexList(name.split('-')[0]);

        // bring up the save changes button.
        findPendingChanges();
    },1000);
}

function testInstance(from, instanceType, index)
{
    from.css({'pointer-events':'none'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
    let fields = '';
    from.closest('tr').find('.client-parameter').each(function() {
        const id = $(this).attr('id')
        if (id !== undefined) {
            fields += '&' + $(this).serialize();
        }
    });

    $.ajax({
        type: 'POST',
        url: URLBase+'checkInstance/'+instanceType+"/"+index,
        data: fields,
        success: function (data){
            from.css({'pointer-events':'auto'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
            toast(instanceType+' Check Successful', data, 'success');
        },
        error: function (response, status, error) {
            from.css({'pointer-events':'auto'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
            if (status === undefined) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast(instanceType+' Check Error', error+': '+response.responseText, 'error', 15000);
            }
        }
    });
}

// The Go app only accepts indexes on lists starting at 0 with no gaps.
// It looks better on screen to see incremental numbers.
// This procedure fixes the numbering on each row when an item is deleted.
function reindexList(group)
{
    let counter = 0;
    let currentIndex = 0;
    let currentApp = '';

    // Loop each form element with this data group and with a non-empty id.
    $('[data-group='+ group+'][id]').each(function(index) {
        const rowIndex = $(this).data("index");
        const itemApp = $(this).data("app");

        if (itemApp !== currentApp) {
            // We get into here when the app (in the group) changes, and on the very first element.
            counter = 0;
            currentIndex = 0;
        } else if (rowIndex !== currentIndex) {
            // We get into here on any new element in the same app.
            counter++;
        }

        //-- CHANGE id ATTR ON LABEL
        $('#'+ itemApp +'IndexLabel'+ rowIndex).attr('id', itemApp +'IndexLabel'+ counter).html(counter + 1);

        //-- CHANGE data-label ON INPUT FIELDS
        $(this).data('label', $(this).data('label').replace((rowIndex + 1), (counter + 1)));

        //-- CHANGE id ATTR ON INPUT FIELDS
        $(this).attr('id', $(this).attr('id').replace(rowIndex, counter));

        //-- SET CURRENT VARIABLES
        $(this).data("index", counter)
        currentIndex = rowIndex;
        currentApp = itemApp;
    });
}

// addInstance is used to build the table rows.
function addInstance(section, app)
{
     //-- DO NOT RELY ON 'index' FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const index = $('.'+ section +'-'+ app).length;
    const instance = index+1;
    const prefix = $('#'+ section +'-'+ app +'-addbutton').data('prefix');  // storage location like Apps or Snapshot.
    const names = $('#'+ section +'-'+ app +'-addbutton').data('names'); // list of thinges like "Name" and "URL"
    // This just sets the first few lines of the row (the action buttons). The more-dynamic bits get added below, in a for loop.
    let row = '<tr class="newRow bk-success '+ section +'-'+ app +'" id="'+ section +'-'+ app +'-'+ instance +'">'+
    '<td style="white-space:nowrap;">'+
        '<div class="btn-group" role="group" style="display:flex;">'+
            '<button onclick="removeInstance(\''+ section +'-'+ app +'\', '+ instance +')" type="button" class="delete-item-button btn btn-danger btn-sm" style="font-size:18px;width:35px;"><i class="fa fa-minus"></i></button>'+
            '<button id="'+ app +'IndexLabel'+ index +'" class="btn btn-sm" style="font-size:18px;width:35px;pointer-events:none;">'+ instance +'</button>'+
            '<button onClick="testInstance($(this), \''+app+'\', \''+index+'\')" type="button" style="font-size:18px;" class="btn btn-success btn-sm"><i class="fas fa-check-double"></i></button>'+
        '</div>'+
    '</td>';

    // Timeout and Interval must have valid go durations, or the parser errors.
    // Host or URL are set with a value, and without an original value to make the save button appear.
    let colspan = 1;
    for (const name of names) {
        let nameval = ""
        let nameori = ""
        let extra = ""
        let itype = "text"

        // Blank items in the name list add a colspan to the next item.
        if (name == "") {
            colspan++;
            continue;
        }

        switch (name) {
            case "Config.Password":
            case "Config.Pass":
            case "Config.APIKey":
            case "Pass":
            case "Password":
            case "APIKey":
                extra = '<div style="width:35px; max-width:35px;" class="input-group-addon input-sm" onClick="togglePassword(\''+ prefix +'.'+ app +'.'+ index +'.'+ name + '\', $(this).find(\'i\'));"><i class="fas fa-low-vision secret-input"></i></div>';
                itype = "password";
                break;
            case "Config.Interval":
            case "Interval":
                nameval = "5m";
                nameori = "5m";
             break;
             case "Config.Timeout":
             case "Timeout":
                nameval = "1m";
                nameori = "1m";
             break;
             case "Config.URL":
             case "URL":
             case "Host":
                nameval = "changeme";
                nameori = "";
             break;
        }

        row += '<td colspan="'+ colspan +'"><form class="form-inline"><div class="form-group" style="width:100%"><div class="input-group" style="width:100%">';
        row += '<input type="'+ itype +'" name="'+ prefix +'.'+ app +'.'+ index +'.'+ name +'" '+
        'id="'+ prefix +'.'+ app +'.'+ index +'.'+ name +'" '+
        'name="'+ prefix +'.'+ app +'.'+ index +'.'+ name +'" '+
        'data-app="'+ app +'" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="'+ section +'" data-label="'+ app +' '+
         instance +' '+name+'" data-original="'+ nameori +'" value="'+ nameval +'">'+ extra;
        row += '</div></div></form></td>';
        colspan = 1;
    }

    // Add this new data row to our table.
    $('#'+ section +'-'+ app +'-container').append(row);

    // Hide the "no instances" item that displays when no instances are configured.
    $('#'+ section +'-'+ app +'-none').hide();

    // Bring up the save changes button.
    findPendingChanges();
}

function addWatchFiles()
{
    const index = $('.files-WatchFiles').length;
    const instance = index+1;
    const names = $('#files-WatchFiles-addbutton').data('names'); // list of thinges like "Name" and "URL"
    // This just sets the first few lines of the row (the action buttons). The more-dynamic bits get added below, in a for loop.
    let row = '<tr class="newRow bk-success files-WatchFiles" id="files-WatchFiles-'+ instance +'">'+
    '<td style="white-space:nowrap;">'+
        '<div class="btn-group" role="group" style="display:flex;">'+
            '<button onclick="removeInstance(\'files-WatchFiles\', '+ instance +')" type="button" class="delete-item-button btn btn-danger btn-sm" style="font-size:18px;width:35px;"><i class="fa fa-minus"></i></button>'+
            '<button id="filesIndexLabel'+ index +'" class="btn btn-sm" style="font-size:18px;width:35px;pointer-events:none;">'+ instance +'</button>'+
        '</div>'+
    '</td>';

    // Timeout and Interval must have valid go durations, or the parser errors.
    // Host or URL are set with a value, and without an original value to make the save button appear.
    for (const name of names) {
        let nameval = ""
        if (name == "Path") {
            nameval = "/some/path";
        }

        row += '<td><form class="form-inline"><div class="form-group" style="width:100%"><div class="input-group" style="width:100%">';

        switch (name) {
             case "Poll":
             case "Pipe":
             case "MustExist":
                 row += '<select id="WatchFiles.'+ index +'.'+ name +'" name="WatchFiles.'+ index +'.'+ name +'" data-index="'+ index +'" data-app="WatchFiles" '+
                    'class="client-parameter form-control input-sm" data-group="files" data-label="Files '+ instance +' '+ name +'" data-original="false" value="false">'+
                    '<option value="true">Enabled</option>'+
                    '<option selected value="false">Disabled</option></select>';
                break;
             default:
                row += '<input type="text" name="WatchFiles.'+ index +'.'+ name +'" '+
                    'id="WatchFiles.'+ index +'.'+ name +'" '+
                    'name="WatchFiles.'+ index +'.'+ name +'" '+
                    'data-app="WatchFiles" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="files" data-label="Files '+
                    instance +' '+name+'" data-original="" value="'+ nameval +'">';
                break;
        }

        row += '</div></div></form></td>';
    }

    // Add this new data row to our table.
    $('#files-WatchFiles-container').append(row);

    // Hide the "no instances" item that displays when no instances are configured.
    $('#files-WatchFiles-none').hide();

    // Bring up the save changes button.
    findPendingChanges();
}
