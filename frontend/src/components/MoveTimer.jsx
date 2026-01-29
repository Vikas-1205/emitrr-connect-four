import { useState, useEffect } from 'react';
import { soundManager } from '../utils/sounds';

export default function MoveTimer({ isMyTurn, isGameActive, onTimeout }) {
    const [timeLeft, setTimeLeft] = useState(30);
    const [isWarning, setIsWarning] = useState(false);

    useEffect(() => {
        if (!isGameActive || !isMyTurn) {
            setTimeLeft(30);
            setIsWarning(false);
            return;
        }

        const timer = setInterval(() => {
            setTimeLeft(prev => {
                if (prev <= 1) {
                    clearInterval(timer);
                    if (onTimeout) onTimeout();
                    return 0;
                }

                // Warning sound when < 10 seconds
                if (prev <= 10 && prev > 0) {
                    setIsWarning(true);
                    soundManager.playWarningSound();
                }

                return prev - 1;
            });
        }, 1000);

        return () => clearInterval(timer);
    }, [isMyTurn, isGameActive, onTimeout]);

    // Reset timer when turn changes
    useEffect(() => {
        setTimeLeft(30);
        setIsWarning(false);
    }, [isMyTurn]);

    if (!isGameActive) return null;

    const percentage = (timeLeft / 30) * 100;

    return (
        <div className={`move-timer ${isWarning ? 'warning' : ''}`}>
            <div className="timer-bar">
                <div
                    className="timer-fill"
                    style={{
                        width: `${percentage}%`,
                        background: isWarning
                            ? 'linear-gradient(90deg, #ff6b6b, #ff4757)'
                            : 'linear-gradient(90deg, var(--accent-primary), #4d96ff)'
                    }}
                />
            </div>
            <span className={`timer-text ${isWarning ? 'pulse' : ''}`}>
                {isMyTurn ? `⏱️ ${timeLeft}s` : `Opponent: ${timeLeft}s`}
            </span>
        </div>
    );
}
