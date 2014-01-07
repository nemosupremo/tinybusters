window.tiny = {
  scenes: {}
  sprites: {}
  ng:
    mod: {}
    service: {}
    ctrl: {}
  const:
    WS_PROTOCOL: ["tinybusters-v1"]
  msg:
    INVALID: 0
    CHAT: 1
}

class tiny.busters
  @$inject: ['$http', 'tinysocket', 'tinychat']

  constructor: (@http, @tinysocket, @tinychat) ->
    @serverInfo = @getServerInfo()

  postError: (err) =>
    @tinychat.postMessage(
      type:"server",
      error:true,
      message:err
    )

  getServerInfo: (serverList) ->
    dfd = new jQuery.Deferred();
    $http = @http
    findServer = (servers) =>
      loadInfo = (server) ->
        $http({method: 'GET', url: "http://#{server.hostname}:#{server.port}/info"})
      success = (data) ->
        dfd.resolve(data)
      error = () ->
        if servers.length > 0
          servers.shift()
          findServer(servers)
        else
          @postError("Failed to find any valid seed servers.")
          dfd.reject()
      if servers.length > 0
        loadInfo(servers[0]).success(success).error(error)
      else
        @postError("Failed to find any valid seed servers.")
        dfd.reject()
    @http({method: 'GET', url: '/config.json'})
    .success((data) =>
      if data.seed_servers.length > 0
        findServer(data.seed_servers)
      else
        @postError("Failed to find seed server list.")
        dfd.reject()
    ).error((data) =>
      @postError("Failed to get seed server list.")
      dfd.reject()
    )

    return dfd.promise();

  attach: (@canvas) ->
    createjs.Ticker.setFPS(60);
    createjs.Ticker.timingMode = createjs.Ticker.RAF_SYNCHED

    @stage = new createjs.Stage(@canvas);
    createjs.Ticker.addEventListener("tick", @stage);

    @scene = new tiny.scenes.title(@stage)
    @fps = new createjs.Text("0fps", "300 18px Helvetica", "#333333");
    @fps.x = 25
    @fps.y = 25

    @stage.addChild(@fps)
    createjs.Ticker.addEventListener("tick", @updateFPS);

    $.when( @serverInfo ).done((server) =>
      @tinysocket.connect("#{server.hostname}:#{server.port}")
    )

    _.delay((() =>
      @scene.exit()
      @scene = new tiny.scenes.game(@stage)
    ), 1000)

    $(@canvas).focus()
    #@stage.addChild(@scene.container)
    #@stage.update();

  updateFPS: () =>
    @fps.text = numeral(createjs.Ticker.getMeasuredFPS()).format('0.00') + "fps"
