import React, { useState, useEffect, useRef } from 'react'
import * as PIXI from 'pixi.js'

import { sendCommand, lerp } from "./utils"

const hlColours = {
  "p": 0x00ffff,
  "s": 0xffff00,
  "b": 0xff0000,
}
const tableColour = 0x35654d
const tableLight = 0x34914a

const animTime = 20 // frames

var cardWidth = 60
var cardHeight = 84
var uiPadd = 30
var screenHeight = 0
var screenWidth = 0
var curUser = ''
var playerOrder = []

// Layout, aspect ratio ~= 1.5, e.g. 375 x 567
// CardHeight = 0.15 * screenHeight
// CardWidth = 0.106 * screenHeight
// Padding = 0.0177 * screenHeight
// UiPadd = 0.075 * screenHeight
// Highlight border = 0.5 * padding
// 5 hands * (cardHeight + padding) = 0.839
// 

const getHandPositions = (viewWidth, handSize) => {
  let ret = []
  const eachCardSpace = (viewWidth - 20 - cardWidth) / (handSize - 1)
  for (let i = 0; i < handSize; i++) {
    ret.push(10 + eachCardSpace * i)
  }
  return ret
}

// order is the player order array
const getOtherHandY = (screenHeight, order, user, otherPlayer) => {
  if (user == otherPlayer) {
    return getPlayerHandY(screenHeight)
  }
  let ownIdx = order.findIndex((element) => element == user)
  let numFromUser = 0
  let otherIdx = ownIdx
  while (order[otherIdx] != otherPlayer) {
    numFromUser++
    otherIdx = (otherIdx + 1) % order.length
  }
  return 10 + (cardHeight + uiPadd) * (numFromUser - 1) // offset by 1 so numFromUser = 1 => at top
}

const getPlayerHandY = (screenHeight) => {
  return screenHeight - cardHeight - uiPadd
}

const getDrawPileY = (screenHeight) => {
  return screenHeight / 2 - 0.5 * cardHeight
}

const getCardSprite = (texRef, card) => {
  let ret = new PIXI.Container()
  let sprite = new PIXI.Sprite(texRef.current[card.slice(0, 2)])
  if (card.length > 2) {
    let graphics = new PIXI.Graphics()
    let hlColour = hlColours[card[2]]
    graphics.beginFill(hlColour)
    graphics.drawRoundedRect(-5, -5, cardWidth + 10, cardHeight + 10, 5)
    graphics.endFill()
    ret.addChild(graphics)
  }
  sprite.width = cardWidth
  sprite.height = cardHeight
  ret.addChild(sprite)
  ret.interactive = true
  return ret
}

const getCardPos = (player, idx, gameState) => {
  let ret = {
    'x': getHandPositions(screenWidth, gameState.PlayerHands[player].length)[idx],
    'y': 0,
  }
  if (player == curUser) {
    ret['y'] = getPlayerHandY(screenHeight)
  } else {
    ret['y'] = getOtherHandY(screenHeight, playerOrder, curUser, player)
  }
  return ret
}

const getActionPositions = (action, gameState) => {
  let ret = { 'Start': { 'x': 0, 'y': 0 }, 'End': { 'x': 0, 'y': 0 } }
  let drawDiscardPositions = getHandPositions(screenWidth, 4)
  let xPosMap = {
    'draw': drawDiscardPositions[1],
    'discard': drawDiscardPositions[2],
  }
  Object.keys(ret).forEach(pos => {
    if (action[pos] == 'draw' || action[pos] == 'discard') {
      ret[pos]['x'] = xPosMap[action[pos]]
      ret[pos]['y'] = getDrawPileY(screenHeight)
    } else {
      ret[pos] = getCardPos(action[pos], action[pos + 'Idx'], gameState)
    }
  })
  return ret
}

const handleAnim = (texRef, animRef, appRef, actionRef, actions, gameState) => {
  // actions is [{Start: ..., StartIdx: ..., End: ..., EndIdx: ... }...]
  // animRef.current is {id: {func: f, state: num, sprite: s}, ...}
  if (actions == null || actions.length == 0) {
    return
  }
  // clean up current animations
  Object.entries(animRef.current).forEach(entry => {
    const [id, data] = entry
    appRef.current.ticker.remove(data['func'], id)
    delete animRef.current[id]
  })
  // add new animations
  actions.forEach((action) => {
    let card = action["Card"] == "" ? "1B" : action["Card"]
    let sprite = getCardSprite(texRef, card)
    // get positions based on start and index
    let poses = getActionPositions(action, gameState)
    let id = actionRef.current
    let animFunc = (delta) => {
      if (!Object.hasOwn(animRef.current, id) || sprite == null) {
        appRef.current.ticker.remove(animFunc, id)
        return
      }
      let frac = (animRef.current[id]['state'] + delta) / animTime
      animRef.current[id]['state'] += delta
      try {
        sprite.x = lerp(poses['Start']['x'], poses['End']['x'], frac)
        sprite.y = lerp(poses['Start']['y'], poses['End']['y'], frac)
      } catch {
        frac = 1
      }
      if (frac >= 1) {
        appRef.current.ticker.remove(animFunc, id)
        animRef.current[id]['sprite'].destroy()
        delete animRef.current[id]
      }
    }
    animRef.current[id] = {
      'func': animFunc,
      'state': 0,
      'sprite': sprite,
    }
    appRef.current.stage.addChild(sprite)
    appRef.current.ticker.add(animFunc, id)
    actionRef.current = (actionRef.current + 1) % 1000
  })
}

const Bunga = (props) => {
  if (props.gameState == null) {
    return (
      <p>Loading...</p>
    )
  }

  const appRef = useRef(null)
  const cardTexturesRef = useRef(null)
  const [resourcesLoaded, setResourcesLoaded] = useState(false)
  const actionRef = useRef(0)
  const animRef = useRef({})

  // Setup pixijs
  useEffect(() => {
    let cards = []
    let cardNums = ['2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A']
    let suits = ['C', 'D', 'H', 'S']
    cardNums.forEach((name) => {
      suits.forEach((suit) => cards.push(name + suit))
    })
    cards.push('1B')
    cards.push('2B')
    cards.push('X')
    cardTexturesRef.current = {}
    PIXI.utils.clearTextureCache()
    const loader = new PIXI.Loader()
    cards.forEach((card) => {
      let cardUrl = 'assets/cards/' + card + '.svg'
      loader.add(card, cardUrl)
    })
    loader.load((loader, resources) => {
      // console.log('loaded!')
      cards.forEach((card) => {
        cardTexturesRef.current[card] = resources[card].texture
      })
      setResourcesLoaded(true)
    })
    let root = document.getElementById('pixiRoot')
    let width = root.offsetWidth
    let height = root.offsetHeight
    appRef.current = new PIXI.Application({
      width: width,
      height: height,
      backgroundColor: tableColour,
      resolution: 1,
    })
    appRef.current.ticker.maxFPS = 60
    root.appendChild(appRef.current.view)
    // Set up heights and widths of things as global variables
    screenHeight = appRef.current.screen.height
    screenWidth = appRef.current.screen.width
    cardHeight = 0.15 * screenHeight
    cardWidth = 0.106 * screenHeight
    uiPadd = 0.5 * cardHeight
    return () => {
      // console.log('deleting pixijs app')
      appRef.current.destroy()
      appRef.current = null
    }
  }, [])

  let redrawBoard = () => {
    if (!resourcesLoaded) {
      // console.log('spinning')
      let spinner = new PIXI.Container()
      let g = new PIXI.Graphics()
      g.lineStyle(10, 0xffffff, 1)
      g.beginFill(tableColour)
      g.drawCircle(0, 0, cardWidth)
      g.endFill()
      spinner.addChild(g)
      g = new PIXI.Graphics()
      g.beginFill(tableColour)
      g.drawRect(cardWidth / 2 + 20, -20, 25, 40)
      spinner.addChild(g)
      spinner.x = screenWidth / 2
      spinner.y = screenHeight / 2
      let spinnerAnimFunc = (delta) => {
        try {
          spinner.rotation += 0.075 * delta
        } catch {
          appRef.current.ticker.remove(spinnerAnimFunc)
        }
      }
      appRef.current.ticker.add(spinnerAnimFunc)
      appRef.current.stage.addChild(spinner)
      return
    }
    // console.log('redrawing cards')
    appRef.current.stage.destroy({ children: true })
    appRef.current.stage = new PIXI.Container()
    playerOrder = props.gameState.PlayerOrder
    curUser = props.user

    // draw highlight areas around hands
    Object.keys(props.gameState.PlayerHands).forEach(player => {
      let c = new PIXI.Container()
      let g = new PIXI.Graphics()
      if (props.gameState.Turn == player) {
        g.beginFill(tableLight)
      } else {
        g.beginFill(tableColour)
      }
      g.drawRoundedRect(0, 0, screenWidth, cardHeight + uiPadd + 10, 5)
      g.endFill()
      c.addChild(g)
      const nameStyle = { fontFamily: 'Arial', fontSize: 16, fill: 0xffffff }
      let nameText = new PIXI.Text(player, nameStyle)
      nameText.anchor.set(0.5, 0)
      nameText.x = screenWidth / 2
      nameText.y = cardHeight + 15
      c.addChild(nameText)
      c.x = 0
      c.y = getOtherHandY(screenHeight, props.gameState.PlayerOrder, props.user, player) - 10
      appRef.current.stage.addChild(c)
    })

    // redraw bunga button only if it's your turn and noone said bunga yet and it's StartTurn
    if (props.gameState.Turn == props.user && props.gameState.SaidBunga == "" &&
      props.gameState.PlayingState == "startTurn") {
      let bungaContainer = new PIXI.Container()
      let textBlock = new PIXI.Graphics()
      textBlock.beginFill(hlColours["p"])
      textBlock.drawRoundedRect(0, 0, cardWidth, cardHeight, 5)
      textBlock.endFill()
      bungaContainer.addChild(textBlock)
      const bungaStyle = {
        fontFamily: 'Arial',
        fontSize: '36',
      }
      let bungaText = new PIXI.Text('Bunga!', bungaStyle)
      bungaText.anchor.set(0.5, 0.5)
      bungaText.x = cardWidth / 2
      bungaText.y = cardHeight / 2
      bungaContainer.addChild(bungaText)
      bungaContainer.interactive = true
      bungaContainer.on('pointerdown', () => {
        // console.log('bunga pressed')
        sendCommand(props.ws, 'game', 'bunga', { 'player': props.user })
      })
      bungaContainer.x = getHandPositions(screenWidth, 4)[0]
      bungaContainer.y = getDrawPileY(screenHeight)
      appRef.current.stage.addChild(bungaContainer)
    }

    // redraw deck
    let card = props.gameState.DrawPile
    if (card == '') {
      card = '1B'
    }
    let sprite = getCardSprite(cardTexturesRef, card)
    sprite.x = getHandPositions(screenWidth, 4)[1]
    sprite.y = getDrawPileY(screenHeight)
    sprite.on('pointerdown', () => {
      // console.log('drawing card...')
      sendCommand(props.ws, 'game', 'draw', { 'player': props.user })
    })
    appRef.current.stage.addChild(sprite)

    // redraw discard
    card = props.gameState.DiscardPile
    sprite = getCardSprite(cardTexturesRef, card)
    sprite.x = getHandPositions(screenWidth, 4)[2]
    sprite.y = getDrawPileY(screenHeight)
    sprite.on('pointerdown', () => {
      // console.log('discard pile clicked')
      sendCommand(props.ws, 'game', 'discard', { 'player': props.user })
    })
    appRef.current.stage.addChild(sprite)

    // redraw players' hands
    Object.entries(props.gameState.PlayerHands).forEach(entry => {
      const [player, hand] = entry
      // console.log('drawing hand for', player, 'hand: ', hand)
      let handPositions = getHandPositions(screenWidth, hand.length)
      hand.forEach((card, idx) => {
        let handCard = getCardSprite(cardTexturesRef, card)
        handCard.on('pointerdown', () => {
          // console.log('sending cmd card', props.user, idx)
          sendCommand(props.ws, 'game', 'card', {
            'index': idx.toString(),
            'owner': player,
            'player': props.user
          })
        })
        handCard.x = handPositions[idx]
        if (player == props.user) {
          handCard.y = getPlayerHandY(screenHeight)
        } else {
          handCard.y = getOtherHandY(screenHeight, props.gameState.PlayerOrder, props.user, player)
        }
        appRef.current.stage.addChild(handCard)
      })
    })

    // handle drawing finishes if 'turn' == 'final'
    if (props.gameState.Turn == "final") {
      Object.entries(props.gameState.Scores).forEach(entry => {
        const [player, score] = entry
        let scoreContainer = new PIXI.Container()
        let textBlock = new PIXI.Graphics()
        textBlock.beginFill(hlColours["p"])
        textBlock.drawRoundedRect(0, 0, cardHeight * 2, 0.5 * cardWidth, 5)
        textBlock.endFill()
        scoreContainer.addChild(textBlock)
        const scoreStyle = {
          fontFamily: 'Arial',
          fontSize: 20,
        }
        let scoreText = new PIXI.Text('Score: ' + score, scoreStyle)
        scoreText.anchor.set(0.5, 0.5)
        scoreText.x = 2 * cardHeight / 2
        scoreText.y = 0.5 * cardWidth / 2
        scoreContainer.addChild(scoreText)
        scoreContainer.x = 0.5 * screenWidth
        if (player == props.user) {
          scoreContainer.y = getPlayerHandY(screenHeight)
        } else {
          scoreContainer.y = getOtherHandY(screenHeight, props.gameState.PlayerOrder, props.user, player)
        }
        scoreContainer.y += 0.5 * cardHeight
        // console.log("drawing score container player", player, scoreContainer.y, "user:", props.user)
        scoreContainer.pivot.x = scoreContainer.width / 2
        scoreContainer.pivot.y = scoreContainer.height / 2
        appRef.current.stage.addChild(scoreContainer)
      })
      let winContainer = new PIXI.Container()
      let textBlock = new PIXI.Graphics()
      textBlock.beginFill(hlColours["s"])
      textBlock.drawRoundedRect(0, 0, 3 * cardHeight, cardWidth, 5)
      textBlock.endFill()
      winContainer.addChild(textBlock)
      const winStyle = {
        fontFamily: 'Arial',
        fontSize: 24,
      }
      let winText = new PIXI.Text('Winner: ' + props.gameState.Winner + "!", winStyle)
      winText.anchor.set(0.5, 0.5)
      winContainer.addChild(winText)
      winText.x = 3 * cardHeight / 2
      winText.y = cardWidth / 2
      winContainer.pivot.x = winContainer.width / 2
      winContainer.pivot.y = winContainer.height / 2
      winContainer.x = screenWidth / 2
      winContainer.y = screenHeight / 2
      appRef.current.stage.addChild(winContainer)
    }
  }

  useEffect(() => {
    // handle animations
    handleAnim(cardTexturesRef, animRef, appRef, actionRef, props.gameState.LatestAction, props.gameState)
    if (props.gameState.LatestAction != null && props.gameState.LatestAction.length > 0) {
      setTimeout(redrawBoard, animTime * (1/60) * 1000)
    } else {
      redrawBoard()
    }
  }, [props.gameState, resourcesLoaded])

  return (
    <>
      <div className="level restheight">
        <div id="pixiRoot" className="level-item fullheight"></div>
      </div>
    </>
  )
}

export default Bunga
