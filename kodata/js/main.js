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
  if (window.location.hash === "#emailSent") {
    alert("Email Sent!");
  }
});
$(function () {
  if (window.location.hash === "#emailFailed") {
    alert("Captcha Failed!");
  }
});

$(document).ready(function () {
  $(window).scroll(function () {
    if ($(this).scrollTop() > 100) {
      $("#scroll").fadeIn();
    } else {
      $("#scroll").fadeOut();
    }
  });
  $("#scroll").click(function () {
    $("html, body").animate({ scrollTop: 0 }, 600);
    return false;
  });
});
