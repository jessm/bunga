import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import Nav from "./nav"

const Home = () => {
  const [formName, setFormName] = useState("")
  const [formLobby, setFormLobby] = useState("")
  const [formError, setFormError] = useState("")
  const navigate = useNavigate()

  const checkValid = async (e) => {
    e.preventDefault()
    let lobbyCode = ""
    if (formLobby != "") {
      const resp = await fetch("/valid", {
        method: "POST",
        body: JSON.stringify({name: formName, lobby: formLobby})
      })
      if (!resp.ok) {
        // Display warning
        // TODO: change the backend to say if the lobby or username is bad instead of an empty error
        // and reflect that in the message
        setFormError("on")
        return
      }
      lobbyCode = formLobby
    } else {
      const resp = await fetch("/newLobby").then(response => response.json())
      if (resp.name.length == 4) {
        lobbyCode = resp.name
      } else {
        throw "Error getting new lobby name"
      }
    }
    navigate("/" + lobbyCode, { state: {"user": formName, "lobby": lobbyCode }})
  }

  return (
    <>
      <Nav showHelp={true} />
      <section className="section">
        <div className="card">
          <header className="card-header">
            <p className="card-header-title">Join or create lobby</p>
          </header>
          <div className="card-content">
            <form onSubmit={e => checkValid(e)}>
              <div className="field">
                <label className="label">Name</label>
                <div className="control">
                  <input
                    className="input"
                    name="username"
                    type="text"
                    placeholder="name"
                    onChange={e => {
                      setFormName(e.target.value)
                      setFormError("")
                    }}
                  ></input>
                </div>
              </div>
              <div className="field">
                <label className="label">Lobby</label>
                <div className="control">
                  <input
                    className="input"
                    name="lobby"
                    type="text"
                    placeholder="abcd"
                    onChange={e => {
                      setFormLobby(e.target.value)
                      setFormError("")
                    }}
                  ></input>
                </div>
              </div>
              { formError != "" &&
                <div className="notification is-danger">Invalid lobby or username</div>
              }
              <div className="field">
                <div className="control">
                  { formLobby == "" &&
                    <button className="button is-success" type="submit">Create lobby</button>
                    ||
                    <button className="button is-link" type="submit">Join lobby</button>
                  }
                </div>
              </div>
            </form>
          </div>
        </div>
      </section>
    </>
  )
}

export default Home
