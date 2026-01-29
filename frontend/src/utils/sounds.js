// Sound effects for the game
// Using Web Audio API for cross-browser compatibility

class SoundManager {
    constructor() {
        this.audioContext = null;
        this.enabled = true;
        this.volume = 0.5;
    }

    init() {
        if (!this.audioContext) {
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        }
    }

    setEnabled(enabled) {
        this.enabled = enabled;
    }

    setVolume(volume) {
        this.volume = Math.max(0, Math.min(1, volume));
    }

    // Generate a tone programmatically
    playTone(frequency, duration, type = 'sine') {
        if (!this.enabled) return;

        try {
            this.init();
            const oscillator = this.audioContext.createOscillator();
            const gainNode = this.audioContext.createGain();

            oscillator.connect(gainNode);
            gainNode.connect(this.audioContext.destination);

            oscillator.frequency.value = frequency;
            oscillator.type = type;

            gainNode.gain.setValueAtTime(this.volume, this.audioContext.currentTime);
            gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioContext.currentTime + duration);

            oscillator.start(this.audioContext.currentTime);
            oscillator.stop(this.audioContext.currentTime + duration);
        } catch (e) {
            console.warn('Sound playback failed:', e);
        }
    }

    // Disc drop sound - descending tone
    playDropSound() {
        if (!this.enabled) return;
        this.init();

        const frequencies = [600, 500, 400, 350];
        frequencies.forEach((freq, i) => {
            setTimeout(() => this.playTone(freq, 0.08, 'square'), i * 40);
        });
    }

    // Win sound - triumphant fanfare
    playWinSound() {
        if (!this.enabled) return;
        this.init();

        const notes = [
            { freq: 523, delay: 0 },     // C5
            { freq: 659, delay: 150 },   // E5
            { freq: 784, delay: 300 },   // G5
            { freq: 1047, delay: 450 },  // C6
        ];

        notes.forEach(note => {
            setTimeout(() => this.playTone(note.freq, 0.3, 'square'), note.delay);
        });
    }

    // Lose sound - sad descending
    playLoseSound() {
        if (!this.enabled) return;
        this.init();

        const notes = [400, 350, 300, 250];
        notes.forEach((freq, i) => {
            setTimeout(() => this.playTone(freq, 0.25, 'sawtooth'), i * 200);
        });
    }

    // Draw sound - neutral
    playDrawSound() {
        if (!this.enabled) return;
        this.init();

        this.playTone(440, 0.3, 'triangle');
        setTimeout(() => this.playTone(440, 0.3, 'triangle'), 350);
    }

    // Click/select sound
    playClickSound() {
        if (!this.enabled) return;
        this.playTone(800, 0.05, 'square');
    }

    // Match found sound
    playMatchSound() {
        if (!this.enabled) return;
        this.init();

        const notes = [440, 550, 660];
        notes.forEach((freq, i) => {
            setTimeout(() => this.playTone(freq, 0.1, 'sine'), i * 100);
        });
    }

    // Turn notification sound
    playTurnSound() {
        if (!this.enabled) return;
        this.playTone(600, 0.1, 'sine');
    }

    // Timer warning sound (when < 10 seconds)
    playWarningSound() {
        if (!this.enabled) return;
        this.playTone(880, 0.1, 'square');
    }
}

// Export singleton instance
export const soundManager = new SoundManager();
