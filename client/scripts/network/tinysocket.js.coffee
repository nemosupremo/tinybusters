class tiny.ng.service.tinysocket
  @$inject: []

  constructor: () ->
    @status = null
    _.extend(@, Backbone.Events)

  makeid: (n) ->
    text = "";
    possible = "0123456789";
    for i in [0..n-1]
      text += possible.charAt(Math.floor(Math.random() * possible.length));
    return text;

  connect: (server, username, password, register) =>
    if window["WebSocket"]
      parts = []
      if !server? || server == ""
        server = "localhost:9001"
      if !username? || username == ""
        username = "buster#{@makeid(8)}"
      parts.push("name=#{encodeURIComponent(username)}")
      if password?
        parts.push("pass=#{encodeURIComponent(password)}")
      if register? and register
        parts.push("register=1")
      query = parts.join("&")
      @conn = new WebSocket("ws://#{server}/connect?#{query}", tiny.const.WS_PROTOCOL);
      @conn.binaryType = "arraybuffer";
      @conn.onopen = @onOpen
      @conn.onoerror = @onError
      @conn.onclose = @onClose
      @conn.onmessage = @onMessage

  sendData: (data) =>
    @conn.send(msgpack.encode(data))

  send: (message) =>
    @conn.send(message)

  onOpen: (evt) =>
    @trigger("open", evt)

  onClose: (evt) =>
    @trigger("close", evt)

  onError: (evt) =>
    @trigger("error", evt)

  onMessage: (evt) =>
    if evt.data instanceof ArrayBuffer
      data = msgpack.decode(evt.data)
      @trigger("data data:#{data._t}", data, evt)
    else
      @trigger("message", evt.data, evt)

