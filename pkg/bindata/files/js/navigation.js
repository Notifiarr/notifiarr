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
