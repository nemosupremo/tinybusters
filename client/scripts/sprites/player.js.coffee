class tiny.sprites.player
  constructor: (@stage, @world, @camera, opts) ->

    @keyStatus =
      left: false
      right: false
      up: false
      down: false
      lmb: false

    @boost = 100
    @booster = false

    headShape = new p2.Rectangle(15 / @camera.ppu, 10 / @camera.ppu)
    headShape.collisionaGroup = 1
    headShape.collisionMask = 1 | 2

    bodyShape = new p2.Rectangle(30 / @camera.ppu, 35 / @camera.ppu)
    bodyShape.collisionGroup = 1
    bodyShape.collisionMask = 1 | 2

    footShape = new p2.Rectangle(22 / @camera.ppu, 10 / @camera.ppu)
    footShape.collisionGroup = 1
    footShape.collisionMask = 1 | 2
    footShape.isFoot = true

    @p2Body = new p2.Body({ mass:1, position:[opts?.x || 5, opts?.y || 15], fixedRotation:true});

    @p2Body.addShape(bodyShape);
    @p2Body.addShape(headShape, [0,(22 / @camera.ppu)]);
    @p2Body.addShape(footShape, [0,-(22 / @camera.ppu)]);

    @world.addBody(@p2Body);

    $(@stage.canvas).on("keydown", @keyDown)
    $(@stage.canvas).on("keyup", @keyUp)

    @world.on("postStep", @onPostStep)
    @world.on("impact", @onImpact)

  onPostStep: () =>
    dVel = 0
    if @keyStatus.left
      dVel = -5
    else if @keyStatus.right
      dVel = 5
    velChange = dVel - @p2Body.velocity[0];
    force = @p2Body.mass * velChange / (1/60);
    @p2Body.force[0] += force

    if @jumpSteps > 0
      force = @p2Body.mass * 7 / (1/60.0);
      force /= 6.0;
      @jumpSteps--;
      @p2Body.force[1] += force
    else if @keyStatus.up
      if @boost
        @p2Body.force[1] += 40
        @boost -= 5
        @boost = Math.max(0, @boost)
    else if @boost <= 100
      @boost = Math.min(100, @boost+3)

  onImpact: (evt) =>
    #console.log evt
    bodyA = evt.bodyA;
    bodyB = evt.bodyB;
    playerBody = null
    playerShape = null
    otherBody = null
    otherShape = null
    if bodyA.id == @p2Body.id
      playerBody = bodyA
      playerShape = evt.shapeA
      otherBody = bodyB
    else if bodyB.id == @p2Body.id
      playerBody = bodyB
      playerShape = evt.shapeB
      otherBody = bodyA
    console.log playerShape
    #console.log
    if playerBody? && (evt.shapeA.isFoot || evt.shapeB.isFoot)
      console.log "caNjump1"
      @canJump = true

  keyDown: (event) =>
    switch event.which
      when 65 # LEFT
        @keyStatus.left = true
      when 68 # RIGHT
        @keyStatus.right = true
      when 87 # JUMP
        @keyStatus.up = true
        if @canJump
          @jumpSteps = 6
          @canJump = false
        else
          console.log Math.abs(@p2Body.velocity[1])

  keyUp: (event) =>
    switch event.which
      when 65 # LEFT
        @keyStatus.left = false
      when 68 # RIGHT
        @keyStatus.right = false
      when 87
        @keyStatus.up = false

  destroy: () =>
    $(@stage.canvas).off("keydown", @keyDown)
    $(@stage.canvas).off("keyup", @keyUp)

    @world.off("postStep", @PostStep)
    @world.off("impact", @onImpact)

    @world.removeBody(@p2Body)
