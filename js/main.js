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
