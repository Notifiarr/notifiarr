// https://vjdesign.com.au/animated-accordion-without-jquery/ 

(function () {
  var accordions, contents, i;
  
  // Make sure the browser supports what we are about to do.
  
  var supports = !!document.querySelector && !!window.addEventListener;

  
  if ( !supports ) {
    return;
  } else {
    // Using a function helps isolate each accordion from the others
    function makeAccordion(accordion) {
      var targets, currentTarget, i;

      targets = accordion.querySelectorAll('.accordion dt');
      contents = accordion.querySelectorAll('.accordion dd');

      for(i = 0; i < contents.length; i++) {
        if(!(contents[i].getAttribute('aria-expanded') === 'true')){
          contents[i].setAttribute('aria-expanded','false');
        }
      }
      
      for(i = 0; i < targets.length; i++) {
        if(targets[i].nextElementSibling.getAttribute('aria-expanded') === 'true'){
          targets[i].style.borderBottomLeftRadius = 0;
          targets[i].style.borderBottomRightRadius = 0;
        }  
        
        targets[i].addEventListener('click', function () {


          if(this.nextElementSibling.getAttribute('aria-expanded') === 'true'){
            this.nextElementSibling.setAttribute('aria-expanded','false');
             this.style.borderRadius = "0.5em";            
          } else {  
            this.nextElementSibling.setAttribute('aria-expanded','true');
            this.style.borderBottomLeftRadius = 0;
            this.style.borderBottomRightRadius = 0;
          }  
        }, false);
      }

      accordion.classList.add('js');
    }

    // Find all the accordions to enable
    accordions = document.querySelectorAll('.accordion');
    for(i = 0; i < accordions.length; i++) {
      makeAccordion(accordions[i]);
    }
  }   
  
})();