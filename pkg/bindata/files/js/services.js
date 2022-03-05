$.fn.dataTableExt.ofnSearch['html-input'] = function(value) {
    if ($(value).find('input').length != 0) {
        return $(value).find('input').val();
    }
    if ($(value).find('select').length != 0) {
        return $(value).find('select').val();
    }
    return "";
};

$.fn.dataTableExt.ofnSearch['span-input'] = function(value) {
    return $(value).text();
};

serviceTable = $('.servicetable').DataTable({
    "autoWidth": false,
    "fixedHeader": {
        headerOffset: 50
    },
    "scrollX": true,
    'scrollCollapse': true,
    "sort": false,
    "responsive": true,
    'scrollY': '79vh',
    "paging": false,
    "oLanguage": {
        "sInfo": "Showing _START_ to _END_ of _TOTAL_ service checks.",
        "sZeroRecords": "No matching service chcks found.",
        "sInfoEmpty": "Showing 0 to 0 of 0 service checks.",
        "sInfoFiltered": "(filtered from _MAX_ total service checks)"
    },
    "columnDefs": [
        { "type": "span-input", "targets": [0] },
        { "type": "html-input", "targets": [1,2,3,4,5,6] },
    ],
    "fnDrawCallback": function() {
        this.api().columns.adjust();
    },
    "columns": [
        { "width": "30px" },
        null,
        null,
        null,
        null,
        { "width": "90px" },
        { "width": "70px" }
    ]
});

// showProcessList displays/shows a hidden page(div) and fills it with the current process list.
function showProcessList()
{
    swapNavigationTemplate('processlist');
    // Start with a spinner because this takes a second or 3.
    $('#process-list-content').html('<h4><i class="fas fa-cog fa-spin"></i> Loading process list...</h4>');

    $.ajax({
        url: URLBase+'ps',
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
        url: URLBase+'services/'+action,
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
    const index = serviceTable.rows().count();
    const instance = index+1;
    const row = '<tr class="bk-success services-Checks newRow" id="services-Checks-'+ index +'">'+
    '<td class="text-center" style="font-size: 22px;white-space: nowrap;"><span '+ (smScreen || mdScreen ? 'style="display:none;" ' : '') +'id="checksIndexLabel'+ index +'" class="mobile-hide tablet-hide">'+ instance +'&nbsp;</span>'+
        '<span class="services-Checks-deletebutton">'+
            '<button onclick="removeInstance(\'services-Checks\', '+ index +')" type="button" title="Delete this Service Check" class="delete-item-button btn btn-danger btn-sm"><i class="fa fa-minus"></i></button>'+
        '</span>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Name" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Name" data-original="" value="Service'+ instance +'">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<select id="Service.'+ index +'.Type" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm serviceTypeSelect" onChange="checkTypeChange($(this));" data-group="services" data-label="Check '+ instance +' Type" value="http" data-original="">'+
                    '<option value="process">Process</option>'+
                    '<option value="http" selected>HTTP</option>'+
                    '<option value="tcp">TCP Port</option>'+
                '</select>'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Value" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Value" data-original="">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input id="Service.'+ index +'.Expect" name="Service.'+ index +'.Expect" data-index="'+ index +'" data-app="checks" class="client-parameter serviceProcessParamExpect" data-group="services" data-label="Check '+ instance +' Expect" value="200" data-original="200" style="display:none;">'+
                '<select id="Service.'+ index +'.Expect.StatusCode" onChange="checkExpectChange($(this));" value="200" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceHTTPParam">'+
                    '<option value="100">100: Continue</option>'+
                    '<option value="101">101: SwitchingProtocols</option>'+
                    '<option value="102">102: Processing</option>'+
                    '<option value="103">103: EarlyHints</option>'+
                    '<option value="200" selected>200: OK</option>'+
                    '<option value="201">201: Created</option>'+
                    '<option value="202">202: Accepted</option>'+
                    '<option value="203">203: NonAuthoritativeInfo</option>'+
                    '<option value="204">204: NoContent</option>'+
                    '<option value="205">205: ResetContent</option>'+
                    '<option value="206">206: PartialContent</option>'+
                    '<option value="207">207: MultiStatus</option>'+
                    '<option value="208">208: AlreadyReported</option>'+
                    '<option value="226">226: IMUsed</option>'+
                    '<option value="300">300: MultipleChoices</option>'+
                    '<option value="301">301: MovedPermanently</option>'+
                    '<option value="302">302: Found</option>'+
                    '<option value="303">303: SeeOther</option>'+
                    '<option value="304">304: NotModified</option>'+
                    '<option value="305">305: UseProxy</option>'+
                    '<option value="307">307: TemporaryRedirect</option>'+
                    '<option value="308">308: PermanentRedirect</option>'+
                    '<option value="400">400: BadRequest</option>'+
                    '<option value="401">401: Unauthorized</option>'+
                    '<option value="402">402: PaymentRequired</option>'+
                    '<option value="403">403: Forbidden</option>'+
                    '<option value="404">404: NotFound</option>'+
                    '<option value="405">405: MethodNotAllowed</option>'+
                    '<option value="406">406: NotAcceptable</option>'+
                    '<option value="407">407: ProxyAuthRequired</option>'+
                    '<option value="408">408: RequestTimeout</option>'+
                    '<option value="409">409: Conflict</option>'+
                    '<option value="410">410: Gone</option>'+
                    '<option value="411">411: LengthRequired</option>'+
                    '<option value="412">412: PreconditionFailed</option>'+
                    '<option value="413">413: RequestEntityTooLarge</option>'+
                    '<option value="414">414: RequestURITooLong</option>'+
                    '<option value="415">415: UnsupportedMediaType</option>'+
                    '<option value="416">416: RequestedRangeNotSatisfiable</option>'+
                    '<option value="417">417: ExpectationFailed</option>'+
                    '<option value="418">418: Teapot</option>'+
                    '<option value="421">421: MisdirectedRequest</option>'+
                    '<option value="422">422: UnprocessableEntity</option>'+
                    '<option value="423">423: Locked</option>'+
                    '<option value="424">424: FailedDependency</option>'+
                    '<option value="425">425: TooEarly</option>'+
                    '<option value="426">426: UpgradeRequired</option>'+
                    '<option value="428">428: PreconditionRequired</option>'+
                    '<option value="429">429: TooManyRequests</option>'+
                    '<option value="431">431: RequestHeaderFieldsTooLarge</option>'+
                    '<option value="451">451: UnavailableForLegalReasons</option>'+
                    '<option value="500">500: InternalServerError</option>'+
                    '<option value="501">501: NotImplemented</option>'+
                    '<option value="502">502: BadGateway</option>'+
                    '<option value="503">503: ServiceUnavailable</option>'+
                    '<option value="504">504: GatewayTimeout</option>'+
                    '<option value="505">505: HTTPVersionNotSupported</option>'+
                    '<option value="506">506: VariantAlsoNegotiates</option>'+
                    '<option value="507">507: InsufficientStorage</option>'+
                    '<option value="508">508: LoopDetected</option>'+
                    '<option value="510">510: NotExtended</option>'+
                    '<option value="511">511: NetworkAuthenticationRequired</option>'+
                '</select>'+
                '<select onChange="checkExpectChange($(this));" style="width:40%;display:none;" class="form-control input-sm serviceProcessParam serviceProcessParamSelector" value="restart">'+
                    '<option value="">None</option>'+
                    '<option value="restart" title="Select this to send alerts when a matched process restarts." selected>'+
                        'Restarts'+
                    '</option>'+
                    '<option value="running" title="Select this to send alerts when a matched process is found running. Uncommon, and not usable with other options.">'+
                        'Running'+
                    '</option>'+
                '</select>'+
                '<input type="number" min="0" onChange="checkExpectChange($(this));" title="Minimum number of proesses allowed to run." placeholder="Min" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceProcessParam serviceProcessParamMin" value="0" style="width:30%;display:none;">'+
                '<input type="number" min="0" onChange="checkExpectChange($(this));" title="Maximm number of process allowed to run." placeholder="Max" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceProcessParam serviceProcessParamMax" value="0" style="width:30%;display:none;">'+
                '<input disabled type="text" data-app="checks" value="unused" class="form-control input-sm serviceTCPParam" style="display:none;">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Interval" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Interval" data-original="5m" value="5m">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Timeout" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Timeout" data-original="1m" value="1m">'+
            '</div>'+
        '</div>'+
    '</td>'+
'</tr>';

    serviceTable.row.add($(row)).draw();

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
            checkExpectChange(ctl.find('.serviceProcessParam').show());
            break;
        case "http":
            checkExpectChange(ctl.find('.serviceHTTPParam').show());
            break;
        case "tcp":
            checkExpectChange(ctl.find('.serviceTCPParam').show());
            break;
    }

    toggleServiceTypeSelects();
}

// This fires when an expect value (for process type) is changed.
// Some values are incompatible with others, and the whole thing is copied
// into the "real" value, so it can be POSTed properly (it's a string).
function checkExpectChange(from)
{
    const ctl = from.closest('.services-Checks'); // just this cell.
    const run = ctl.find('.serviceProcessParamSelector').val() == 'running';

    if (from.hasClass('serviceHTTPParam')) { // it's an "http" check.
        ctl.find('.serviceProcessParamExpect').val(from.find(":selected").val());
    } else if (from.hasClass('serviceTCPParam')) { // it's a "tcp" check.
        ctl.find('.serviceProcessParamExpect').val('');
    } else if (run) { // it's a "process" check.
        // The "running" checkbox does not allow any other arguments, so deal with that here.
        // Copy "running" into real value that is POSTed.
        ctl.find('.serviceProcessParamExpect').val('running');
        // Disable incompatible options.
        ctl.find('.serviceProcessParamMin').prop('disabled', true);
        ctl.find('.serviceProcessParamMax').prop('disabled', true);
    } else {
        const min = ctl.find('.serviceProcessParamMin').val();
        const max = ctl.find('.serviceProcessParamMax').val();
        const res = ctl.find('.serviceProcessParamSelector').val() == 'restart';
        // Copy the concatenated string (from three sources) into the real value.
        ctl.find('.serviceProcessParamExpect').val('count:'+ min +':'+ max + (res ? ',restart' : ''));
        // Enable all values.
        ctl.find('.serviceProcessParamMin').prop('disabled', false);
        ctl.find('.serviceProcessParamMax').prop('disabled', false);
    }
}
