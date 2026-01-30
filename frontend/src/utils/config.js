// API Configuration - handles different environments
// Production backend URL (Render deployment)
const PRODUCTION_API = 'https://connect-four-backend-xalu.onrender.com';
const PRODUCTION_WS = 'wss://connect-four-backend-xalu.onrender.com/ws';

const getApiUrl = () => {
    // Environment variable takes priority
    if (import.meta.env.VITE_API_URL) {
        return import.meta.env.VITE_API_URL;
    }
    // Production - use hardcoded backend URL
    if (import.meta.env.PROD) {
        return PRODUCTION_API;
    }
    // Development
    return 'http://localhost:8080';
};

const getWsUrl = () => {
    // Environment variable takes priority
    if (import.meta.env.VITE_WS_URL) {
        return import.meta.env.VITE_WS_URL;
    }
    // Production - use hardcoded backend URL
    if (import.meta.env.PROD) {
        return PRODUCTION_WS;
    }
    // Development
    return 'ws://localhost:8080/ws';
};

export const API_URL = getApiUrl();
export const WS_URL = getWsUrl();

