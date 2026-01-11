import { useState, useEffect, useRef } from 'react'
import GameBoard from './components/GameBoard'
import './App.css'

function App() {
  const [username, setUsername] = useState('')
  const [connected, setConnected] = useState(false)
  const [gameID, setGameID] = useState(null)
  const [playerCode, setPlayerCode] = useState(null) // P1 or P2
  const [turn, setTurn] = useState(null)
  const [board, setBoard] = useState(Array(6).fill(null).map(() => Array(7).fill(null)))
  const [winner, setWinner] = useState(null)
  const [status, setStatus] = useState('Enter username to join')
  const [leaderboard, setLeaderboard] = useState([])

  const socketRef = useRef(null)


  const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081'
  const WS_URL = API_URL.replace(/^http/, 'ws')

  useEffect(() => {
    fetchLeaderboard()
  }, [])

  const fetchLeaderboard = async () => {
    try {
      const res = await fetch(`${API_URL}/api/leaderboard`)
      const data = await res.json()
      setLeaderboard(data)
    } catch (e) {
      console.error(e)
    }
  }

  const joinGame = () => {
    if (!username) return alert('Please enter a username')

    // Connect WebSocket
    const socket = new WebSocket(`${WS_URL}/ws?username=${username}`)
    socketRef.current = socket

    socket.onopen = () => {
      setConnected(true)
      setStatus('Waiting for opponent...')
    }

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      console.log('Received:', msg)

      if (msg.type === 'start') {
        setGameID(msg.gameID)
        setPlayerCode(msg.you)
        setTurn(msg.turn)
        setStatus(`Game Started! You are ${msg.you} vs ${msg.opponent}`)
      } else if (msg.type === 'update') {
        // Update board
        setBoard(prev => {
          const newBoard = prev.map(row => [...row])
          newBoard[msg.row][msg.col] = msg.player
          return newBoard
        })
        setTurn(msg.turn)
        if (msg.winner) {
          setWinner(msg.winner)
          setStatus(`Game Over! Winner: ${msg.winner}`)
          fetchLeaderboard() // Refresh stats
        }
      }
    }

    socket.onclose = () => {
      setConnected(false)
      setStatus('Disconnected')
    }
  }

  const sendMove = (col) => {
    if (socketRef.current && turn === playerCode) {
      socketRef.current.send(JSON.stringify({ type: 'move', col }))
    }
  }

  return (
    <div className="App">
      <h1>4 in a Row</h1>

      {!connected ? (
        <div className="login">
          <input
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <button onClick={joinGame}>Join Game</button>
        </div>
      ) : (
        <div className="game-container">
          <div className="status">{status}</div>
          {playerCode && (
            <div className="info">
              You are: <strong>{username}</strong> <span className={`badge ${playerCode}`}>{playerCode}</span>
              {' '}- Turn: <span className={`badge ${turn}`}>{turn}</span>
            </div>
          )}

          <GameBoard
            boardState={board}
            player={playerCode}
            turn={turn}
            onMove={sendMove} // Pass the function to handle moves
            winner={winner}
          />
        </div>
      )}

      <div className="leaderboard">
        <h2>Leaderboard</h2>
        <ul>
          {leaderboard.map((u, i) => (
            <li key={i}>{u.username}: {u.wins} Wins</li>
          ))}
        </ul>
      </div>
    </div>
  )
}

export default App
