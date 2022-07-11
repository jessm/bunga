import React, { useState } from 'react'
import * as ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, useLocation } from 'react-router-dom'

import Help from "./help"
import Home from "./home"
import Lobby from "./lobby"

const App = () => {
  return (
    <>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/help" element={<Help />} />
          <Route path="/:lobby" element={<Lobby />} />
        </Routes>
      </BrowserRouter>
    </>
  )
}

document.addEventListener('DOMContentLoaded', () => {
  const root = ReactDOM.createRoot(document.getElementById('root'))
  root.render(
    <App />
  )
})
