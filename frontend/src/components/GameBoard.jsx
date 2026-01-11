import { useState, useEffect } from 'react';
import './GameBoard.css';

const GameBoard = ({ socket, gameID, player, turn, onMove, winner, boardState }) => {
  const [grid, setGrid] = useState(Array(6).fill(null).map(() => Array(7).fill(null)));

  useEffect(() => {
    if (boardState) {
      setGrid(boardState);
    }
  }, [boardState]);

  const handleColumnClick = (colRel) => {
    if (winner || turn !== player) return;
    onMove(colRel);
  };

  return (
    <div className="game-board">
      {grid.map((row, rIndex) => (
        <div key={rIndex} className="row">
          {row.map((cell, cIndex) => (
            <div 
              key={cIndex} 
              className={`cell ${cell ? cell : ''}`}
              onClick={() => handleColumnClick(cIndex)}
            >
              {cell && <div className={`disc disc-${cell}`} />}
            </div>
          ))}
        </div>
      ))}
      {winner && <div className="winner-banner">Winner: {winner}</div>}
    </div>
  );
};

export default GameBoard;
