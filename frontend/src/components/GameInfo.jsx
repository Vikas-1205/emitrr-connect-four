import { getAvatarForUsername, avatarStyles } from '../utils/avatars';

export default function GameInfo({ gameState, playerNum, message }) {
    if (!gameState) return null;

    const player1 = gameState.player1Username || 'Player 1';
    const player2 = gameState.player2Username || 'Player 2';
    const isPlayer1Turn = gameState.currentPlayer === 1;
    const isGameActive = gameState.status === 'in_progress';

    const player1Avatar = getAvatarForUsername(player1);
    const player2Avatar = getAvatarForUsername(player2);

    return (
        <div className="game-info">
            {/* Player 1 */}
            <div className={`player-info ${isPlayer1Turn && isGameActive ? 'active' : ''}`}>
                <div className="player-header">
                    <div style={avatarStyles.container(player1Avatar.color, 36)}>
                        {player1Avatar.emoji}
                    </div>
                    <div className="name">
                        <span className="disc red"></span>
                        {player1}
                        {playerNum === 1 && <span className="you-badge">You</span>}
                    </div>
                </div>
                <div className="status">
                    {isGameActive && isPlayer1Turn ? '🎯 Playing...' : 'Waiting...'}
                </div>
            </div>

            {/* VS Divider */}
            <div className="vs-divider">VS</div>

            {/* Player 2 */}
            <div className={`player-info ${!isPlayer1Turn && isGameActive ? 'active' : ''}`}>
                <div className="player-header">
                    <div style={avatarStyles.container(player2Avatar.color, 36)}>
                        {gameState.isBotGame ? '🤖' : player2Avatar.emoji}
                    </div>
                    <div className="name">
                        <span className="disc yellow"></span>
                        {player2}
                        {playerNum === 2 && <span className="you-badge">You</span>}
                        {gameState.isBotGame && <span className="bot-badge">Bot</span>}
                    </div>
                </div>
                <div className="status">
                    {isGameActive && !isPlayer1Turn ? '🎯 Playing...' : 'Waiting...'}
                </div>
            </div>

            {/* Message */}
            {message && (
                <div className="game-message">
                    {message}
                </div>
            )}

            {/* Game Rules */}
            <div className="game-rules">
                <h4>How to Play</h4>
                <ul>
                    <li>Click a column to drop your disc</li>
                    <li>Connect 4 in a row to win</li>
                    <li>Horizontal, vertical, or diagonal</li>
                </ul>
            </div>
        </div>
    );
}
