import { useState, useEffect } from 'react';
import { getAvatarForUsername, avatarStyles } from '../utils/avatars';
import { soundManager } from '../utils/sounds';

export default function LobbyScreen({
    onJoinGame,
    isConnected,
    defaultUsername = '',
    difficulty = 'medium',
    onDifficultyChange
}) {
    const [username, setUsername] = useState(defaultUsername);
    const [error, setError] = useState('');
    const [selectedDifficulty, setSelectedDifficulty] = useState(difficulty);

    // Update username if defaultUsername changes (e.g., from Play Again)
    useEffect(() => {
        if (defaultUsername) {
            setUsername(defaultUsername);
        }
    }, [defaultUsername]);

    const avatar = getAvatarForUsername(username);

    const handleSubmit = (e) => {
        e.preventDefault();

        const trimmedUsername = username.trim();
        if (!trimmedUsername) {
            setError('Please enter a username');
            return;
        }

        if (trimmedUsername.length < 2) {
            setError('Username must be at least 2 characters');
            return;
        }

        if (trimmedUsername.length > 20) {
            setError('Username must be less than 20 characters');
            return;
        }

        setError('');
        soundManager.playClickSound();
        onJoinGame(trimmedUsername);
    };

    const handleDifficultySelect = (diff) => {
        setSelectedDifficulty(diff);
        soundManager.playClickSound();
        if (onDifficultyChange) {
            onDifficultyChange(diff);
        }
    };

    const difficulties = [
        { id: 'easy', label: 'Easy', emoji: '🟢', depth: 2 },
        { id: 'medium', label: 'Medium', emoji: '🟡', depth: 4 },
        { id: 'hard', label: 'Hard', emoji: '🔴', depth: 6 }
    ];

    return (
        <div className="lobby">
            <div className="card lobby-card">
                <h2>🎮 Join Game</h2>

                {/* Avatar Preview */}
                {username.trim() && (
                    <div className="avatar-preview">
                        <div style={avatarStyles.container(avatar.color, 60)}>
                            {avatar.emoji}
                        </div>
                        <span className="avatar-name">{username}</span>
                    </div>
                )}

                <form onSubmit={handleSubmit}>
                    <div className="input-group">
                        <label htmlFor="username">Enter your username</label>
                        <input
                            type="text"
                            id="username"
                            className="input"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="Your nickname..."
                            autoFocus
                            maxLength={20}
                        />
                        {error && <span style={{ color: '#ff6b6b', fontSize: '0.875rem' }}>{error}</span>}
                    </div>

                    {/* Bot Difficulty Selector */}
                    <div className="difficulty-selector">
                        <label>Bot Difficulty (if no opponent)</label>
                        <div className="difficulty-options">
                            {difficulties.map(diff => (
                                <button
                                    key={diff.id}
                                    type="button"
                                    className={`difficulty-btn ${selectedDifficulty === diff.id ? 'selected' : ''}`}
                                    onClick={() => handleDifficultySelect(diff.id)}
                                >
                                    <span className="emoji">{diff.emoji}</span>
                                    <span className="label">{diff.label}</span>
                                </button>
                            ))}
                        </div>
                    </div>

                    <button
                        type="submit"
                        className="btn btn-primary"
                        style={{ width: '100%' }}
                        disabled={!isConnected}
                    >
                        {isConnected ? '🚀 Find Match' : '⏳ Connecting...'}
                    </button>
                </form>

                <div style={{ marginTop: '1.5rem', textAlign: 'center', color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                    <p>You'll be matched with another player or play against our AI bot!</p>
                </div>
            </div>
        </div>
    );
}
