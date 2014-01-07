class tiny.scenes.game extends tiny.scenes.scene
  constructor: () ->
    super
    @camera =
      ppu: 32
      focus: null

    world = new p2.World();
    world.gravity[1] = -20

    planeShape = new p2.Plane();
    planeShape.collisonGroup = 2
    planeShape.collisionMask = 1 | 2
    planeBody = new p2.Body({position:[0, 0]});
    planeBody.addShape(planeShape);
    world.addBody(planeBody);

    randShape = new p2.Rectangle(6,3)
    randShape.collisonGroup = 2
    randShape.collisionMask = 1 | 2
    randBody = new p2.Body({position:[7, 4], mass: 0});
    randBody.addShape(randShape);
    world.addBody(randBody);

    randShape = new p2.Rectangle(6,3)
    randShape.collisonGroup = 2
    randShape.collisionMask = 1 | 2
    randBody = new p2.Body({position:[11, 4], mass: 0, angle: Math.PI*1/4});
    randBody.addShape(randShape);
    world.addBody(randBody);

    @worldStep = setInterval((() -> world.step(1/60)), 16.67)

    player = new tiny.sprites.player(@stage, world, @camera)

    @updatePositions = () =>
      @simpleRender(world)

    @simpleRender(world, true)

    createjs.Ticker.addEventListener("tick", @updatePositions)

  p22camera: (position) =>
    if @camera.focus?
      console.log("yeesh")
    else
      x = position[0]*@camera.ppu
      y = position[1]*@camera.ppu*-1 + @stage.canvas.height
      return [x,y]

  simpleRender: (world, x) =>
    ppu = @camera.ppu
    for body in world.bodies
      for idx, shape of body.shapes
        offset = [body.shapeOffsets[idx][0], body.shapeOffsets[idx][1]*-1]
        angle = body.shapeAngles[idx]
        unless shape.easel?
          if shape instanceof p2.Circle
            g = new createjs.Graphics()
            g.setStrokeStyle(1)
            .beginStroke(createjs.Graphics.getRGB(0,0,0))
            .drawCircle(offset[0]*ppu, offset[1]*ppu, shape.radius*ppu)
            .moveTo(offset[0]*ppu, offset[1]*ppu)
            .lineTo(
              ppu*(offset[0]+shape.radius*Math.cos(body.angle)),
              ppu*(offset[1]+shape.radius*Math.sin(body.angle))
            )
            shape.easel = new createjs.Shape(g);
          else if shape instanceof p2.Rectangle
            g = new createjs.Graphics()
            g.setStrokeStyle(1)
            .beginStroke(createjs.Graphics.getRGB(0,0,0))
            .drawRect(
              ppu*(offset[0]),
              ppu*(offset[1]),
              ppu*shape.width,
              ppu*shape.height)
            shape.easel = new createjs.Shape(g);
            shape.easel.regX = ppu*shape.width/2;
            shape.easel.regY = ppu*shape.height/2;
          else if shape instanceof p2.Plane
            g = new createjs.Graphics()
            g.setStrokeStyle(1)
            .beginStroke(createjs.Graphics.getRGB(0,0,0))
            .moveTo(-10*ppu,0)
            .lineTo(10*ppu,0)
            shape.easel = new createjs.Shape(g);
          if shape.easel?
            @scene.addChild(shape.easel)
        pos = @p22camera(body.position)
        shape.easel.rotation = body.angle * (180/Math.PI) * -1
        shape.easel.x = pos[0]
        shape.easel.y = pos[1]
    return

  exit: () =>
    clearInterval(@worldStep)
    createjs.Ticker.removeEventListener("tick", @updatePositions)
    super