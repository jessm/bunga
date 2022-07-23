import React, { useState, useEffect, useRef } from 'react'
import { useLocation } from 'react-router-dom'

import { sendCommand } from './utils'
import Nav from "./nav"
import LobbyInfo from "./lobbyInfo"
import Bunga from "./bunga"

const Lobby = () => {
  let location = useLocation()
  const user = location.state.user
  const lobby = location.state.lobby

  // Save lobby state in the component
  const [lobbyState, setLobbyState] = useState({
    Players: [],
    Status: "lobby",
    Scores: {},
  })

  const [gameState, setGameState] = useState(null)
  const wsRef = useRef(null)

  const handleQuit = (quitWsRef) => {
    sendCommand(quitWsRef.current, 'lobby', 'backToLobby')
  }

  // Create a websocket on component mount
  // don't cleanup until component is unmounted
  useEffect(() => {
    // create socket
    const host = window.location.host
    const wsUri = encodeURI(`wss://${host}/joinLobby?user=${user}&lobby=${lobby}`)
    wsRef.current = new WebSocket(wsUri)
    wsRef.current.onopen = () => {
      // console.log('Connected!')
    }
    wsRef.current.onmessage = (event) => {
      let msg = JSON.parse(event.data)
      let newState = msg.State
      if (msg.Target == 'lobby') {
        // console.log('new lobby state:', newState)
        setLobbyState(newState)
        setGameState(null)
      } else if (msg.Target == 'game') {
        // console.log('new game state:', newState)
        setGameState(newState)
      }
    }

    return () => {
      wsRef.current.close()
    }
  }, [])

  return (
    <>
      <Nav lobbyState={lobbyState.Status} handleQuit={() => handleQuit(wsRef)}/>
      {lobbyState.Status == "lobby" &&
        <LobbyInfo user={user} lobby={lobby} lobbyState={lobbyState} ws={wsRef.current} />
      }
      {lobbyState.Status == "game" &&
        <Bunga user={user} lobbyState={lobbyState} gameState={gameState} ws={wsRef.current} />
      }
    </>
  )
}

export default Lobby
