// Configuration
const API_URL = '/api';
let ws = null;
let currentUser = null;
let currentChat = null;
let chats = new Map();
let pendingRequests = new Map();
let requestIdCounter = 0;
let accessToken = localStorage.getItem('access_token');
let refreshToken = localStorage.getItem('refresh_token');
let lastMessageSender = null;

// Auth pages
function showLogin() {
    document.getElementById('login-form').classList.remove('hidden');
    document.getElementById('register-form').classList.add('hidden');
    clearErrors();
}

function showRegister() {
    document.getElementById('login-form').classList.add('hidden');
    document.getElementById('register-form').classList.remove('hidden');
    clearErrors();
}

function clearErrors() {
    document.querySelectorAll('.error-message').forEach(el => el.classList.remove('visible'));
    document.querySelectorAll('input').forEach(el => el.classList.remove('error'));
}

function showError(elementId, message) {
    const errorEl = document.getElementById(elementId + '-error');
    const inputEl = document.getElementById(elementId);
    if (errorEl && inputEl) {
        errorEl.textContent = message;
        errorEl.classList.add('visible');
        inputEl.classList.add('error');
    }
}

// Toast notifications
function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
        <span>${message}</span>
        <button class="toast-close" onclick="this.parentElement.remove()">&times;</button>
    `;
    container.appendChild(toast);

    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
    }, 5000);
}

// Login
async function handleLogin(event) {
    event.preventDefault();
    clearErrors();

    const username = document.getElementById('login-username').value.trim();
    const password = document.getElementById('login-password').value;

    if (!username) {
        showError('login-username', 'Введите имя пользователя');
        return;
    }
    if (!password) {
        showError('login-password', 'Введите пароль');
        return;
    }

    const btn = document.getElementById('login-btn');
    const btnText = btn.querySelector('.btn-text');
    btnText.textContent = 'Вход...';
    btn.disabled = true;

    try {
        const response = await fetch(`${API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            const text = await response.text();
            if (response.status === 401) {
                throw new Error(text || 'Неверный логин или пароль');
            }
            throw new Error(text || 'Ошибка входа');
        }

        const data = await response.json();
        accessToken = data.access_token;
        refreshToken = data.refresh_token;
        localStorage.setItem('access_token', accessToken);
        localStorage.setItem('refresh_token', refreshToken);

        currentUser = username;
        showToast('Успешный вход!', 'success');

        showMainPage();
    } catch (error) {
        showToast(error.message, 'error');
        showError('login-password', error.message);
    } finally {
        btnText.textContent = 'Войти';
        btn.disabled = false;
    }
}

// Register
async function handleRegister(event) {
    event.preventDefault();
    clearErrors();

    const username = document.getElementById('register-username').value.trim();
    const password = document.getElementById('register-password').value;

    if (!username) {
        showError('register-username', 'Введите имя пользователя');
        return;
    }
    if (password.length < 8) {
        showError('register-password', 'Пароль должен содержать минимум 8 символов');
        return;
    }

    const btn = document.getElementById('register-btn');
    const btnText = btn.querySelector('.btn-text');
    btnText.textContent = 'Регистрация...';
    btn.disabled = true;

    try {
        const response = await fetch(`${API_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            const text = await response.text();
            if (response.status === 409) {
                throw new Error(text || 'Пользователь уже существует');
            }
            throw new Error(text || 'Ошибка регистрации');
        }

        showToast('Регистрация успешна! Вход...', 'success');

        // После регистрации автоматически входим
        await login(username, password);
    } catch (error) {
        showToast(error.message, 'error');
        showError('register-password', error.message);
    } finally {
        btnText.textContent = 'Зарегистрироваться';
        btn.disabled = false;
    }
}

async function login(username, password) {
    const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    if (!response.ok) {
        throw new Error('Ошибка входа');
    }

    const data = await response.json();
    accessToken = data.access_token;
    refreshToken = data.refresh_token;
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);

    currentUser = username;
}

// Logout
async function handleLogout() {
    if (refreshToken) {
        try {
            await fetch(`${API_URL}/auth/logout`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ refresh_token: refreshToken })
            });
        } catch (e) {
            console.error('Logout error:', e);
        }
    }

    disconnect();
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    currentUser = null;
    currentChat = null;
    chats.clear();
    // Возвращаем область чата в исходное состояние
    document.getElementById('empty-state').classList.remove('hidden');
    document.getElementById('chat-container').classList.add('hidden');
    document.getElementById('chat-header').textContent = 'Выберите чат';
    document.getElementById('main-page').classList.add('hidden');
    document.getElementById('auth-page').classList.remove('hidden');
    document.getElementById('login-form').classList.remove('hidden');
    showToast('Вы вышли из системы', 'info');
}

// Main page
function showMainPage() {
    document.getElementById('auth-page').classList.add('hidden');
    document.getElementById('main-page').classList.remove('hidden');
    connect();
}

// WebSocket connection
function connect() {
    if (!currentUser) {
        showLogin();
        return;
    }

    const wsProto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProto}//${location.host}/ws?token=${accessToken}`;
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        updateStatus(true);
        loadChats();
        startHeartbeat();
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);

        if (data.type === 'message.new') {
            const message = data.payload;

            if (currentChat && message.chat_uuid === currentChat) {
                addMessageToChat(message);
            } else {
                loadChats();
            }

            updateChatLastMessage(message.chat_uuid, message);

            if (!isSelfMessage(message)) {
                showToast(`Новое сообщение от ${message.sender_username}`, 'info');
            }
        } else {
            handleMessage(data);
        }
    };

    ws.onerror = () => {
        updateStatus(false);
    };

    ws.onclose = () => {
        updateStatus(false);
        if (heartbeatInterval) {
            clearInterval(heartbeatInterval);
        }
    };
}

function disconnect() {
    if (ws) {
        ws.close();
        ws = null;
    }
}

let heartbeatInterval = null;

function startHeartbeat() {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
    }
    heartbeatInterval = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            send('user.update_last_seen', {});
        }
    }, 30000);
}

// Refresh token
async function refreshAccessToken() {
    try {
        const response = await fetch(`${API_URL}/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        if (!response.ok) {
            throw new Error('Срок сессии истёк');
        }

        const data = await response.json();
        accessToken = data.access_token;
        refreshToken = data.refresh_token;
        localStorage.setItem('access_token', accessToken);
        localStorage.setItem('refresh_token', refreshToken);
    } catch (error) {
        showToast('Срок сессии истёк. Пожалуйста, войдите снова.', 'error');
        disconnect();
        handleLogout();
    }
}

// Send message
function send(type, payload, callback) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        if (callback) {
            callback(null);
        }
        return;
    }

    const requestId = getRequestId();
    if (callback) {
        pendingRequests.set(requestId, callback);
    }

    ws.send(JSON.stringify({
        type: type,
        request_id: requestId,
        payload: payload
    }));
}

function getRequestId() {
    return `req_${++requestIdCounter}_${Date.now()}`;
}

// Create chat
function openCreateChatModal() {
    document.getElementById('create-chat-modal').classList.add('active');
    document.getElementById('chat-participant-username').value = '';
    document.getElementById('chat-participant-error').textContent = '';
    document.getElementById('chat-participant-error').classList.remove('visible');
}

function closeCreateChatModal() {
    document.getElementById('create-chat-modal').classList.remove('active');
}

async function createChat() {
    const username = document.getElementById('chat-participant-username').value.trim();

    if (!username) {
        showError('chat-participant', 'Введите имя пользователя');
        return;
    }

    // Модалку закрываем только при успехе: при ошибке она остаётся открытой,
    // а тост с причиной покажет обработчик error-ответа
    send('chat.create.personal', {
        receiver_name: username
    }, (response) => {
        if (!response || !response.chat_uuid) {
            showToast('Не удалось создать чат', 'error');
            return;
        }
        closeCreateChatModal();
        showToast('Чат создан!', 'success');
        loadChats();
        selectChat(response.chat_uuid);
    });
}

// Chat list
function loadChats() {
    send('chat.list', {}, (chatsList) => {
        renderChats(chatsList);
    });
}

function renderChats(chatsList) {
    const container = document.getElementById('chat-list');
    container.innerHTML = '';

    if (!chatsList || !Array.isArray(chatsList)) {
        return;
    }

    // "New Chat" button
    const newChatBtn = document.createElement('div');
    newChatBtn.className = 'chat-item';
    newChatBtn.style.background = '#667eea';
    newChatBtn.textContent = '+ Новый чат';
    newChatBtn.onclick = openCreateChatModal;
    container.appendChild(newChatBtn);

    chatsList.forEach(chat => {
        const chatDiv = document.createElement('div');
        chatDiv.className = 'chat-item';
        if (currentChat === chat.uuid) {
            chatDiv.classList.add('active');
        }

        let chatName = chat.name || 'Чат';
        if (chat.type === 'personal') {
            const otherUser = chat.participants?.find(p => p.username !== currentUser);
            chatName = otherUser ? otherUser.username : 'Неизвестный';
        }

        chatDiv.innerHTML = `
            <div class="chat-item-info">
                <div class="chat-item-name">${escapeHtml(chatName)}</div>
                <div class="chat-item-last-message">${chat.last_message ? escapeHtml(chat.last_message.body) : 'Нет сообщений'}</div>
            </div>
            <div class="chat-item-meta">
                <span class="chat-item-time">${formatTime(chat.updated_at)}</span>
            </div>
        `;

        chatDiv.onclick = () => selectChat(chat.uuid);
        container.appendChild(chatDiv);
        chats.set(chat.uuid, chat);
    });
}

// Select chat
function selectChat(chatUUID) {
    currentChat = chatUUID;
    document.getElementById('empty-state').classList.add('hidden');
    document.getElementById('chat-container').classList.remove('hidden');
    document.getElementById('messages').innerHTML = '';

    send('chat.get', { chat_uuid: chatUUID }, (chatData) => {
        if (!chatData || !chatData.chat) return;
        const chat = chatData.chat;
        let chatName = chat.name || 'Чат';
        if (chat.type === 'personal') {
            const otherUser = chatData.participants?.find(p => p.username !== currentUser);
            chatName = otherUser ? otherUser.username : 'Неизвестный';
        }
        document.getElementById('chat-header').textContent = chatName;

        loadMessages(chatUUID);
    });

    document.querySelectorAll('.chat-item').forEach(chat => {
        chat.classList.remove('active');
    });
}

function loadMessages(chatUUID) {
    send('message.history', { chat_uuid: chatUUID, limit: 50, offset: 0 }, (messages) => {
        (messages || []).reverse().forEach(msg => addMessageToChat(msg));
    });
}

// Messages
function sendMessage() {
    const input = document.getElementById('message-input');
    const body = input.value.trim();

    if (!body || !currentChat) return;

    send('message.send', {
        chat_uuid: currentChat,
        body: body
    }, (response) => {
        if (!response) return;
        input.value = '';

        const tempMessage = {
            message_uuid: response.message_uuid || 'temp_' + Date.now(),
            chat_uuid: currentChat,
            sender_uuid: currentUser,
            sender_username: currentUser,
            body: body,
            created_at: new Date().toISOString()
        };

        addMessageToChat(tempMessage);
        // Себе broadcast не приходит (отправитель исключён) — обновляем список сами
        updateChatLastMessage(currentChat, tempMessage);
    });
}

function addMessageToChat(message) {
    const messagesContainer = document.getElementById('messages');
    const messageDiv = document.createElement('div');

    const isSelf = isSelfMessage(message);
    const senderName = message.sender_username || message.sender_uuid || 'Неизвестный';

    let timeStr = '';
    if (message.created_at && message.created_at !== '0001-01-01T00:00:00Z') {
        const time = new Date(message.created_at);
        if (!isNaN(time.getTime())) {
            timeStr = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        }
    }
    if (!timeStr) {
        timeStr = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    const showSender = !isSelf && (!lastMessageSender || lastMessageSender !== senderName);

    messageDiv.className = `message ${isSelf ? 'self' : 'other'}`;
    messageDiv.dataset.sender = senderName;

    let html = '';

    if (showSender && !isSelf) {
        html += `<div class="sender">${escapeHtml(senderName)}</div>`;
    }

    html += `<div class="message-content">`;
    html += `<div class="text">${escapeHtml(message.body)}</div>`;
    html += `<div class="time">${timeStr}</div>`;
    html += `</div>`;

    messageDiv.innerHTML = html;
    messagesContainer.appendChild(messageDiv);

    messagesContainer.scrollTop = messagesContainer.scrollHeight;
    lastMessageSender = senderName;
}

function handleKeyPress(event) {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        sendMessage();
    }
}

function updateChatLastMessage(chatUUID, message) {
    const chat = chats.get(chatUUID);
    if (chat) {
        chat.last_message = message;
        renderChats(Array.from(chats.values()));
    }
}

function updateStatus(connected) {
    const statusDot = document.getElementById('status-dot');
    const statusText = document.getElementById('status-text');
    const statusBadge = document.getElementById('status-badge');

    if (connected) {
        statusDot.classList.add('connected');
        statusText.textContent = 'Подключено';
        statusBadge.classList.add('connected');
    } else {
        statusDot.classList.remove('connected');
        statusText.textContent = 'Отключено';
        statusBadge.classList.remove('connected');
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatTime(isoString) {
    if (!isoString || isoString === '0001-01-01T00:00:00Z') {
        return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    return new Date(isoString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function isSelfMessage(message) {
    if (message.sender_username) {
        return message.sender_username === currentUser;
    }
    if (message.sender_uuid) {
        return message.sender_uuid === currentUser;
    }
    return false;
}

// Handle server responses
function handleMessage(data) {
    if (data.type === 'response' && data.request_id) {
        const callback = pendingRequests.get(data.request_id);
        if (callback) {
            callback(data.payload);
            pendingRequests.delete(data.request_id);
        }
    } else if (data.type === 'error') {
        if (data.request_id) {
            pendingRequests.delete(data.request_id);
        }
        showToast(`Ошибка: ${data.payload.error}`, 'error');
    }
}

// Initialize
if (accessToken && refreshToken) {
    showMainPage();
} else {
    showLogin();
}

// Request notification permission
if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission();
}
