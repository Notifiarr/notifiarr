// Remove an instance from a list of applications.
// Example: Click delete button on one of the Radarrs.
// Works for Snapshot and Download Clients too.
function removeInstance(name, index)
{
    $('#'+ name +'-'+ index).remove();
    // redo tooltips since some got nuked.
    setTooltips();
    // if all instances are deleted, show the "no instances" item.
    if (!$('.'+ name).length) {
        $('#'+ name +'-none').show();
    }
    // mark this instance as deleted (to bring up save changes button).
    $('.'+ name + index +'-deleted').val(true);

    //-- FIX THE INDEXING
    reindexList(name);

    // bring up the save changes button.
    findPendingChanges();
}

// addInstance is used to build the table rows.
function addInstance(section, app)
{
     //-- DO NOT RELY ON 'index' FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const index = $('.'+ section +'-'+ app).length;
    const instance = index+1;
    const prefix = $('#'+ section +'-'+ app +'-addbutton').data('prefix');  // storage location like Apps or Snapshot.
    const names = $('#'+ section +'-'+ app +'-addbutton').data('names'); // list of thinges like "Name" and "URL"
    let row = '<tr class="'+ section +'-'+ app +'" id="'+ section +'-'+ app +'-'+ instance +'">'+
                '   <td style="font-size: 22px;"><span id="'+ lowercaseWord(app) +'-index-label-'+ (instance-1) +'">'+instance+"</span>"+
                '       <div class="'+ section +'-'+ app +'-deletebutton" style="float: right;">'+
                '           <button onclick="removeInstance(\''+ section +'-'+ app +'\', '+ instance +
                            ')" type="button" title="Delete this instance of '+ app +'.'+
                            '" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
                '       </div>'+
                '   </td>';

    // Timeout and Interval must have valid go durations, or the parser errors.
    // Host or URL are set without an original value to make the save button appear.
    let colspan = 1
    for (const name of names) {
        let nameval = ""
        let nameori = ""
        if (name == "") {
            colspan++
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
        '" class="client-parameter" data-group="'+ section +'" data-label="'+ app +' '+
         instance +' '+name+'" data-original="'+ nameori +'" value="'+ nameval +'" style="width: 100%;"></td>';
         colspan = 1
    }


    $('#'+ section +'-'+ app +'-container').append(row);

    // redo tooltips since some got added.
    setTooltips();

    // hide the "no instances" item that displays if no instances are configured.
    $('#'+ section +'-'+ app +'-none').hide();

    // bring up the save changes button.
    findPendingChanges();
}
