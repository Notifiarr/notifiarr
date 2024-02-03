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
        "sZeroRecords": "No matching service checks found.",
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
            if (response.status == 0) {
                $('#process-list-content').html('<h4>Web Server Error</h4>Notifiarr client appears to be down! Hard refresh recommended.');
            } else {
                $('#process-list-content').html('<h4>'+ (error!=''?error:'Bad Request') +'</h4>'+ request.responseText);
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
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Error', (error!=''?error:'Bad Request')+': '+request.responseText, 'error', 10000);
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

    let options = '<option value="0s">No Timeout</option><option value="1s">1 second</option>';
    for (i = 2 ; i < 60 ; i++) {
        options += '<option value="'+i+'s">'+i+' seconds</option>';
    }
    options += '<option selected value="1m">1 minute</option>';
    for (i = 1 ; i < 60 ; i++) {
        options += '<option value="1m'+i+'s">1 min '+i+' sec</option>';
    }

    const row = '<tr class="bk-success services-Checks newRow" id="services-Checks-'+ index +'">'+
    '<td style="white-space:nowrap;">'+
        '<div class="btn-group" role="group" style="display:flex;">'+
            '<button onclick="removeInstance(\'services-Checks\', '+ index +')" type="button" class="delete-item-button btn btn-danger btn-sm" style="font-size:18px;width:35px;"><i class="fa fa-minus"></i></button>'+
            '<button id="checksIndexLabel'+ index +'" class="btn btn-sm" style="font-size:18px;width:35px;pointer-events:none;">'+ instance +'</button>'+
            '<button onClick="testService($(this), \''+index+'\')" type="button" style="font-size:18px;" class="btn btn-success btn-sm"><i class="fas fa-check-double"></i></button>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Name" name="Service.'+ index +'.Name" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Name" data-original="" value="Service'+ instance +'">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<select id="Service.'+ index +'.Type" name="Service.'+ index +'.Type" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm serviceTypeSelect" onChange="checkTypeChange($(this));" data-group="services" data-label="Check '+ instance +' Type" value="http" data-original="">'+
                    '<option value="process">Process</option>'+
                    '<option value="http" selected>HTTP</option>'+
                    '<option value="tcp">TCP Port</option>'+
                    '<option value="ping">UDP Ping</option>'+
                    '<option value="icmp">ICMP Ping</option>'+
                '</select>'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input type="text" id="Service.'+ index +'.Value" name="Service.'+ index +'.Value" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Value" data-original="">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<input id="Service.'+ index +'.Expect" name="Service.'+ index +'.Expect" data-index="'+ index +'" data-app="checks" class="client-parameter serviceProcessParamExpect" data-group="services" data-label="Check '+ instance +' Expect" value="200" data-original="200" style="display:none;">'+
                '<select multiple id="Service.'+ index +'.Expect.StatusCode" onChange="checkExpectChange($(this));" value="200" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceHTTPParam" style="width:100%">'+
                    '<option value="SSL">SSL: Validate Certificate</option>'+
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
                    '<option value="restart" selected>'+
                        'Restarts'+
                    '</option>'+
                    '<option value="running">'+
                        'Running'+
                    '</option>'+
                '</select>'+
                '<input type="number" min="1" max=50 onChange="checkExpectChange($(this));" title="Count of packets to sent." class="form-control input-sm servicePingParam servicePingParamCount" value="3" style="width:30%;display:none;">'+
                '<input type="number" min="1" max=50 onChange="checkExpectChange($(this));" title="Packets that must be received." class="form-control input-sm servicePingParam servicePingParamMin" value="2" style="width:30%;display:none;">'+
                '<input type="number" min="100" max=10000 onChange="checkExpectChange($(this));" title="Interval in milliseconds between packets." class="form-control input-sm servicePingParam servicePingParamInt" value="500" style="width:40%;display:none;">'+
                '<input type="number" min="0" onChange="checkExpectChange($(this));" title="Minimum number of proesses allowed to run." placeholder="Min" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceProcessParam serviceProcessParamMin" value="0" style="width:30%;display:none;">'+
                '<input type="number" min="0" onChange="checkExpectChange($(this));" title="Maximm number of process allowed to run." placeholder="Max" data-index="'+ index +'" data-app="checks" class="form-control input-sm serviceProcessParam serviceProcessParamMax" value="0" style="width:30%;display:none;">'+
                '<input disabled type="text" data-app="checks" value="unused" class="form-control input-sm serviceTCPParam" style="display:none;">'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<select type="select" id="Service.'+ index +'.Interval" name="Service.'+ index +'.Interval" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Interval" data-original="5m">'+
                '<option value="45s">45 seconds</option>'+
                '<option value="1m">1 minute</option>'+
                '<option value="1m15s">1 min 15 sec</option>'+
                '<option value="1m30s">1 min 30 sec</option>'+
                '<option value="2m">2 minutes</option>'+
                '<option value="2m30s">2 min 30 sec</option>'+
                '<option value="3m">3 minutes</option>'+
                '<option value="4m">4 minutes</option>'+
                '<option selected value="5m">5 minutes</option>'+
                '<option value="10m">10 minutes</option>'+
                '<option value="15m">15 minutes</option>'+
                '<option value="30m">30 minutes</option>'+
                '<option value="45m">45 minutes</option>'+
                '<option value="60m">60 minutes</option>'+
                '</select>'+
            '</div>'+
        '</div>'+
    '</td>'+
    '<td>'+
        '<div class="form-group" style="width:100%">'+
            '<div class="input-group" style="width:100%">'+
                '<select type="select" id="Service.'+ index +'.Timeout" name="Service.'+ index +'.Timeout" data-app="checks" data-index="'+ index +'" class="client-parameter form-control input-sm" data-group="services" data-label="Check '+ instance +' Timeout" data-original="1m">'+
                options+
                '</select>'+
            '</div>'+
        '</div>'+
    '</td>'+
'</tr>';

    // Add the new row through servicetable.
    serviceTable.row.add($(row)).draw();

    // setup the select2 selector for the http status codes.
    $('[id="Service.'+ index +'.Expect.StatusCode"]').select2({
        placeholder: 'HTTP Status Codes..',
        templateSelection: function(state) {
            return state.id ? state.id : state.text
        },
    });

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
    ctl.find('.servicePingParam').hide();

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
        case "ping":
        case "icmp":
            checkExpectChange(ctl.find('.servicePingParam').show());
            break;
    }

    toggleServiceTypeSelects();
}

// This fires when an expect value (for process type) is changed.
// Some values are incompatible with others, and the whole thing is copied
// into the "real" value, so it can be POSTed properly (it's a string).
function checkExpectChange(from) {
    const ctl = from.closest('.services-Checks'); // just this row.
    const run = ctl.find('.serviceProcessParamSelector').val() == 'running';
    const expect = ctl.find('.serviceProcessParamExpect')

    if (from.hasClass('servicePingParam')) { // it's a "ping" or "icmp" check.
        expect.val(ctl.find('.servicePingParamCount').val() +':'+ 
            ctl.find('.servicePingParamMin').val() + ':'+ 
            ctl.find('.servicePingParamInt').val());
    } else if (from.hasClass('serviceHTTPParam')) { // it's an "http" check.
        // Copy comma-concatenated values into real 'expect' value.
        expect.val(from.val().join());
    } else if (from.hasClass('serviceTCPParam')) { // it's a "tcp" check.
        expect.val('');
    } else if (run) { // it's a "process" check in "running" mode.
        // Copy "running" into real 'expect' value that is POSTed.
        expect.val('running');
        // Disable 'running'-incompatible options.
        ctl.find('.serviceProcessParamMin').prop('disabled', true);
        ctl.find('.serviceProcessParamMax').prop('disabled', true);
    } else { // it's a process check not in "running" node.
        const min = ctl.find('.serviceProcessParamMin').val();
        const max = ctl.find('.serviceProcessParamMax').val();
        const res = ctl.find('.serviceProcessParamSelector').val() == 'restart';
        // Copy the concatenated string (from three sources) into the real value.
        expect.val('count:'+ min +':'+ max + (res ? ',restart' : ''));
        // Enable all values.
        ctl.find('.serviceProcessParamMin').prop('disabled', false);
        ctl.find('.serviceProcessParamMax').prop('disabled', false);
    }
}


function testService(from, index)
{
    const ctl = from.closest('.services-Checks');
    const checkType = titleCaseWord($('[id="Service.'+index+'.Type"]').val())

    from.css({'pointer-events':'none'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
    let fields = '';
    ctl.find('.client-parameter').each(function() {
        const id = $(this).attr('id')
        if (id !== undefined) {
            fields += '&' + $(this).serialize();
        }
    });

    $.ajax({
        type: 'POST',
        url: URLBase+'checkInstance/'+checkType+"/"+index,
        data: fields,
        success: function (data){
            from.css({'pointer-events':'auto'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
            toast(checkType+' Check Successful', data, 'success');
        },
        error: function (response, status, error) {
            from.css({'pointer-events':'auto'}).find('i').toggleClass('fa-cog fa-spin fa-check-double');
            if (response.status == 0) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast(checkType+' Check Error', (error!=''?error:'Bad Request')+': '+response.responseText, 'error', 15000);
            }
        }
    });
}
