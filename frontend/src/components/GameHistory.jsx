import { useState, useEffect } from 'react';
import { API_URL } from '../utils/config';

export default function GameHistory() {
    const [games, setGames] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchGames();
    }, []);

    const fetchGames = async () => {
        try {
            const response = await fetch(`${API_URL}/api/games/history`);
            if (response.ok) {
                const data = await response.json();
                setGames(data.games || []);
            }
        } catch (error) {
            console.error('Failed to fetch game history:', error);
        } finally {
            setLoading(false);
        }
    };

    const formatDuration = (seconds) => {
        if (!seconds) return '-';
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return mins > 0 ? `${mins}m ${secs}s` : `${secs}s`;
    };

    const formatTime = (timestamp) => {
        if (!timestamp) return '-';
        const date = new Date(timestamp);
        return date.toLocaleString();
    };

    if (loading) {
        return (
            <div className="game-history">
                <h3>📜 Recent Games</h3>
                <p style={{ textAlign: 'center', color: 'var(--text-secondary)' }}>Loading...</p>
            </div>
        );
    }

    if (games.length === 0) {
        return (
            <div className="game-history">
                <h3>📜 Recent Games</h3>
                <p style={{ textAlign: 'center', color: 'var(--text-secondary)' }}>No games played yet</p>
            </div>
        );
    }

    return (
        <div className="game-history">
            <h3>📜 Recent Games</h3>
            <div className="history-list">
                {games.slice(0, 10).map((game, index) => (
                    <div key={game.id || index} className="history-item">
                        <div className="players">
                            <span className={game.winner === 1 ? 'winner' : ''}>
                                {game.player1}
                            </span>
                            <span className="vs">vs</span>
                            <span className={game.winner === 2 ? 'winner' : ''}>
                                {game.player2}
                            </span>
                        </div>
                        <div className="details">
                            <span className="duration">⏱️ {formatDuration(game.duration)}</span>
                            <span className="result">
                                {game.winner === 0 ? '🤝 Draw' : `🏆 ${game.winner === 1 ? game.player1 : game.player2}`}
                            </span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
