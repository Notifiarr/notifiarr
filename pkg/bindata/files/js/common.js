function jsLoader() {
    let path        = '';
    let script      = '';
    const files     = ['navigation'];

    for (const file of files){
        path        = 'files/js/' + file + '.js';
        script      = document.createElement('script');
        script.src  = path;
        document.head.appendChild(script);
    }
}
// -------------------------------------------------------------------------------------------

function ajax(url, method, type) {
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

function setTooltips() {
    $('[class*="balloon-tooltip"]').hide();

    $('a, div, i, img, input, span, td, button').balloon({
        position: 'bottom',
        classname: 'balloon-tooltip',
        css: {
            fontSize: '18px',
            borderRadius: '12px'
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

function findPendingChanges() {
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
            changes += group.charAt(0).toUpperCase() + group.slice(1) +': '+ label +'<br>';
        }
    });

    if (changes) {
        $('#pending-change-list').html(changes);
        $('#pending-change-counter').html(counter);
        $('#pending-change-container').show();
    }
}
// ---------------------------------------------------------------------------------------------

$(document).ready((function() {
// ----- Navbar
    $(".nav-link").click((function() {
      $("nav.ts-sidebar").toggleClass("menu-open", false)
    }))

    $(".menu-btn").click((function() {
        $("nav.ts-sidebar").toggleClass("menu-open")
    }))

// ----- Log File Display
    $('#LogFileSelector').change(function(){
      var logFileID = $(this).val();
      var filename = $("#fileinfo-"+logFileID).data('filename')
      var used = $("#fileinfo-"+logFileID).data('used')
      var lineCount = 500

      // start a spinner because this takes a few seconds.
      $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading content for file '+filename+' ...');

      $("[id^=fileinfo-]").hide() // hide all other file-info divs.
      $("#fileinfo-"+logFileID).show() // show the file info div for the file requested.
      $('#log-file-actions').hide(); // hide actions (until file load completes).
      $('#logHelpMsg').hide() // hide help msg (this is the active file warning tooltip).
      $('#log-file-ajax-error').hide();  // hide any error messages.

      if (used == true) {
        $('#logHelpMsg').show() // display the help tooltip for this file.
      }

      $.ajax({
        url: "getLog/"+logFileID+"/"+lineCount+"/0", // the zero is optional, skip counter.
        context: document.body,
        success:function(data) {
          $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+lineCount+'</span> log lines.');
          $('#log-file-actions').show()
          $('#log-file-content').text($.trim(data));
          updatePreCounters($('#log-file-content'))
        },
        error: function (request, status, error) {
          $('#log-file-ajax-error').show();
          $('#log-file-ajax-error').html("<h4>"+error+"</h4>\n"+request.responseText)
          // $('#log-file-content').html("An error occurred getting the file contents:\n"+error+"\n"+request.responseText);
        },
      });
  });

    $('#triggerLogAction').click(function(){
        var action = $('#logfileAction').val();
        var logFileID = $('#LogFileSelector').val();

        if (action == "download") {
            window.location.href = "downloadLog/"+logFileID;
        } else if (action == "delete") {
            
        }

    });

    $('#triggerLogLoad').click(function(){
      var logFileID = $('#LogFileSelector').val();
      var filename = $("#fileinfo-"+logFileID).data('filename');
      var used = $("#fileinfo-"+logFileID).data('used');
      var lineCount = parseInt($('#logLinesCount').val());
      var lineAction = $('#logLinesAction').val(); // add/reload
      var offSetCount = parseInt($('#logsCurrentLineCount').html());
      var totalLines = lineCount+offSetCount;

      $('#log-file-ajax-error').hide();

      // make go button spin?

      if (lineAction == "linesAdd") {
        $.ajax({
          url: "getLog/"+logFileID+"/"+lineCount+"/"+offSetCount,
          context: document.body,
          success:function(data) {
            $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+totalLines+'</span> log lines.');
            $('#log-file-content').prepend(document.createTextNode(data));
            updatePreCounters($('#log-file-content'))
          },
          error: function (request, status, error) {
            $('#log-file-ajax-error').show();
            $('#log-file-ajax-error').html("<h4>"+error+"</h4>\n"+request.responseText)
          },
        });
      } else {
        // start a spinner because this takes a few seconds.
        $('#log-file-content').html('<i class="fas fa-cog fa-spin fa-2x"></i> Loading content for file '+filename+' ...');

        $.ajax({
          url: "getLog/"+logFileID+"/"+lineCount+"/0",
          context: document.body,
          success:function(data) {
            $('#log-file-action-msg').html('Displaying last <span id="logsCurrentLineCount">'+lineCount+'</span> log lines.');
            $('#log-file-actions').show()
            $('#log-file-content').text($.trim(data));
            updatePreCounters($('#log-file-content'))
          },
          error: function (request, status, error) {
            $('#log-file-ajax-error').show();
            $('#log-file-ajax-error').html("<h4>"+error+"</h4>\n"+request.responseText)
          },
        });
      }
    });
    updatePreCounters($('pre'))
}));

function updatePreCounters(tags) {
    var pl = tags.length
    for (var i = 0; i < pl; i++) {
        tags[i].innerHTML = '<span class="line-number"></span>' +
          tags[i].innerHTML + '<span class="cl"></span>';
        var num = tags[i].innerHTML.split(/\n/).length;
        for (var j = 0; j < num; j++) {
            var line_num = tags[i].getElementsByTagName('span')[0];
            line_num.innerHTML += '<span>' + (j + 1) + '</span>';
        }
    }
}
