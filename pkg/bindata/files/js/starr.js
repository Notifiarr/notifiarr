function removeStarrInstance(app, index)
{
    $('#starr-'+ app +'-'+ index).remove();
    setTooltips();
    if (!$('.starr-'+ app).length) {
        $('#starr-'+ app +'-none').show();
    }
}
// -------------------------------------------------------------------------------------------

function addStarrInstance(app)
{
    const titleCase = titleCaseWord(app);
    const index = $('.starr-'+ app).length + 1; //-- DO NOT RELY ON THIS FOR DIRECT IMPLEMENTATION, USE IT TO SORT AND RE-INDEX DURING THE SAVE
    const row = '<tr class="starr-'+ app +'" id="starr-'+ app +'-'+ index +'">'+
                '   <td style="font-size: 22px;">*'+
                '       <div style="float: right;">'+
                '           <button onclick="removeStarrInstance(\''+ app +'\', '+ index +')" type="button" title="Delete this instance of '+ titleCase +'" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
                '       </div>'+
                '   </td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.Name['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * Name" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.URL['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * URL" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.APIKey['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * API Key" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.Username['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * Username" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.Password['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * Password" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.Interval['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * Interval" data-original="" value="" style="width: 100%;"></td>'+
                '   <td><input type="text" id="Apps.'+ titleCase +'.Timeout['+ index +']" class="client-parameter" data-group="starr" data-label="'+ titleCase +' * Timeout" data-original="" value="" style="width: 100%;"></td>'+
                '</tr>';

    $('#starr-'+ app +'-container').append(row);
    $('#starr-'+ app +'-none').hide();
}
// -------------------------------------------------------------------------------------------
