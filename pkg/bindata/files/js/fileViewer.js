// updateFileContentCounters is used to add numbers to lines in a <pre> block.
// It's also used by the process list viewer, not just the fileViewer.
function updateFileContentCounters()
{
    $.each($('.file-content'), function() {
        let fileContainer    = $(this);
        let lines           = fileContainer.html().split(/\n/);
        const length        = getCharacterLength(lines.length.toString().trim());

        if (fileContainer.find('.line-number').length == 0) { // avoid repeating this to an already formatted container.
            fileContainer.html('');
            $.each(lines, function(index, line) {
                if (index !== (lines.length-1)) { // skip the last newline.
                    let number = fileContainer.find('.line-number').length+1;
                    fileContainer.append('<span class="line-number">' + number.toString().padStart(length, ' ') + '</span>' + line + '<br>');
                }
            });
        }
    });
}

// Update them all on page load.
updateFileContentCounters();

// ----- File Display
function fileSelectorChange(caller, fileID)
{
    const ctl = caller.parents('.fileController');

    if (!fileID) {
        fileID = ctl.data('currentid');
    }

    if (!fileID) {
        return; // a file has not been selected yet.
    }

    const source    = ctl.data("kind");
    const filename  = ctl.find('#fileinfo-'+ fileID).data('filename');
    const used      = ctl.find('#fileinfo-'+ fileID).data('used');
    const lineCount = parseInt(ctl.find('.fileLinesCount').val());

    let from = "last";
    let sort = ctl.find('.fileSortDirection').data('sort');
    if (sort  != "heads" && sort != "tails") {
        sort = "tails";
    } else if (sort == "heads") {
        from = "first";
    }

    destroyWebsocket(source);

    // set the current ID to a global area.
    ctl.data('currentid', fileID)
    // start a spinner because it may take a few seconds to load a file.
    ctl.find('.file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
    // hide all other file-info divs, hide actions (until file load completes), hide help msg (this is the active file warning tooltip), hide any error or info messages.
    ctl.find('.fileinfo-table, .tailControl, .file-control').hide()
    ctl.find('.file-actions, .fileTablesList, #fileinfo-'+ fileID).show() // show the file info div (right side panel) for the file requested.
    ctl.find('.fileLinesAdd').prop('disabled', false);

    $.ajax({
        url: URLBase+'getFile/'+ source +'/'+ fileID +'/'+ lineCount +'/0' + '?sort='+sort,
        success: function (data){
            ctl.find('.file-content').text(data);
            const gotLines = ctl.find('.file-content').html().split(/\n/).length-1;
            if (gotLines === lineCount) {
                ctl.find('.file-action-msg').html('Displaying '+ from +' <span class="currentLineCount">'+ gotLines +'</span> file lines.');
            } else { // EOF
                ctl.find('.fileLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                ctl.find('.fileLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                ctl.find('.file-action-msg').html('Displaying the whole file; <span class="currentLineCount">'+ gotLines +'</span> lines.');
            }
            ctl.find('.file-control').show();
            updateFileContentCounters();
        },
        error: function (request, status, error) {
            if (response.status == 0) {
                toast('Web Server Error', 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 10000);
            } else {
                toast(error!=''?error:'Bad Request', request.responseText, 'error', 10000);
            }
        },
    });
}

// This powers the Action menu: Send to notifiarr, delete, download.
function triggerFileAction(caller, action, source, fileID)
{
    const ctl      = caller.parents('.fileController');
    const filename = ctl.find('#fileinfo-'+ fileID).data('filename');

    if (filename === undefined) {
        return;
    }

    if (action == "download") {
        toast('Downloading File', filename+".zip", 'success')
        $(location).attr("href", "downloadFile/"+ source +"/"+ fileID);
    } else if (action == "delete") {
        $.ajax({
            url: URLBase+'deleteFile/'+ source +'/'+ fileID,

            success: function (data){
                let files = ' files';
                const count = ctl.find('.fileRow').length-1;
                if (count == 1) {
                    files = ' file';
                }

                ctl.find('.file-control').hide();                        // hide controls for this deleted file.
                ctl.data('currentid', "");                               // set selection back to placeholder.
                ctl.find('.fileListDirInfo').html(
                    count + files +' in ' +ctl.find('.fileListDirInfo').data('dirdata')); // update the counter at the top of the table.
                ctl.find("#fileRow"+fileID).fadeOut(1200, function() {   // hide the table row.
                    ctl.find("#fileRow"+fileID).remove()                 // remove this file from the list.
                });
                toast('Deleted File', filename, 'success', 8000)         // give the user a success message.
            },
            error: function (request, status, error){
                if (response.status == 0) {
                    toast('Web Server Error', 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 10000);
                } else {
                    toast(error!=''?error:'Bad Request', request.responseText, 'error', 10000);
                }
            },
        });
        setTooltips();
    } else if (action == "upload") {
        $.ajax({
            url: URLBase+'uploadFile/'+ source +'/'+ fileID,

            success: function (data){
                toast('Uploading File', filename, 'success', 8000)
            },
            error: function (request, status, error){
                if (response.status == 0) {
                    toast('Web Server Error', 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 10000);
                } else {
                    toast(error!=''?error:'Bad Request', request.responseText, 'error', 10000);
                }
            },
        });
    }

}

// This powers the file add/reload menu.
function triggerFileLoad(caller)
{
    const ctl         = caller.parents('.fileController');
    const source      = ctl.data("kind");
    const fileID      = ctl.data('currentid');
    const filename    = ctl.find('#fileinfo-'+ fileID).data('filename');
    const used        = ctl.find('#fileinfo-'+ fileID).data('used');
    const lineCount   = parseInt(ctl.find('.fileLinesCount').val());
    const lineAction  = ctl.find('.fileLinesAction').val(); // add/reload
    const offSetCount = parseInt(ctl.find('.currentLineCount').html());

    let from = "last";
    let sort = ctl.find('.fileSortDirection').data('sort');
    if (sort == "tails") {
        from = "last";
    } else if (sort  != "heads" && sort != "tails") {
        sort = "tails";
    } else {
        from = "first";
    }

    // needs to update an "ok, working on that" box (that does not exist right now),

    if (lineAction == 'linesAdd') { // addlines
        toast('Working', 'Getting '+ lineCount +' more lines!', 'success')
        ctl.find('.file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');

        $.ajax({
            url: URLBase+'getFile/'+ source +'/'+ fileID +'/'+ lineCount +'/'+ offSetCount + '?sort='+sort,
            success: function (data){
                // These are slow when there's a lot of lines.
                ctl.find('.line-number').remove();
                ctl.find('.file-content br').after('\n').remove();
                if (sort == "tails") {
                    ctl.find('.file-content').prepend($('<div/>').text(data).html());
                } else {
                    ctl.find('.file-content').append($('<div/>').text(data).html());
                }

                const gotLines = data.split(/\n/).length-1;
                const totalLines  = gotLines + offSetCount;

                if (gotLines === lineCount) {
                    ctl.find('.file-action-msg').html('Displaying '+ from +' <span class="currentLineCount">'+ totalLines +'</span> file lines.');
                } else {
                    ctl.find('.fileLinesAdd').prop('disabled', true) // Disable the 'Add' action.
                    ctl.find('.fileLinesAction').val("linesReload"); // Set the selector to reload (instead of Add).
                    ctl.find('.file-action-msg').html('Displaying the whole file; <span class="currentLineCount">'+ totalLines +'</span> lines.');
                }

                updateFileContentCounters();
                ctl.find('.file-small-msg').html('');
            },
            error: function (request, status, error){
                if (response.status == 0) {
                    toast('Web Server Error', 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 10000);
                } else {
                    toast(error!=''?error:'Bad Request', request.responseText, 'error', 10000);
                }
            },
        });
    } else { // reload (vs add).
        // start a spinner because this might take a few seconds.
        ctl.find('.file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading up to '+lineCount+' lines from file '+ filename +'...');
        $.ajax({
            url: URLBase+'getFile/'+ source +'/'+ fileID +'/'+ lineCount +'/0' + '?sort='+sort,
            success: function (data){
                ctl.find('.file-small-msg').html('<i class="fas fa-cog fa-spin"></i> Still Loading...');
                ctl.find('.file-content').text(data);
                const gotLines = ctl.find('.file-content').html().split(/\n/).length-1;
                if (gotLines === lineCount) {
                    ctl.find('.fileLinesAdd').prop('disabled', false) // Enable the 'Add' action.
                    ctl.find('.file-action-msg').html('Displaying '+ from +' <span class="currentLineCount">'+ gotLines +'</span> file lines.');
                } else {
                    ctl.find('.file-action-msg').html('Displaying the whole file; <span class="currentLineCount">'+ gotLines +'</span> lines.');
                }

                updateFileContentCounters();
                ctl.find('.file-small-msg').html('');
            },
            error: function (request, status, error){
                if (response.status == 0) {
                    toast('Web Server Error', 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 10000);
                } else {
                    toast(error!=''?error:'Bad Request', request.responseText, 'error', 10000);
                }
            },
        });
    }
}

// toggleButton handles the heads/tails toggle switch. Reloads current file.
function toggleButton(from)
{
    const ctl = from.parent('.fileSortDirection');
    ctl.find('.toggleButton').toggleClass('btn-seondary btn-brand');

    if (ctl.data('sort') == 'heads') {
        ctl.data('sort', 'tails'); // was heads now, going to tails.
        ctl.find('.toggleIcon').attr('title', 'Tails, showing bottom of file first.').toggleClass('fa-sort-amount-up fa-sort-amount-down');
    } else {
        ctl.data('sort', 'heads'); // was tails now, going to heads.
        ctl.find('.toggleIcon').attr('title', 'Heads, showing top of file first.').toggleClass('fa-sort-amount-up fa-sort-amount-down');
    }

    const source = ctl.closest('.fileController').data("kind");
    if (websockets.source) {
        openWebsocket(ctl);
    } else {
        fileSelectorChange(ctl); // update the file viewer.
        setTooltips();
    }
}
