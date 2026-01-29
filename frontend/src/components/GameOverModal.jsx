export default function GameOverModal({
    gameState,
    playerNum,
    onPlayAgain,
    onGoHome,
    onRematch
}) {
    const winner = gameState?.winner;
    const isWin = winner === playerNum;
    const isDraw = winner === 0;

    let resultClass = 'draw';
    let title = "It's a Draw!";
    let emoji = '🤝';

    if (!isDraw) {
        if (isWin) {
            resultClass = 'win';
            title = 'You Won!';
            emoji = '🎉';
        } else {
            resultClass = 'lose';
            title = 'You Lost';
            emoji = '😔';
        }
    }

    return (
        <div className="modal-overlay">
            <div className={`modal ${resultClass}`}>
                <div className="modal-emoji">{emoji}</div>
                <h2>{title}</h2>
                <p>
                    {isDraw
                        ? 'Great game! The board is full with no winner.'
                        : isWin
                            ? 'Congratulations! You connected 4 in a row!'
                            : 'Better luck next time! Keep practicing!'
                    }
                </p>

                <div className="modal-stats">
                    <div className="stat">
                        <span className="stat-value">{gameState?.moveCount || 0}</span>
                        <span className="stat-label">Moves</span>
                    </div>
                    <div className="stat">
                        <span className="stat-value">{gameState?.isBotGame ? '🤖' : '👤'}</span>
                        <span className="stat-label">Opponent</span>
                    </div>
                </div>

                <div className="modal-actions">
                    {onRematch && (
                        <button className="btn btn-secondary" onClick={onRematch}>
                            🔄 Rematch
                        </button>
                    )}
                    <button className="btn btn-primary" onClick={onPlayAgain}>
                        🎮 Play Again
                    </button>
                    <button className="btn btn-secondary" onClick={onGoHome}>
                        🏠 Home
                    </button>
                </div>
            </div>
        </div>
    );
}
