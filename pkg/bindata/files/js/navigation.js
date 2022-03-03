// swapNavigationTemplate changes the currently displayed navigation page(div).
function swapNavigationTemplate(template)
{
    // only swap if there is 1 page to swap to.
    if ($('#template-'+ template).length === 1) {
        $('.navigation-item').hide();
        $('#template-'+ template).show();
        // The following fixes any data tables.
        // https://datatables.net/examples/api/tabs_and_scrolling.html
        $.fn.dataTable.tables({ visible: true, api: true }).columns.adjust();
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
        url: URLBase+'template/'+ template,
        async: false,
        success: function (data){
            if (notice) {
                // Sometimes refreshes happen so quick we need a message to tell us it worked.
                toast('Refreshed', 'Refresh complete.', 'success', 2000);
            }

            $('#template-'+ template).html(data);
            // refreshPage is used on at least 3 pages that have line counter boxes, so update those.
            updateFileContentCounters();
            setTooltips();
            // Load any data tables on this page.
            $('#template-'+ template).find('.filetable').each(function() {
                loadDataTable($(this));
            });
            $('#template-'+ template).find('.monitortable').each(function() {
                loadMonitorTable($(this));
            });
            $('#template-'+ template).find('.configtable').each(function() {
                loadConfigTable($(this));
            });

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

function shutDownApp()
{
    if (confirm('Really shut down application?')) {
        $.ajax({
            url: URLBase+'shutdown',
            async: false,
            success: function (data){
                toast('Shuting Down', 'Application Shutdown Initiated', 'success', 15000);
            },
             error: function (request, status, error) {
                 if (error == "") {
                     toast('Web Server Error',
                         'Notifiarr client appears to be down aleady.', 'error', 30000);
                 } else {
                     toast('Shutdown Error', error+': '+response.responseText, 'error', 10000);
                 }
             },
         });
     }
}

function reloadConfig()
{
    $.ajax({
        url: URLBase+'reload',
        success: function (data){
            setTimeout(function(){location.reload();}, 5000); // reload window in 5 seconds.
            toast('Reloading', 'Application Reload Initiated', 'success');
        },
         error: function (request, status, error) {
             if (error == "") {
                 toast('Web Server Error',
                 'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
             } else {
                 toast('Reload Error', error+': '+response.responseText, 'error', 10000);
             }
         },
     });
}
