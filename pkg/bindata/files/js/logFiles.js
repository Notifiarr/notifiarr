function updateLogFileContentCounters()
{
    $.each($('.log-file-content'), function() {
        let logContainer    = $(this);
        let lines           = logContainer.html().split(/\n/);
        const length        = getCharacterLength(lines.length.toString().trim());
        logContainer.html('');

        $.each(lines, function(index, line) {
            if (index !== (lines.length-1)) { // skip the last newline.
                let number = logContainer.find('.line-number').length+1;
                logContainer.append('<span class="line-number">' + number.toString().padStart(length, ' ') + '</span>' + line + '<span class="cl"></span>');
            }
        });
    });

}
// ---------------------------------------------------------------------------------------------

updateLogFileContentCounters();

$(document).ready((function() {
    // ----- Log File Display
    $('#LogFileSelector').change(function() {
        const logFileID   = $(this).val();
        const filename    = $('#fileinfo-'+ logFileID).data('filename');
        const used        = $('#fileinfo-'+ logFileID).data('used');
        const lineCount   = 500;

        // start a spinner because it takes a few seconds to load a log file.
        $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
        // hide all other file-info divs, hide actions (until file load completes), hide help msg (this is the active file warning tooltip), hide any error or info messages.
        $('[id^=fileinfo-], .log-file-control, #logHelpMsg, #log-file-ajax-error, #log-file-ajax-msg').hide()
        $('.log-file-actions, #fileinfo-'+ logFileID).show() // show the file info div (right side panel) for the file requested.
        $('#logLinesAdd').prop('disabled', false);

        if (used == 'true') {
            $('#logHelpMsg').show() // display the help tooltip for this 'special' file.
        }

        $.ajax({
            url: 'getLog/'+ logFileID +'/'+ lineCount +'/0', // the zero is optional, skip counter.
            success: function (data){
                $('#log-file-content').text(data);
                const gotLines = $('.log-file-content').html().split(/\n/).length-1;
                if (gotLines === lineCount) {
                    $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ gotLines +'</span> log lines.');
                } else { // EOF
                    $('#logLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                    $('#logLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                    $('#log-file-action-msg').html('Displaying all <span id="logsCurrentLineCount">'+ gotLines +'</span> log lines. This is the whole file.');
                }
                $('.log-file-control').show();
                updateLogFileContentCounters();
            },
            error: function (request, status, error) {
                $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText).animate({opacity:'100'}).show().fadeOut(10000);
            },
        });
    });

    // This powers the Action menu: Send to notifiarr, delete, download.
    $('#triggerLogAction').click(function(){
        const action    = $('#logfileAction').val();
        const logFileID = $('#LogFileSelector').val();
        const filename  = $('#fileinfo-'+ logFileID).data('filename');

        if (filename === undefined) {
            return;
        }

        if (action == "download") {
            $('#log-file-ajax-msg').html("<h4>Downloading File</h4>"+filename+".zip").stop().animate({opacity:'100'}).show().fadeOut(3000);
            window.location.href = "downloadLog/"+logFileID; // this works so nice!
        } else if (action == "delete") {
            $.ajax({
                url: 'deleteLogFile/'+ logFileID,
                success: function (data){
                    // TODO: remove the item from the select
                    $('#log-file-ajax-msg').html("<h4>Deleted File</h4>"+filename).stop().animate({opacity:'100'}).show().fadeOut(10000);
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText);
                    $('#log-file-ajax-error').stop().animate({opacity:'100'}).show().fadeOut(10000);
                },
            });
        } else if (action == "notifiarr") {
            $('#log-file-ajax-error').html('<h4>Invalid!</h4>This does not work yet.')
                .stop().animate({opacity:'100'}).show().fadeOut(4000);
        }
    });

    // This powers the log file add/reload menu.
    $('#triggerLogLoad').click(function() {
        const logFileID   = $('#LogFileSelector').val();
        const filename    = $('#fileinfo-'+ logFileID).data('filename');
        const used        = $('#fileinfo-'+ logFileID).data('used');
        const lineCount   = parseInt($('#logLinesCount').val());
        const lineAction  = $('#logLinesAction').val(); // add/reload
        const offSetCount = parseInt($('#logsCurrentLineCount').html());

        $('#log-file-ajax-error, #log-file-ajax-msg').hide();

        // needs to update an "ok, working on that" box (that does not exist right now),

        if (lineAction == 'linesAdd') {
            $('.line-number').remove();
            $('#log-file-content').html($('#log-file-content').html().toString().replaceAll('<span class="cl"></span>', '\n'));
            $('#log-file-ajax-msg').html('Getting '+ lineCount +' more lines!').stop().animate({opacity:'100'}).show().fadeOut(lineCount);
            $('#log-file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');

            $.ajax({
                url: 'getLog/'+ logFileID +'/'+ lineCount +'/'+ offSetCount,
                success: function (data){
                    $('#log-file-content').prepend(data);
                    const gotLines = data.split(/\n/).length-1;
                    const totalLines  = gotLines + offSetCount;

                    if (gotLines === lineCount) {
                        $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ totalLines +'</span> log lines.');
                    } else {
                        $('#logLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                        $('#logLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                        $('#log-file-action-msg').html('Displaying all <span id="logsCurrentLineCount">'+ totalLines +'</span> log lines. This is the whole file.');
                    }

                    updateLogFileContentCounters();
                    $('#log-file-small-msg').html('');
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText).stop().animate({opacity:'100'}).show().fadeOut(10000);
                },
            });
        } else { // reload
            // start a spinner because this takes a few seconds.
            $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
            $.ajax({
                url: 'getLog/'+ logFileID +'/'+ lineCount +'/0',
                success: function (data){
                    $('#log-file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');
                    $('#log-file-content').text(data);
                    const gotLines = $('.log-file-content').html().split(/\n/).length-1;
                    if (gotLines === lineCount) {
                        $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+ gotLines +'</span> log lines.');
                    } else {
                        $('#log-file-action-msg').html('Displaying all <span id="logsCurrentLineCount">'+ gotLines +'</span> log lines. This is the whole file.');
                    }

                    updateLogFileContentCounters();
                    $('#log-file-small-msg').html('');
                },
                error: function (request, status, error){
                    $('#log-file-ajax-error').show();
                    $('#log-file-ajax-error').html('<h4>'+ error +'</h4>\n'+ request.responseText);
                },
            });
        }
    });
}));
