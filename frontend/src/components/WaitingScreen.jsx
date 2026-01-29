import { useState, useEffect } from 'react';

export default function WaitingScreen({ message }) {
    const [countdown, setCountdown] = useState(10);

    useEffect(() => {
        const timer = setInterval(() => {
            setCountdown((prev) => {
                if (prev <= 1) {
                    clearInterval(timer);
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);

        return () => clearInterval(timer);
    }, []);

    return (
        <div className="lobby">
            <div className="card lobby-card">
                <div className="waiting">
                    <div className="waiting-dots">
                        <span></span>
                        <span></span>
                        <span></span>
                    </div>
                    <h2>Finding Opponent</h2>

                    {/* Countdown Timer */}
                    <div className="countdown">
                        <div className="countdown-number">{countdown}</div>
                        <p>seconds remaining</p>
                    </div>

                    <p style={{ color: 'var(--text-secondary)', marginTop: '1rem' }}>
                        {countdown > 0
                            ? 'Looking for a worthy opponent...'
                            : 'Starting game with AI bot...'
                        }
                    </p>
                </div>
            </div>
        </div>
    );
}
