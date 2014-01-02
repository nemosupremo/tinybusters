class tiny.ng.tinyApp
  @$inject: ['$httpProvider']

  constructor: (@httpProvider) ->
    @httpProvider.defaults.useXDomain = true;
    @httpProvider.defaults.withCredentials = true;
    delete @httpProvider.defaults.headers.common['X-Requested-With'];

class tiny.ng.ctrl.nav
  @$inject: ['$scope']

  constructor: (@scope) ->
    @isFullscreen = false
    setFs = () =>
      @scope.$apply( () =>
        @isFullscreen = document.fullScreen || document.mozFullScreen || document.webkitIsFullScreen || document.msFullscreenEnabled;;
      )
    $("body")[0].addEventListener('fullscreeneventchange', setFs, true);
    document.addEventListener('mozfullscreenchange', setFs, true);
    document.addEventListener('MSFullscreenChange', setFs, true);
    document.addEventListener('webkitfullscreenchange', setFs, true);

  fullscreen: () ->
    fullscreenEnabled = document.fullScreen || document.mozFullScreen || document.webkitIsFullScreen || document.msFullscreenEnabled;;

    if fullscreenEnabled
      if document.exitFullscreen
        document.exitFullscreen()
      else if document.mozExitFullScreen
        document.mozExitFullScreen()
      else if document.webkitExitFullscreen()
        document.webkitExitFullscreen
      else if document.msExitFullscreen
        document.msExitFullscreen
    else
      element = $("body")[0]
      if element.requestFullscreen
        element.requestFullscreen()
      else if element.mozRequestFullScreen
        element.mozRequestFullScreen()
      else if element.webkitRequestFullscreen
        element.webkitRequestFullscreen()
      else if element.msRequestFullscreen
        element.msRequestFullscreen

class tiny.ng.ctrl.game
  @$inject: ['$scope', 'tinysocket', 'tinycore', 'tinychat', '$element']

  constructor: (@scope, @tinysocket, @tinycore, @tinychat, @element) ->
    @scope.chat = []
    @scope.server = "TINYBUSTERS"
    @scope.location = "TITLE SCREEN"
    @scope.connected = false
    @scope.chattxt = ""
    @scope.sendMessage = @sendMessage

    if $(@element).prop("fillwindow") || $(@element).attr("fillwindow")?
      @resizeView()
      $( window ).resize( =>
        @resizeView();
      );

    @tinycore.attach($(@element).find("#gameport")[0])
    @tinychat.on("message", @onChat)
    @tinysocket.on("open", @inScope(() => @scope.connected = true))
    @tinysocket.on("close", @inScope(() =>
      if @scope.connected
        @tinychat.postMessage(
          type:"server",
          error:true,
          message:"Disconnected from server."
        )
      @scope.connected = false))

  inScope: (f) =>
    return () =>
      args = arguments
      @scope.$apply(() =>
        f.apply(this, args);
      )

  onChat: (data) =>
    if data.s
      @scope.chat.push(
        type: "server",
        message: data.m
        error: !!data.e
      )
    else
      @scope.chat.push(
        type: "user",
        name: data.n
        message: data.m
        error: !!data.e
      )

    $(@element).find(".chatlist").scrollTop($(@element).find(".chatlist").height());

  sendMessage: () =>
    @tinychat.sendMessage(@scope.chattxt)
    @scope.chattxt = ""

  resizeView: =>
    find = $(@element).find.bind($(@element))
    tnH = $(".topnav").height()
    bodyHeight = window.innerHeight - tnH
    chatWidth = find(".chat").width()
    portWidth = find(".main").width()

    w = Math.floor(portWidth - portWidth*.01)
    h = Math.floor(bodyHeight - bodyHeight*.01)
    find('#gameport').css({
      "width": w,
      "height": h,
    });

    find(".chat").css({
      "height": h,
    })
    find('#gameport')[0].width = w
    find('#gameport')[0].height = h

construct = (constructor) ->
  F = (args) ->
    return constructor.apply(this, args);
  F.prototype = constructor.prototype;

  g = () ->
    return new F(arguments);
  g.$inject = constructor.$inject
  return g

tiny.ng.app = angular.module('tinybusters', []);
tiny.ng.app.controller("tiny.busterctrl", construct(tiny.ng.ctrl.game))
tiny.ng.app.controller("tiny.busternav", construct(tiny.ng.ctrl.nav))

tiny.ng.app.service('tinysocket', construct(tiny.ng.service.tinysocket))
tiny.ng.app.service('tinychat', construct(tiny.ng.service.tinychat))
tiny.ng.app.service('tinycore', construct(tiny.busters))

tiny.ng.app.directive "ngEnter", ->
  (scope, element, attrs) ->
    element.bind "keydown keypress", (event) ->
      if event.which is 13
        scope.$apply ->
          scope.$eval attrs.ngEnter
        event.preventDefault()

tiny.ng.app.config construct(tiny.ng.tinyApp)