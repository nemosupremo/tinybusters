window.tiny = {
  scenes: {}
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
    findServer = (servers) ->
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
    @http({method: 'GET', url: '/server.json'})
    .success((data) =>
      if data.length > 0
        findServer(data)
      else
        @postError("Failed to find seed server list.")
        dfd.reject()
    ).error((data) ->
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


    #@stage.addChild(@scene.container)
    #@stage.update();

  updateFPS: () =>
    @fps.text = numeral(createjs.Ticker.getMeasuredFPS()).format('0.00') + "fps"

class tiny.scenes.title

    constructor: (@stage) ->
      @container = new createjs.Container()
      @loadingContainer = new createjs.Container()
      @assetsLoaded = false

      @title = new createjs.Text("TINY BUSTERS", "700 92px Raleway", "#333333");
      @title.x = 25
      @title.y = @stage.canvas.height / 2 - @title.getMeasuredLineHeight() / 2

      @loadingBar = new createjs.Shape();
      @lbH = 15
      @lbW = 600

      @loadingBar.graphics.beginFill("#333").drawRect(0, 0, 1, @lbH).endFill();

      @lbFrame = new createjs.Shape();
      @lbFrame.graphics.setStrokeStyle(1).beginStroke("#333").drawRect(-3/2, -3/2, @lbW+3, @lbH+3).endStroke();

      @loadingContainer.addChild(@loadingBar, @lbFrame);
      @loadingContainer.x = 50
      @loadingContainer.y = @title.y + @title.getMeasuredLineHeight() + 10;

      @container.addChild(@title)
      @stage.addChild(@container)
      @stage.addChild(@loadingContainer)

      createjs.Ticker.addEventListener("tick", @onFrame);
      @preload()

    preload: () ->
      preload = new createjs.LoadQueue(false);
      preload.addEventListener("complete",  () => @onLoaded(preload));
      preload.addEventListener("progress", () => @loadingBar.scaleX = preload.progress * @lbW;);
      preload.loadFile({id: "scenes.title.bg", src:"/assets/title/titlebg.png"});

    onLoaded: (preload) ->
      bg = new createjs.Bitmap(preload.getResult("scenes.title.bg"));
      width = Math.ceil(@stage.canvas.width / bg.getBounds().width)+1
      height = Math.ceil(@stage.canvas.height / bg.getBounds().height)+1
      @bg = []
      @bgContainer = new createjs.Container()
      for x in [0..(width-1)]
        for y in [0..(height-1)]
          bg = new createjs.Bitmap(preload.getResult("scenes.title.bg"));
          bg.x = (x-1)*bg.getBounds().width
          bg.y = y*bg.getBounds().height
          @bgContainer.addChild(bg)
          @bg.push(bg)
      @stage.addChildAt(@bgContainer, 0)
      @loaded = true

    onFrame: () =>
      @title.y = @stage.canvas.height / 2 - @title.getMeasuredLineHeight() / 1.125
      @loadingContainer.x = 55
      @loadingContainer.y = @title.y + @title.getMeasuredLineHeight() + 10;
      if @loaded
        snp = _.map(@bg, (x) -> [x.x, x.y])
        for bg in @bg
          bg.x += 1
          bg.y -= 1
          if bg.x > @stage.canvas.width
            bg.x = _.min(snp, (x) -> x[0])[0] - bg.getBounds().width
          if bg.y < -bg.getBounds().height
            bg.y = _.max(snp, (x) -> x[1])[1] + bg.getBounds().height
