// updateFileContentCounters is used for things besides just the file viewer.
// It's also used by the process list viewer. It adds numbers to lines.
function updateFileContentCounters()
{
    $.each($('.file-content'), function() {
        let fileContainer    = $(this);
        let lines           = fileContainer.html().split(/\n/);
        const length        = getCharacterLength(lines.length.toString().trim());

        if (fileContainer.find('.cl').length == 0) { // avoid repeating this to an already formatted container.
            fileContainer.html('');
            $.each(lines, function(index, line) {
                if (index !== (lines.length-1)) { // skip the last newline.
                    let number = fileContainer.find('.line-number').length+1;
                    fileContainer.append('<span class="line-number">' + number.toString().padStart(length, ' ') + '</span>' + line + '<span class="cl"></span>');
                }
            });
        }
    });
}

updateFileContentCounters();

// ----- File Display
function fileSelectorChange(caller) {
    const ctl       = caller.parents('.fileController');
    const kind      = ctl.data("kind");
    const fileID    = ctl.find('.file-selector').val();
    const filename  = ctl.find('#fileinfo-'+ fileID).data('filename');
    const used      = ctl.find('#fileinfo-'+ fileID).data('used');
    const lineCount = parseInt(ctl.data("lines"));

    // start a spinner because it may take a few seconds to load a file.
    ctl.find('.file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
    // hide all other file-info divs, hide actions (until file load completes), hide help msg (this is the active file warning tooltip), hide any error or info messages.
    ctl.find('.fileinfo-table, .file-control, .file-ajax-error, .file-ajax-msg').hide()
    ctl.find('.file-actions, #fileinfo-'+ fileID).show() // show the file info div (right side panel) for the file requested.
    ctl.find('.fileLinesAdd').prop('disabled', false);

    $.ajax({
        url: 'getFile/'+ kind +'/'+ fileID +'/'+ lineCount +'/0', // the zero is optional, skip counter.
        success: function (data){
            ctl.find('.file-content').text(data);
            const gotLines = ctl.find('.file-content').html().split(/\n/).length-1;
            if (gotLines === lineCount) {
                ctl.find('.file-action-msg').html('Displaying last <span class="currentLineCount">'+ gotLines +'</span> file lines.');
            } else { // EOF
                ctl.find('.fileLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                ctl.find('.fileLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                ctl.find('.file-action-msg').html('Displaying all <span class="currentLineCount">'+ gotLines +'</span> file lines. This is the whole file.');
            }
            ctl.find('.file-control').show();
            updateFileContentCounters();
        },
        error: function (request, status, error) {
            if (error == "") {
                ctl.find('.file-ajax-error').html('<h4>Web Server Error</h4>Notifiarr client appears to be down! Hard refresh recommended.');
            } else {
                ctl.find('.file-ajax-error').html('<h4>'+ error +'</h4>'+ request.responseText);
            }
            ctl.find('.file-ajax-error').stop().animate({opacity:'100'}).show().fadeOut(10000);
        },
    });
}

// This powers the Action menu: Send to notifiarr, delete, download.
function triggerFileAction(caller) {
    const ctl      = caller.parents('.fileController');
    const kind     = ctl.data("kind");
    const action   = ctl.find('.fileAction').val();
    const fileID   = ctl.find('.file-selector').val();
    const filename = ctl.find('#fileinfo-'+ fileID).data('filename');

    if (filename === undefined) {
        return;
    }

    if (action == "download") {
        ctl.find('.file-ajax-msg').html("<h4>Downloading File</h4>"+filename+".zip").stop().animate({opacity:'100'}).show().fadeOut(3000);
        $(location).attr("href", "downloadFile/"+ kind +"/"+ fileID);
    } else if (action == "delete") {
        $.ajax({
            url: 'deleteFile/'+ kind +'/'+ fileID,
            success: function (data){
                ctl.find('.file-control').hide();                                       // hide controls for this deleted file.
                ctl.find('.file-selector').val("placeholder");                    // set selection back to placeholder.
                ctl.find(".file-selector option[value='"+ fileID +"']").remove(); // remove this file from the list.
                ctl.find('.file-ajax-msg').html("<h4>Deleted File</h4>"+filename).      // give the user a success message.
                    stop().animate({opacity:'100'}).show().fadeOut(10000);
            },
            error: function (request, status, error){
                if (error == "") {
                    ctl.find('.file-ajax-error').html('<h4>Web Server Error</h4>Notifiarr client appears to be down! Hard refresh recommended.');
                } else {
                    ctl.find('.file-ajax-error').html('<h4>'+ error +'</h4>'+ request.responseText);
                }
                ctl.find('.file-ajax-error').stop().animate({opacity:'100'}).show().fadeOut(10000);
            },
        });
    } else if (action == "notifiarr") {
        ctl.find('.file-ajax-error').html('<h4>Invalid!</h4>This does not work yet.')
            .stop().animate({opacity:'100'}).show().fadeOut(4000);
    }
}

// This powers the file add/reload menu.
function triggerFileLoad(caller) {
    const ctl         = caller.parents('.fileController');
    const kind        = ctl.data("kind");
    const fileID      = ctl.find('.file-selector').val();
    const filename    = ctl.find('#fileinfo-'+ fileID).data('filename');
    const used        = ctl.find('#fileinfo-'+ fileID).data('used');
    const lineCount   = parseInt(ctl.find('.fileLinesCount').val());
    const lineAction  = ctl.find('.fileLinesAction').val(); // add/reload
    const offSetCount = parseInt(ctl.find('.currentLineCount').html());

    ctl.find('.file-ajax-error, .file-ajax-msg').hide();

    // needs to update an "ok, working on that" box (that does not exist right now),

    if (lineAction == 'linesAdd') {
        ctl.find('.line-number').remove();
        ctl.find('.file-content').html(ctl.find('.file-content').html().toString().replaceAll('<span class="cl"></span>', '\n'));
        ctl.find('.file-ajax-msg').html('Getting '+ lineCount +' more lines!').stop().animate({opacity:'100'}).show().fadeOut(lineCount);
        ctl.find('.file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');

        $.ajax({
            url: 'getFile/'+ kind +'/'+ fileID +'/'+ lineCount +'/'+ offSetCount,
            success: function (data){
                ctl.find('.file-content').prepend(data);
                const gotLines = data.split(/\n/).length-1;
                const totalLines  = gotLines + offSetCount;

                if (gotLines === lineCount) {
                    ctl.find('.file-action-msg').html('Displaying last <span class="currentLineCount">'+ totalLines +'</span> file lines.');
                } else {
                    ctl.find('.fileLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                    ctl.find('.fileLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                    ctl.find('.file-action-msg').html('Displaying all <span class="currentLineCount">'+ totalLines +'</span> file lines. This is the whole file.');
                }

                updateFileContentCounters();
                ctl.find('.file-small-msg').html('');
            },
            error: function (request, status, error){
                if (error == "") {
                    ctl.find('.file-ajax-error').html('<h4>Web Server Error</h4>Notifiarr client appears to be down! Hard refresh recommended.');
                } else {
                    ctl.find('.file-ajax-error').html('<h4>'+ error +'</h4>'+ request.responseText);
                }
                ctl.find('.file-ajax-error').stop().animate({opacity:'100'}).show().fadeOut(10000);
            },
        });
    } else { // reload (vs add).
        // start a spinner because this might take a few seconds.
        ctl.find('.file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
        $.ajax({
            url: 'getFile/'+ kind +'/'+ fileID +'/'+ lineCount +'/0',
            success: function (data){
                ctl.find('.file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');
                ctl.find('.file-content').text(data);
                const gotLines = ctl.find('.file-content').html().split(/\n/).length-1;
                if (gotLines === lineCount) {
                    ctl.find('.fileLinesAdd').prop('disabled', false) // Enable the 'Add' action.
                    ctl.find('.file-action-msg').html('Displaying last <span class="currentLineCount">'+ gotLines +'</span> file lines.');
                } else {
                    ctl.find('.file-action-msg').html('Displaying all <span class="currentLineCount">'+ gotLines +'</span> file lines. This is the whole file.');
                }

                updateFileContentCounters();
                ctl.find('.file-small-msg').html('');
            },
            error: function (request, status, error){
                ctl.find('.file-ajax-error').show().html('<h4>'+ error +'</h4>'+ request.responseText);
                ctl.find('.file-ajax-error').stop().animate({opacity:'100'}).show().fadeOut(10000);
            },
        });
    }
}
