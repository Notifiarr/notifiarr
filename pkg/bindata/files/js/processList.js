function showProcessList() {
    swapNavigationTemplate('processlist');
    $('#process-list-content').html('<h4><i class="fas fa-cog fa-spin"></i> Loading process list...</h4>');

    $.ajax({
        url: 'ps',
        success: function (data){
            $('#process-list-content').text(data);
            updateLogFileContentCounters();
        },
        error: function (request, status, error) {
            $('#process-list-content').html('<h4>'+ error +'</h4>\n'+ request.responseText).animate({opacity:'100'}).show().fadeOut(10000);
        },
    });
}
