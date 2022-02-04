function jsLoader() {
    let path        = '';
    let script      = '';
    const files     = ['navigation'];

    for (const file of files){
        path        = 'files/js/' + file + '.js';
        script      = document.createElement('script');
        script.src  = path;
        document.head.appendChild(script);
    }
}
// -------------------------------------------------------------------------------------------

function ajax(url, method, type) {
    return new Promise((resolve) => {
        $.ajax({
            type: method,
            url: url,
            dataType: type,
            success: function (resultData) {
                resolve(resultData);
            }
        });
    });
}

// -------------------------------------------------------------------------------------------

function setTooltips() {
    $('[class*="balloon-tooltip"]').hide();

    $('a, div, i, img, input, span, td, button').balloon({
        position: 'bottom',
        classname: 'balloon-tooltip',
        css: {
            fontSize: '18px',
            borderRadius: '12px',
            height: 'auto',
            maxWidth: '400px',
            minWidth: '100px',
            padding: '0.5em',
            opacity: 0.95,
            borderColor: '#FFF',
        }
    });
    /*
        contents:null,
        url:null,
        ajaxComplete:null,
        ajaxContentsMaxAge: -1,
        html:false,
        classname:null,
        position:"top",
        offsetX: 0,
        offsetY: 0,
        tipSize: 12,
        tipPosition: 2,
        delay: 0,
        minLifetime: 200,
        maxLifetime: 0,
        showDuration: 100,
        show<a href="https://www.jqueryscript.net/animation/">Animation</a>:null,
        hideDuration:  80,
        hideAnimation:function(d) {this.fadeOut(d); },
        showComplete:null,
        hideComplete:null,
        css: {
          fontSize       :".7rem",
          minWidth       :"20px",
          padding        :"5px",
          borderRadius   :"6px",
          border         :"solid 1px #777",
          boxShadow      :"4px 4px 4px #555",
          color          :"#666",
          backgroundColor:"#efefef",
          zIndex         :"32767",
          textAlign      :"left"
        }
    */

}
// ---------------------------------------------------------------------------------------------

function findPendingChanges() {
    $('#pending-change-container').hide();
    $('#pending-change-list').html('');
    $('#pending-change-counter').html('0');

    let group;
    let label;
    let original;
    let current;
    let id;
    let changes = '';
    let counter = 0;

    $.each($('.client-parameter'), function() {
        id          = $(this).attr('id');
        label       = $(this).attr('data-label');
        group       = $(this).attr('data-group');
        original    = $(this).attr('data-original');
        current     = $(this).val();

        if (original != current) {
            counter++;
            changes += group.charAt(0).toUpperCase() + group.slice(1) +': '+ label +'<br>';
        }
    });

    if (changes) {
        $('#pending-change-list').html(changes);
        $('#pending-change-counter').html(counter);
        $('#pending-change-container').show();
    }
}
// ---------------------------------------------------------------------------------------------

$(document).ready((function() {
    // ----- Navbar
    $('.nav-link').click((function() {
        $('nav.ts-sidebar').toggleClass('menu-open', false);
    }))

    $(".menu-btn").click((function() {
        $('nav.ts-sidebar').toggleClass('menu-open');
    }))

    // ----- Log File Display
    $('#LogFileSelector').change(function() {
        const logFileID   = $(this).val();
        const filename    = $('#fileinfo-'+ logFileID).data('filename');
        const used        = $('#fileinfo-'+ logFileID).data('used');
        const lineCount   = 500;

        // start a spinner because this takes a few seconds.
        $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading content for file '+ filename +'...');
        // hide all other file-info divs, hide actions (until file load completes), hide help msg (this is the active file warning tooltip), hide any error or info messages.
        $('[id^=fileinfo-], #log-file-actions, #logHelpMsg, #log-file-ajax-error, #log-file-ajax-msg').hide()
        $('#fileinfo-'+ logFileID).show() // show the file info div for the file requested.

        if (used == 'true') {
            $('#logHelpMsg').show() // display the help tooltip for this 'special' file.
        }

        $.ajax({
            url: 'getLog/'+ logFileID +'/'+ lineCount +'/0', // the zero is optional, skip counter.
            context: document.body,
            success: function (data){
                $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ lineCount +'</span> log lines.');
                $('#log-file-actions').show();
                $('#log-file-content').html(data);
                updatePreCounters();
            },
            error: function (request, status, error) {
                $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText).show().fadeOut(10000);
            },
        });
    });

    $('#triggerLogAction').click(function(){
        const action    = $('#logfileAction').val();
        const logFileID = $('#LogFileSelector').val();
        const filename  = $('#fileinfo-'+ logFileID).data('filename');

        if (action == "download") {
            $('#log-file-ajax-msg').html("<h4>Downloading File</h4>"+filename+".zip").show().fadeOut(5000);
            window.location.href = "downloadLog/"+logFileID; // this works so nice!
        } else if (action == "delete") {
            // ajax call to  deleteFile/logFileID. needs to update an "ok, file deleted" box, or produce an error if there was an error (which does exist)

            $.ajax({
                url: 'deleteLogFile/'+ logFileID,
                context: document.body,
                success: function (data){
                    // refresh file list? I can return the new list here, how....?
                    $('#log-file-ajax-msg').html("<h4>Deleted File</h4>"+filename).show().fadeOut(10000);
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText);
                    $('#log-file-ajax-error').show().fadeOut(6000);
                },
            });
        }
    });

    $('#triggerLogLoad').click(function() {
        const logFileID   = $('#LogFileSelector').val();
        const filename    = $('#fileinfo-'+ logFileID).data('filename');
        const used        = $('#fileinfo-'+ logFileID).data('used');
        const lineCount   = parseInt($('#logLinesCount').val());
        const lineAction  = $('#logLinesAction').val(); // add/reload
        const offSetCount = parseInt($('#logsCurrentLineCount').html());
        const totalLines  = lineCount + offSetCount;

        $('#log-file-ajax-error').hide();
        $('#log-file-ajax-msg').hide();

        // needs to update an "ok, working on that" box (that does not exist right now),

        if (lineAction == 'linesAdd') {
            $('#log-file-ajax-msg').html("Getting "+lineCount+" more lines!");
            $('#log-file-ajax-msg').show().fadeOut(5000);
            $.ajax({
                url: 'getLog/'+ logFileID +'/'+ lineCount +'/'+ offSetCount,
                context: document.body,
                success: function (data){
                    $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ totalLines +'</span> log lines.');
                    $('#log-file-content').prepend(document.createTextNode(data));
                    updatePreCounters();
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText).show().fadeOut(10000);
                },
            });
        } else { // reload
            // start a spinner because this takes a few seconds.
            $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading content for file '+ filename +'...');

            $.ajax({
                url: "getLog/"+logFileID+"/"+lineCount+"/0",
                context: document.body,
                success: function (data){
                    $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ lineCount +'</span> log lines.');
                    $('#log-file-actions').show()
                    $('#log-file-content').html(data);
                    updatePreCounters();
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').show();
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText);
                },
            });
        }
    });

    updatePreCounters();
}));
// ---------------------------------------------------------------------------------------------

function updatePreCounters() {
    $.each($('.log-file-content'), function() {
        let logContainer = $(this);
        let lines = logContainer.html().split(/\n/);
        logContainer.html('');

        $.each(lines, function(index, line) {
            logContainer.append('<span class="line-number">'+ (index + 1) +'</span> '+ line +'<span class="cl"></span>');
        });
    });
}
// ---------------------------------------------------------------------------------------------
