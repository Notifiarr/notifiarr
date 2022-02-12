function removeStarrInstance(app, index)
{
    $('#starr-'+ app +'-'+ index).remove();
    // redo tooltips since some got nuked.
    setTooltips();
    // if all instances are deleted, show the "no instances" item.
    if (!$('.starr-'+ app).length) {
        $('#starr-'+ app +'-none').show();
    }
    // mark this instance as deleted (to bring up save changes button).
    $('.' + app + index + '-deleted').val(true);
    // bring up the delete button on the next last item.
    $('.delete-'+app+'-button').last().show();
    // bring up the save changes button.
    findPendingChanges();
}
// -------------------------------------------------------------------------------------------

function addStarrInstance(app)
{
     //-- DO NOT RELY ON 'index' FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const index = $('.starr-'+ app).length;
    const instance = index+1;
    const titleCase = titleCaseWord(app);
    const row = '<tr class="starr-'+ app +'" id="starr-'+ app +'-'+ instance +'">'+
                '   <td style="font-size: 22px;">'+instance+
                '       <div class="delete-'+ app +'-button" style="float: right; display:none;">'+
                '           <button onclick="removeStarrInstance(\''+ app +'\', '+ instance +')" type="button" title="Delete this instance of '+ titleCase +'" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
                '       </div>'+
                '   </td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.Name" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' Name" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.URL" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' URL" data-original="" value="http://changeme" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.APIKey" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' API Key" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.Username" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' Username" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.Password" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' Password" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.Interval" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' Interval" data-original="5m" value="5m" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.'+ index +'.Timeout" class="client-parameter" data-group="starr" data-label="'+ titleCase +' '+ instance +' Timeout" data-original="1m" value="1m" style="width: 100%;"></td>'+
                '</tr>';
    $('#starr-'+ app +'-container').append(row);
    // hide all delete buttons, and show only the one we just added.
    $('.delete-'+app+'-button').hide().last().show();
    // redo tooltips since some got added.
    setTooltips();
    // hide the "no instances" item that displays if no instances are configured.
    $('#starr-'+ app +'-none').hide();
    // bring up the save changes button.
    findPendingChanges();
}
// -------------------------------------------------------------------------------------------
