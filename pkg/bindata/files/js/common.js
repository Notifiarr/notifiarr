$(document).ready((function() {
    jsLoader();
    setTooltips();

    $(document).bind('change keyup mouseup', '.client-parameter', function(){
        findPendingChanges();
    });

    // ----- Navbar
    $('.nav-link').click((function() {
        $('nav.ts-sidebar').toggleClass('menu-open', false);
    }))

    $(".menu-btn").click((function() {
        $('nav.ts-sidebar').toggleClass('menu-open');
    }))
}));

// ---------------------------------------------------------------------------------------------

function jsLoader()
{
    let path        = '';
    let script      = '';
    const files     = ['navigation', 'golists', 'fileViewer', 'services', 'triggers'];

    for (const file of files) {
        path        = 'files/js/' + file + '.js';
        script      = document.createElement('script');
        script.src  = path;
        document.head.appendChild(script);
    }
}
// -------------------------------------------------------------------------------------------

function ajax(url, method, type)
{
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

function setTooltips()
{
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

function findPendingChanges()
{
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
            changes += titleCaseWord(group) +': '+ label +'<br>';
        }
    });

    if (changes) {
        $('#pending-change-list').html(changes);
        $('#pending-change-counter').html(counter);
        $('#pending-change-container').show();
    }
}
// ---------------------------------------------------------------------------------------------

function savePendingChanges()
{
    let fields = '';

    $.each($('.client-parameter'), function() {
        const id = $(this).attr('id')
        if (id !== undefined) {
            fields += '&'+ $(this).attr('id') +'='+ $(this).val();
        }
    });

    $.ajax({
        type: 'POST',
        url: 'reconfig',
        data: fields,
        success: function (data){
            $('#pending-change-container').remove();          // remove save button.
            setTimeout(function(){location.reload();}, 5000); // reload window in 5 seconds.
            toast('Config Saved', 'Wait 5 seconds; reloading the new configuration...', 'success');
        },
        error: function (response, status, error) {
            if (status === undefined) {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Save Error', error+': '+response.responseText, 'error', 15000);
            }
        }
    });
}

function saveProfileChanges()
{
    let fields = '';

    $.each($('.profile-parameter'), function() {
        const id = $(this).attr('id')
        if (id !== undefined) {
            fields += '&'+ $(this).attr('id') +'='+ $(this).val();
        }
    });

    $.ajax({
        type: 'POST',
        url: 'profile',
        data: fields,
        success: function (data){
            $('#current-username').html($('#NewUsername').val()); // update the html username.
            toast('Profile Saved', data, 'success');
        },
        error: function (response, status, error) {
            if (error == "") {
                toast('Web Server Error',
                    'Notifiarr client appears to be down! Hard refresh recommended.', 'error', 30000);
            } else {
                toast('Save Error', error+': '+response.responseText, 'error', 15000);
            }
        }
    });
}
// ---------------------------------------------------------------------------------------------

function getCharacterLength (str)
{
    return [...str].length;
}
// ---------------------------------------------------------------------------------------------

function toast(title, message, type, duration=5000)
{
    $.Toast(title, message, type, {
        has_icon: true,
        has_close_btn: true,
        stack: true,
        fullscreen: false,
        timeout: duration,
        sticky: false,
        has_progress: true,
        rtl: false,
    });
}
// -------------------------------------------------------------------------------------------

function titleCaseWord(word)
{
    return word.charAt(0).toUpperCase() + word.slice(1);
}
// -------------------------------------------------------------------------------------------

function lowercaseWord(word)
{
    return word.toLowerCase();
}
// -------------------------------------------------------------------------------------------

function reindexList(group)
{
    group = group.split('-');
    group = group[0];

    let counter = 0;
    let uiCounter = 1;
    let currentIndex = 0;
    let app = '';

    $.each($('[data-group='+ group +']'), function(index) {
        switch (group) {
            case 'starr':
                const id = $(this).attr('id'); // Apps.<app>.<index>.<field>

                if (id) {
                    const parts = id.split('.');

                    if (parts[1] != app) {
                        counter = 0;
                        currentIndex = 0;
                    }

                    if (parts[2] != currentIndex) {
                        counter++;
                    }

                    //-- CHANGE id ATTR ON INPUT FIELDS
                    $(this).attr('id', parts[0] +'.'+ parts[1] +'.'+ counter +'.'+ parts[3]);

                    //-- CHANGE data-label ON INPUT FIELDS
                    const dataLabel = $(this).attr('data-label');
                    $(this).attr('data-label', dataLabel.replace(dataLabel.replace(/[^.\d]/g, ''), (counter + 1)));

                    //-- CHANGE id ATTR ON LABEL
                    $('#'+ lowercaseWord(parts[1]) +'-index-label-'+ parts[2]).attr('id', lowercaseWord(parts[1]) +'-index-label-'+ counter).html(counter + 1);

                    //-- SET CURRENT VARIABLES
                    app = parts[1];
                    currentIndex = parts[2];
                }
                break;
        }
    });
}
// -------------------------------------------------------------------------------------------
