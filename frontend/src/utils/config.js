// API Configuration - handles different environments
const getApiUrl = () => {
    // Environment variable takes priority (for Render/production with separate backend)
    if (import.meta.env.VITE_API_URL) {
        return import.meta.env.VITE_API_URL;
    }
    // Production same-origin
    if (import.meta.env.PROD) {
        return '';
    }
    // Development
    return 'http://localhost:8080';
};

const getWsUrl = () => {
    // Environment variable takes priority
    if (import.meta.env.VITE_WS_URL) {
        return import.meta.env.VITE_WS_URL;
    }
    // Production same-origin
    if (import.meta.env.PROD) {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}/ws`;
    }
    // Development
    return 'ws://localhost:8080/ws';
};

export const API_URL = getApiUrl();
export const WS_URL = getWsUrl();
