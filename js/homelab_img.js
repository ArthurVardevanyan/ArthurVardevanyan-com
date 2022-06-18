jQuery;
// Get the modal https://www.w3schools.com/howto/howto_css_modal_images.asp
var modal = document.getElementById("modal1");
// Get the image and insert it inside the modal - use its "alt" text as a caption
var img = document.getElementById("img1");
var modalImg = document.getElementById("img01");
var captionText = document.getElementById("caption1");
img.onclick = function () {
  modal.style.display = "block";
  modalImg.src = this.src;
  captionText.innerHTML = this.alt;
};

// Get the <span> element that closes the modal
var span = modalImg;

// When the user clicks on <span> (x), close the modal
span.onclick = function () {
  modal.style.display = "none";
};

// Get the modal https://www.w3schools.com/howto/howto_css_modal_images.asp
var modal = document.getElementById("modal2");
// Get the image and insert it inside the modal - use its "alt" text as a caption
var img = document.getElementById("img2");
var modalImg = document.getElementById("img02");
var captionText = document.getElementById("caption2");
img.onclick = function () {
  modal.style.display = "block";
  modalImg.src = this.src;
  captionText.innerHTML = this.alt;
};

// Get the <span> element that closes the modal
var span = modalImg;

// When the user clicks on <span> (x), close the modal
span.onclick = function () {
  modal.style.display = "none";
};

var coll = document.getElementsByClassName("collapsible");
var i;

for (i = 0; i < coll.length; i++) {
  coll[i].addEventListener("click", function () {
    this.classList.toggle("active");
    var content = this.nextElementSibling;
    if (content.style.maxHeight) {
      content.style.maxHeight = null;
    } else {
      content.style.maxHeight = window.innerHeight - 150 + "px";
      content.style.overflow = "auto";
    }
  });
}
(() => {
  "use strict";
  // Page is loaded
  const objects = document.getElementsByClassName("asyncImage");
  Array.from(objects).map((item) => {
    // Start loading image
    const img = new Image();
    img.src = item.dataset.src;
    // Once image is loaded replace the src of the HTML element
    img.onload = () => {
      item.classList.remove("asyncImage");
      return item.nodeName === "IMG"
        ? (item.src = item.dataset.src)
        : (item.style.backgroundImage = `url(${item.dataset.src})`);
    };
  });
})();
