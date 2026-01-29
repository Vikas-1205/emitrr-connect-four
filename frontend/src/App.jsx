import { useState, useEffect } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import LobbyScreen from './components/LobbyScreen';
import WaitingScreen from './components/WaitingScreen';
import GameBoard from './components/GameBoard';
import GameOverModal from './components/GameOverModal';
import Leaderboard from './components/Leaderboard';
import MoveTimer from './components/MoveTimer';
import ThemeToggle from './components/ThemeToggle';
import SoundToggle from './components/SoundToggle';
import GameHistory from './components/GameHistory';
import { createConfetti } from './utils/confetti';
import { soundManager } from './utils/sounds';
import { getAvatarForUsername, avatarStyles } from './utils/avatars';

export default function App() {
    const {
        isConnected,
        status,
        gameState,
        playerNum,
        message,
        lastMove,
        joinQueue,
        makeMove,
        resetGame,
        requestRematch,
        username,
        difficulty,
        setDifficulty,
        lastOpponent
    } = useWebSocket();

    const [showLeaderboard, setShowLeaderboard] = useState(false);
    const [selectedDifficulty, setSelectedDifficulty] = useState('medium');

    // Trigger confetti on win
    useEffect(() => {
        if (status === 'game_over' && gameState?.winner === playerNum) {
            createConfetti();
        }
    }, [status, gameState?.winner, playerNum]);

    const handlePlayAgain = () => {
        resetGame();
        if (username) {
            joinQueue(username, selectedDifficulty);
        }
    };

    const handleRematch = () => {
        requestRematch();
    };

    const handleGoHome = () => {
        resetGame();
        setShowLeaderboard(false);
    };

    const handleDifficultyChange = (diff) => {
        setSelectedDifficulty(diff);
        setDifficulty(diff);
    };

    const handleJoinGame = (name) => {
        joinQueue(name, selectedDifficulty);
    };

    // Create empty board if no game state
    const board = gameState?.board || Array(6).fill(null).map(() => Array(7).fill(0));
    const isMyTurn = gameState?.currentPlayer === playerNum;
    const isGameActive = status === 'playing';

    // Get player info
    const player1Name = gameState?.player1Username || 'Player 1';
    const player2Name = gameState?.player2Username || 'Player 2';
    const player1Avatar = getAvatarForUsername(player1Name);
    const player2Avatar = getAvatarForUsername(player2Name);

    return (
        <>
            {/* Header */}
            <header className="header">
                <h1>4 in a Row</h1>
                <p>Connect Four • Real-time Multiplayer</p>
                <div className="header-controls">
                    <SoundToggle />
                    <ThemeToggle />
                </div>
            </header>

            <div className="container">
                {/* Show leaderboard */}
                {showLeaderboard && status !== 'playing' && (
                    <div className="leaderboard-container">
                        <Leaderboard onClose={() => setShowLeaderboard(false)} />
                        <GameHistory />
                    </div>
                )}

                {/* Lobby - Enter username */}
                {(status === 'connected' || status === 'connecting' || status === 'disconnected') && !showLeaderboard && (
                    <>
                        <LobbyScreen
                            onJoinGame={handleJoinGame}
                            isConnected={isConnected}
                            defaultUsername={username}
                            difficulty={selectedDifficulty}
                            onDifficultyChange={handleDifficultyChange}
                        />
                        <div style={{ textAlign: 'center', marginTop: '1rem' }}>
                            <button
                                className="btn btn-secondary"
                                onClick={() => setShowLeaderboard(true)}
                            >
                                🏆 View Leaderboard
                            </button>
                        </div>
                    </>
                )}

                {/* Waiting for opponent */}
                {status === 'waiting' && (
                    <WaitingScreen message={message} difficulty={selectedDifficulty} />
                )}

                {/* Game in progress */}
                {(status === 'playing' || status === 'game_over') && gameState && (
                    <div className="game-layout-vertical">
                        {/* Timer */}
                        <MoveTimer
                            isMyTurn={isMyTurn}
                            isGameActive={isGameActive}
                        />

                        {/* Player 2 (Opponent) - Top */}
                        <div className={`player-bar ${gameState.currentPlayer === 2 && isGameActive ? 'active' : ''}`}>
                            <div style={avatarStyles.container(gameState.isBotGame ? '#6c63ff' : player2Avatar.color, 48)}>
                                {gameState.isBotGame ? '🤖' : player2Avatar.emoji}
                            </div>
                            <div className="player-bar-info">
                                <span className="player-bar-name">
                                    {player2Name}
                                    {playerNum === 2 && <span className="you-badge">You</span>}
                                    {gameState.isBotGame && <span className="bot-badge">Bot</span>}
                                </span>
                                <span className="player-bar-disc yellow"></span>
                            </div>
                            {gameState.currentPlayer === 2 && isGameActive && (
                                <span className="turn-badge">Playing...</span>
                            )}
                        </div>

                        {/* Game Board */}
                        <div className="game-board-wrapper">
                            <GameBoard
                                board={board}
                                currentPlayer={gameState.currentPlayer}
                                playerNum={playerNum}
                                onColumnClick={makeMove}
                                gameState={gameState}
                                lastMove={lastMove}
                            />
                        </div>

                        {/* Player 1 (You or First Player) - Bottom */}
                        <div className={`player-bar ${gameState.currentPlayer === 1 && isGameActive ? 'active' : ''}`}>
                            <div style={avatarStyles.container(player1Avatar.color, 48)}>
                                {player1Avatar.emoji}
                            </div>
                            <div className="player-bar-info">
                                <span className="player-bar-name">
                                    {player1Name}
                                    {playerNum === 1 && <span className="you-badge">You</span>}
                                </span>
                                <span className="player-bar-disc red"></span>
                            </div>
                            {gameState.currentPlayer === 1 && isGameActive && (
                                <span className="turn-badge">Playing...</span>
                            )}
                        </div>
                    </div>
                )}

                {/* Game Over Modal */}
                {status === 'game_over' && gameState && (
                    <GameOverModal
                        gameState={gameState}
                        playerNum={playerNum}
                        onPlayAgain={handlePlayAgain}
                        onGoHome={handleGoHome}
                        onRematch={lastOpponent && !gameState.isBotGame ? handleRematch : null}
                    />
                )}
            </div>

            {/* Connection Status */}
            <div className={`connection-status ${isConnected ? 'connected' : 'disconnected'}`}>
                <span className="dot"></span>
                {isConnected ? 'Connected' : 'Reconnecting...'}
            </div>
        </>
    );
}
