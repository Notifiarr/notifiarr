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

        // redo tooltips since some got nuked.
        setTooltips();
    },1000);
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
    // This just sets the first few lines of the row. The more-dynamic bits get added below, in a for loop.
    let row = '<tr class="newRow bk-success '+ section +'-'+ app +'" id="'+ section +'-'+ app +'-'+ instance +'">'+
                '   <td style="font-size: 22px;"><span id="'+ app +'IndexLabel'+ (instance-1) +'">'+instance+"</span>"+
                '       <div class="'+ section +'-'+ app +'-deletebutton" style="float: right;">'+
                '           <button onclick="removeInstance(\''+ section +'-'+ app +'\', '+ instance +
                            ')" type="button" title="Delete this instance of '+ app +'.'+
                            '" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
                '       </div>'+
                '   </td>';

    // Timeout and Interval must have valid go durations, or the parser errors.
    // Host or URL are set with a value, and without an original value to make the save button appear.
    let colspan = 1;
    for (const name of names) {
        let nameval = ""
        let nameori = ""
        // Blank items in the name list add a colspan to the next item.
        if (name == "") {
            colspan++;
            continue;
        }

        switch (name) {
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

        row +=  '<td colspan="'+ colspan +'"><input type="text" id="'+ prefix +'.'+ app +'.'+ index +'.'+ name +
        '" data-app="'+ app +'" data-index="'+ index +'" class="client-parameter" data-group="'+ section +'" data-label="'+ app +' '+
         instance +' '+name+'" data-original="'+ nameori +'" value="'+ nameval +'" style="width: 100%;"></td>';
         colspan = 1;
    }

    // Add this new data row to our table.
    $('#'+ section +'-'+ app +'-container').append(row);

    // Redo tooltips since some got added.
    setTooltips();

    // Hide the "no instances" item that displays when no instances are configured.
    $('#'+ section +'-'+ app +'-none').hide();

    // Bring up the save changes button.
    findPendingChanges();
}
