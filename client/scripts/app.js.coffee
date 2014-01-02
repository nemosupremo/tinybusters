class tiny.ng.tinyApp
  @$inject: ['$httpProvider']

  constructor: (@httpProvider) ->
    @httpProvider.defaults.useXDomain = true;
    @httpProvider.defaults.withCredentials = true;
    delete @httpProvider.defaults.headers.common['X-Requested-With'];

class tiny.ng.ctrl.game
  @$inject: ['$scope', 'tinysocket', 'tinycore', '$element']

  constructor: (@scope, @tinysocket, @tinycore, @element) ->
    @scope.chat = []
    @scope.server = "TINYBUSTERS"
    @scope.location = "TITLE SCREEN"
    @scope.connected = false
    @scope.chattxt = ""
    @scope.sendMessage = @sendMessage

    @resizeView()
    $( window ).resize( =>
      @resizeView();
    );

    @tinycore.attach($(@element).find("#gameport")[0])
    @tinysocket.on("data:#{tiny.msg.CHAT}", @inScope(@onChat))
    @tinysocket.on("open", @inScope(() => @scope.connected = true))
    @tinysocket.on("close", @inScope(() =>
      if @scope.connected
        @scope.chat.push(
          type:"server",
          message:"Disconnected from server."
        )
      @scope.connected = false))

    #shouldn't auto connect here
    @tinysocket.connect()

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
      )
    else
      @scope.chat.push(
        type: "user",
        name: data.n
        message: data.m
      )

  sendMessage: () =>
    @tinysocket.sendData(
      _t: tiny.msg.CHAT,
      m: @scope.chattxt
    )
    @scope.chattxt = ""

  resizeView: =>
    find = $(@element).find.bind($(@element))
    tnH = find(".topnav").height()
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
tiny.ng.app.controller("tiny.tinyctrl", construct(tiny.ng.ctrl.game))

tiny.ng.app.service('tinysocket', construct(tiny.ng.service.tinysocket))
tiny.ng.app.service('tinycore', construct(tiny.busters))

tiny.ng.app.directive "ngEnter", ->
  (scope, element, attrs) ->
    element.bind "keydown keypress", (event) ->
      if event.which is 13
        scope.$apply ->
          scope.$eval attrs.ngEnter
        event.preventDefault()

tiny.ng.app.config construct(tiny.ng.tinyApp)