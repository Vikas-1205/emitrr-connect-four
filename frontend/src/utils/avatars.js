// Player avatar generation
const AVATAR_STYLES = [
    // Animals
    '🦊', '🐼', '🦁', '🐯', '🐨', '🐵', '🦄', '🐶', '🐱', '🐰',
    // Objects
    '🎮', '🎯', '🎲', '🏆', '⭐', '💎', '🔥', '⚡', '🌟', '🎪',
    // Faces
    '😎', '🤖', '👾', '👻', '🎃', '💀', '🤠', '🥷', '🧙', '🦸'
];

const AVATAR_COLORS = [
    '#ff6b6b', '#ffd93d', '#6bcb77', '#4d96ff', '#ff6bff',
    '#6bffff', '#ff9f43', '#a55eea', '#26de81', '#fd79a8'
];

// Generate consistent avatar from username
export function getAvatarForUsername(username) {
    if (!username) return { emoji: '👤', color: '#6c63ff' };

    // Create a simple hash from username
    let hash = 0;
    for (let i = 0; i < username.length; i++) {
        hash = ((hash << 5) - hash) + username.charCodeAt(i);
        hash = hash & hash; // Convert to 32bit integer
    }
    hash = Math.abs(hash);

    const emoji = AVATAR_STYLES[hash % AVATAR_STYLES.length];
    const color = AVATAR_COLORS[hash % AVATAR_COLORS.length];

    return { emoji, color };
}

// Avatar component styles
export const avatarStyles = {
    container: (color, size = 40) => ({
        width: `${size}px`,
        height: `${size}px`,
        borderRadius: '50%',
        background: `linear-gradient(135deg, ${color}33 0%, ${color}66 100%)`,
        border: `2px solid ${color}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: `${size * 0.5}px`,
        flexShrink: 0
    })
};
