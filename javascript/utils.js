export const lerp = (a, b, t) => {
  return (1-t)*a + t*b;
}

export const sendCommand = (ws, target, cmd, args) => {
  ws.send(JSON.stringify({
    "target": target,
    "cmd": cmd,
    "args": args,
  }))
}
