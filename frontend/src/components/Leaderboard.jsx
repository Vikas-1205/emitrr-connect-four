import { useState, useEffect } from 'react';
import { API_URL } from '../utils/config';

export default function Leaderboard({ onClose }) {
    const [players, setPlayers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        fetchLeaderboard();
    }, []);

    const fetchLeaderboard = async () => {
        try {
            const response = await fetch(`${API_URL}/api/leaderboard`);
            if (!response.ok) throw new Error('Failed to fetch leaderboard');
            const data = await response.json();
            setPlayers(data || []);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="leaderboard card">
                <h2>🏆 Leaderboard</h2>
                <div style={{ textAlign: 'center', padding: '2rem' }}>
                    Loading...
                </div>
            </div>
        );
    }

    return (
        <div className="leaderboard card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
                <h2 style={{ margin: 0 }}>🏆 Leaderboard</h2>
                {onClose && (
                    <button className="btn btn-secondary" onClick={onClose} style={{ padding: '0.5rem 1rem' }}>
                        ✕
                    </button>
                )}
            </div>

            {error ? (
                <div style={{ textAlign: 'center', color: 'var(--accent-secondary)', padding: '1rem' }}>
                    {error}
                </div>
            ) : players.length === 0 ? (
                <div style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: '2rem' }}>
                    No games played yet. Be the first! 🎮
                </div>
            ) : (
                <div className="leaderboard-list">
                    {players.map((player, index) => (
                        <div key={player.username} className="leaderboard-item">
                            <div className="rank">{index + 1}</div>
                            <div className="name">{player.username}</div>
                            <div className="stats">
                                <span className="wins">🏆 {player.wins}</span>
                                <span className="losses">💔 {player.losses || 0}</span>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            <button
                className="btn btn-secondary"
                onClick={fetchLeaderboard}
                style={{ width: '100%', marginTop: '1rem' }}
            >
                🔄 Refresh
            </button>
        </div>
    );
}
