function browseFiles(target, findFile) {
    $('<div>').dialog({
        title: 'Choose File: ' + $(target).data('label') + (findFile === undefined ? '' : ' ('+findFile+')'),
        modal: true,
        height: 'auto',
        resizable: false,
        position: {my: "center", at: "center", of: window},
        dialogClass: 'modal-body', // customized widths.
        show: {
            effect: 'fade',
            duration: 150
        },
        hide: {
            effect: 'fade',
            duration: 150
        }
    }).dialog('open').browse({
        separator: DirSep,
        root: DirSep == '\\' ? '' : '/',
        name: $(target).data('label'),
        dir: function(path) {
            return new Promise(function(resolve, reject) {
                $.ajax({
                    type: 'GET',
                    url: URLBase+'browse',
                    data: {dir: path},
                    dataType: 'json',
                    success: resolve,
                    error: function (response, status, error) {
                        if (response.responseText === undefined) {
                            toast('Web Server Error',
                                'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
                        } else {
                            toast('Browser Error', error+': '+response.responseText, 'error', 15000);
                        }
                        reject(); // this doesn't seem to work right.
                    }
                });
            });
        },
        open: function(filename) {
            // Add findFile is a folder was selected.
            filename = filename.match(/\/$|\\$/) ? filename+DirSep+findFile : filename;
            $(target).val(filename);
            $('.ui-widget-overlay').siblings('.ui-dialog').find('.ui-dialog-content').dialog('close');
            // Bring up the save changes button.
            findPendingChanges();
        },
        item_class: function(_, name) {
            return name.match(/^[A-Z]:|^\/$/) ? 'drive' : '';
        },
    });

};