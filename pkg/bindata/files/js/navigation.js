// swapNavigationTemplate changes the currently displayed navigation page(div).
function swapNavigationTemplate(template)
{
    // only swap if there is 1 page to swap to.
    if ($('#template-'+ template).length === 1) {
        $('.navigation-item').hide();
        $('#template-'+ template).show();
    }
}

// checkHashForNavPage allows passing in a URL #hash as a navigation page.
function checkHashForNavPage()
{
    const hash = $(location).attr('hash');
    if (hash != "") {
        swapNavigationTemplate(hash.substring(1)); // Remove the # on the beginning.
    }
}

// This only needs to run once on startup. This sends the user to the correct page (like when they refresh).
checkHashForNavPage();

// refreshPage will re-download any template and replace it with new data.
function refreshPage(template, notice = true)
{
    $.ajax({
        url: 'template/'+ template,
        async: false,
        success: function (data){
            if (notice) {
                // Sometimes refreshes happen so quick we need a message to tell us it worked.
                toast('Refreshed', 'Refresh complete.', 'success', 2000);
            }

            $('#template-'+ template).html(data);
            // refreshPage is used on at least 3 pages that have line counter boxes, so update those.
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
