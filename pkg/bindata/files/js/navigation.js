function swapNavigationTemplate(template) {
    // only swap if there is 1 page to swap to.
    if ($('#template-'+ template).length === 1) {
        $('.navigation-item').hide();
        $('#template-'+ template).show();
    }
}

// checkHashForNavPage allows passing in a URL #hash as a navigation page.
function checkHashForNavPage() {
    const hash = $(location).attr('hash');
    if (hash != "") {
        swapNavigationTemplate(hash.substring(1)); // Remove the # on the beginning.
    }
}

checkHashForNavPage()

// refreshPage will re-download any template and replace it with new data.
function refreshPage(template) {
    $.ajax({
        url: 'template/'+ template,
        success: function (data){
            toast('Refreshed', 'Refresh complete.', 'success', 2000);
            $('#template-'+ template).html(data);
            updateFileContentCounters();
        },
        error: function (request, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Template Error', error+': '+response.responseText, 'error', 10000);
            }
        },
    });
}
