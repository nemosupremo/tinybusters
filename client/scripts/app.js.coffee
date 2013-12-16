window.tiny = {
  scenes: {}
}

window.resizeView = () ->
  tnH = $(".topnav").height()
  bodyHeight = window.innerHeight - tnH
  chatWidth = $(".chat").width()
  portWidth = $(".main").width()

  w = Math.floor(portWidth - portWidth*.01)
  h = Math.floor(bodyHeight - bodyHeight*.01)
  $('#gameport').css({
    "width": w,
    "height": h,
  });

  $(".chat").css({
    "height": h,
  })
  $('#gameport')[0].width = w
  $('#gameport')[0].height = h


$(document).ready(() ->
  window.resizeView()
  $( window ).resize( ->
    window.resizeView();
  );
  window.tinybusters = new tiny.busters($('#gameport')[0])
);


