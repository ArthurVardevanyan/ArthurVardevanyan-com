/*
(function ($) {

  "use strict";


  $('.navigation').singlePageNav({
    currentClass: 'active'
  });


  $('.toggle-menu').click(function () {
    $('.responsive-menu').stop(true, true).slideToggle();
    return false;
  });
}
*/
$(function () {
  if (window.location.hash === "#emailsent") {
    alert("Email Sent!");
  }
});

$(document).ready(function(){ 
  $(window).scroll(function(){ 
      if ($(this).scrollTop() > 100) { 
          $('#scroll').fadeIn(); 
      } else { 
          $('#scroll').fadeOut(); 
      } 
  }); 
  $('#scroll').click(function(){ 
      $("html, body").animate({ scrollTop: 0 }, 600); 
      return false; 
  }); 
});