class tiny.ng.service.tinychat
  @$inject: ['$rootScope', 'tinysocket']

  constructor: (@rootScope, @tinysocket) ->
    _.extend(@, Backbone.Events)
    @tinysocket.on("data:#{tiny.msg.CHAT}", @inScope(@onChatMsg))

  inScope: (f) =>
    return () =>
      args = arguments
      @rootScope.$apply(() =>
        f.apply(this, args);
      )

  sendMessage: (message) =>
    @tinysocket.sendData(
      _t: tiny.msg.CHAT,
      m: message
    )

  postMessage: (data) =>
    d =
      s: data.s || data.server || data.type == "server"
      n: data.n || data.name
      m: data.m || data.message
      e: data.e || data.error
    @onChatMsg(d)

  onChatMsg: (data) =>
    evts = ["message"]
    if data.s
      evts.push("message:server")
    else
      evts.push("message:user")
    if data.e
      evts.push("message:error")
    evts = evts.join(" ")
    @trigger(evts, data)
