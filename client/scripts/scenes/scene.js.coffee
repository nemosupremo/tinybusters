class tiny.scenes.scene

  constructor: (@stage) ->
    @scene = new createjs.Container()
    @stage.addChild(@scene)

  exit: () =>
    @stage.removeChild(@scene)
    @scene = null
