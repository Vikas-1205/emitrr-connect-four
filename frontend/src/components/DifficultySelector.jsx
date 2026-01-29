import { useState } from 'react';

export default function DifficultySelector({ onSelect }) {
    const [selected, setSelected] = useState('medium');

    const difficulties = [
        { id: 'easy', label: 'Easy', emoji: '🟢', description: 'For beginners' },
        { id: 'medium', label: 'Medium', emoji: '🟡', description: 'Balanced challenge' },
        { id: 'hard', label: 'Hard', emoji: '🔴', description: 'Expert level' }
    ];

    const handleSelect = (id) => {
        setSelected(id);
        if (onSelect) onSelect(id);
    };

    return (
        <div className="difficulty-selector">
            <label>Bot Difficulty</label>
            <div className="difficulty-options">
                {difficulties.map(diff => (
                    <button
                        key={diff.id}
                        className={`difficulty-btn ${selected === diff.id ? 'selected' : ''}`}
                        onClick={() => handleSelect(diff.id)}
                        title={diff.description}
                    >
                        <span className="emoji">{diff.emoji}</span>
                        <span className="label">{diff.label}</span>
                    </button>
                ))}
            </div>
        </div>
    );
}
