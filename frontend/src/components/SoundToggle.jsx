import { useState } from 'react';
import { soundManager } from '../utils/sounds';

export default function SoundToggle() {
    const [enabled, setEnabled] = useState(true);

    const toggle = () => {
        const newState = !enabled;
        setEnabled(newState);
        soundManager.setEnabled(newState);
        if (newState) {
            soundManager.playClickSound();
        }
    };

    return (
        <button
            className="sound-toggle"
            onClick={toggle}
            title={enabled ? 'Mute sounds' : 'Enable sounds'}
        >
            {enabled ? '🔊' : '🔇'}
        </button>
    );
}
