import React from 'react'

import { sendCommand } from "./utils"

const LobbyInfo = (props) => {
  const handleStartGame = () => {
    sendCommand(props.ws, "lobby", "startGame")
  }

  return (
    <div className="card restheight">
      <header className="card-header">
        <p className="card-header-title">Lobby code: {props.lobby}</p>
      </header>
      <div className="card-content">
        <div className="field">
          <div className="control">
            <button className="button is-primary" onClick={handleStartGame}>Start Game</button>
          </div>
        </div>
        <div className="content">
          <nav className="panel">
            <div className="panel-heading">
              <p>Players</p>
            </div>
            {
              Object.keys(props.lobbyState.Players).map((player) => {
                return (
                  <div key={player} className="panel-block">
                    <div className="control">{player}</div>
                  </div>
                )
              })
            }
          </nav>
        </div>
      </div>
    </div>
  )
}

export default LobbyInfo
