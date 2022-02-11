function updateLogFileContentCounters() 
{
    $.each($('.log-file-content'), function() {
        let logContainer    = $(this);
        let lines           = logContainer.html().split(/\n/);
        const length        = getCharacterLength(lines.length.toString().trim());
        logContainer.html('');

        $.each(lines, function(index, line) {
            if (index !== (lines.length-1)) { // skip the last newline.
                let number = $('.line-number').length + 1;
                logContainer.append('<span class="line-number">' + number.toString().padStart(length, ' ') + '</span>' + line + '<span class="cl"></span>');
            }
        });
    });

}
// ---------------------------------------------------------------------------------------------

updateLogFileContentCounters();
