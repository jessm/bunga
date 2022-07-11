import React from 'react'
import { Link } from 'react-router-dom'

const Nav = (props) => {
  return (
    <nav className="level is-mobile has-background-primary mb-0">
      <div className="level-left">
        <Link to="/" className="card-header-title">Bunga</Link>
        { props.showHelp &&
          <div className="level-item">
            <Link to="/help" className="button is-small">Help</Link>
          </div>
        }
        {
          props.lobbyState == "game" &&
          <div className="level-item">
            <button className="button is-small is-danger" onClick={props.handleQuit}>Back to lobby</button>
          </div>
        }
      </div>
    </nav>
  )
}

export default Nav
