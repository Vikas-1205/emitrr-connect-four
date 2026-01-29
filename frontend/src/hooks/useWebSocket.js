import { useState, useEffect, useCallback, useRef } from 'react';
import { soundManager } from '../utils/sounds';

const WS_URL = import.meta.env.PROD
    ? `wss://${window.location.host}/ws`
    : 'ws://localhost:8080/ws';

export function useWebSocket() {
    const [isConnected, setIsConnected] = useState(false);
    const [gameState, setGameState] = useState(null);
    const [status, setStatus] = useState('disconnected'); // disconnected, connecting, waiting, playing
    const [playerNum, setPlayerNum] = useState(null);
    const [message, setMessage] = useState('');
    const [lastMove, setLastMove] = useState(null);
    const [savedUsername, setSavedUsername] = useState('');
    const [difficulty, setDifficulty] = useState('medium');
    const [lastOpponent, setLastOpponent] = useState(null);
    const wsRef = useRef(null);
    const reconnectTimeoutRef = useRef(null);
    const usernameRef = useRef('');

    const connect = useCallback(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        setStatus('connecting');
        const ws = new WebSocket(WS_URL);

        ws.onopen = () => {
            setIsConnected(true);
            setStatus('connected');
            setMessage('');

            // If we had a username, try to reconnect to game
            if (usernameRef.current && gameState?.id) {
                ws.send(JSON.stringify({
                    type: 'reconnect',
                    payload: {
                        username: usernameRef.current,
                        gameId: gameState.id
                    }
                }));
            }
        };

        ws.onclose = () => {
            setIsConnected(false);
            setStatus('disconnected');

            // Attempt reconnect after 2 seconds
            reconnectTimeoutRef.current = setTimeout(() => {
                connect();
            }, 2000);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                handleMessage(msg);
            } catch (e) {
                console.error('Failed to parse message:', e);
            }
        };

        wsRef.current = ws;
    }, [gameState?.id]);

    const handleMessage = useCallback((msg) => {
        const { type, payload } = msg;
        const data = typeof payload === 'string' ? JSON.parse(payload) : payload;

        switch (type) {
            case 'waiting':
                setStatus('waiting');
                setMessage(data.message || 'Waiting for opponent...');
                break;

            case 'game_start':
                setStatus('playing');
                setPlayerNum(data.playerNum);
                setGameState(data.state);
                setMessage('');
                soundManager.playMatchSound();
                // Save opponent for rematch
                if (data.state) {
                    setLastOpponent(data.playerNum === 1 ? data.state.player2Username : data.state.player1Username);
                }
                break;

            case 'move_made':
                setLastMove({ col: data.column, row: data.row, player: data.player });
                soundManager.playDropSound();
                setGameState(prev => {
                    if (!prev) return prev;
                    const newBoard = prev.board.map(row => [...row]);
                    newBoard[data.row][data.column] = data.player;
                    return {
                        ...prev,
                        board: newBoard,
                        currentPlayer: data.currentPlayer
                    };
                });
                // Play turn sound if it's now my turn
                if (data.currentPlayer === playerNum) {
                    soundManager.playTurnSound();
                }
                break;

            case 'game_over':
                setGameState(prev => ({
                    ...prev,
                    status: 'completed',
                    winner: data.winner,
                    winningCells: data.winningCells || []
                }));
                setStatus('game_over');
                // Play appropriate sound
                if (data.winner === playerNum) {
                    soundManager.playWinSound();
                } else if (data.winner === 0) {
                    soundManager.playDrawSound();
                } else {
                    soundManager.playLoseSound();
                }
                break;

            case 'opponent_left':
                setMessage(data.message);
                break;

            case 'reconnected':
                setStatus('playing');
                setGameState(data);
                setMessage('');
                // Determine player number from reconnection
                if (data.player1Username === usernameRef.current) {
                    setPlayerNum(1);
                } else {
                    setPlayerNum(2);
                }
                break;

            case 'error':
                setMessage(data.error);
                break;
        }
    }, [playerNum]);

    const joinQueue = useCallback((username, botDifficulty = 'medium') => {
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
            connect();
            // Retry after connection
            setTimeout(() => joinQueue(username, botDifficulty), 500);
            return;
        }

        usernameRef.current = username;
        setSavedUsername(username);
        setDifficulty(botDifficulty);
        wsRef.current.send(JSON.stringify({
            type: 'join_queue',
            payload: { username, difficulty: botDifficulty }
        }));
        setStatus('waiting');
        soundManager.playClickSound();
    }, [connect]);

    const makeMove = useCallback((column) => {
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
            return false;
        }

        wsRef.current.send(JSON.stringify({
            type: 'make_move',
            payload: { column }
        }));
        return true;
    }, []);

    const requestRematch = useCallback(() => {
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
            return false;
        }

        wsRef.current.send(JSON.stringify({
            type: 'rematch_request',
            payload: {}
        }));
        setMessage('Rematch requested...');
        return true;
    }, []);

    const resetGame = useCallback(() => {
        setGameState(null);
        setPlayerNum(null);
        setStatus('connected');
        setMessage('');
        setLastMove(null);
    }, []);

    useEffect(() => {
        connect();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [connect]);

    return {
        isConnected,
        status,
        gameState,
        playerNum,
        message,
        lastMove,
        joinQueue,
        makeMove,
        resetGame,
        requestRematch,
        username: savedUsername,
        difficulty,
        setDifficulty,
        lastOpponent
    };
}
