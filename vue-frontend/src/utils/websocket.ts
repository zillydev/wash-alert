function send(socket: WebSocket, type: string, message: any) {
  socket.send(JSON.stringify({ "Type": type, "Message": message }))
}

export { send }