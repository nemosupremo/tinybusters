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
  sendMessage = ->
    message = $(".txt-chat").val()
    message =
      _t: 1,
      s: false,
      m: message
    message = msgpack.encode(message)
    window.conn.send(message)
    $(".txt-chat").val("")

  $(".btn-chat").click(sendMessage)
  $(".txt-chat").keypress (e) ->
      if e.which == 13
        sendMessage()
        e.preventDefault()
);

makeid = (n) ->
    text = "";
    possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    for i in [0..n-1]
      text += possible.charAt(Math.floor(Math.random() * possible.length));
    return text;

if window["WebSocket"]
  template =
    chat: Handlebars.compile($("#chatTemplate").html());
    server: Handlebars.compile($("#serverMessageTemplate").html());

  window.conn = new WebSocket("ws://localhost:9001/connect?name=#{makeid(8)}");
  conn.binaryType = "arraybuffer";
  conn.onopen = () ->
    $(".chatlist").empty()
  conn.onclose = (evt) ->
    $(".chatlist").append(template.server(message: "Server connection closed."))
  conn.onmessage = (evt) ->
    mp = msgpack.decode(evt.data)
    if mp.s
      $(".chatlist").append(template.server(message: mp.m))
    else
      $(".chatlist").append(template.chat(name: mp.n, message: mp.m))
