let ws = null;
let myUsername = '';
let myRoom = '';

// Paleta de colores para avatares
const COLORS = ['#60A5FA','#34D399','#FBBF24','#F472B6','#A78BFA','#FB923C'];
const userColors = {};

function getColor(username) {
  if (!userColors[username]) {
    userColors[username] = COLORS[Object.keys(userColors).length % COLORS.length];
  }
  return userColors[username];
}

// ── CONECTAR ────────────────────────────────────────────────────────
function connect() {
  const username = document.getElementById('input-username').value.trim();
  const room = document.getElementById('input-room').value.trim() || 'General';
  const errEl = document.getElementById('login-error');

  if (!username) {
    errEl.textContent = 'Escribe tu nombre de usuario';
    errEl.style.display = 'block';
    return;
  }

  myUsername = username;
  myRoom = room;

  // Crear conexión WebSocket
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const url = `${proto}://${location.host}/ws?username=${encodeURIComponent(username)}&room=${encodeURIComponent(room)}`;

  ws = new WebSocket(url);

  ws.onopen = () => {
    // Mostrar pantalla de chat
    document.getElementById('login-screen').style.display = 'none';
    document.getElementById('chat-screen').style.display = 'flex';
    document.getElementById('header-room').textContent = '# ' + room;
    setStatus(true);
    document.getElementById('msg-input').focus();
  };

  ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    handleMessage(msg);
  };

  ws.onclose = () => {
    setStatus(false);
    addSystemMessage('Conexión cerrada — recarga la página para reconectar');
  };

  ws.onerror = () => {
    errEl.textContent = 'No se pudo conectar. ¿Está corriendo el servidor?';
    errEl.style.display = 'block';
  };
}

// ── DESCONECTAR ──────────────────────────────────────────────────────
function disconnect() {
  if (ws) ws.close();
  document.getElementById('chat-screen').style.display = 'none';
  document.getElementById('login-screen').style.display = 'flex';
  document.getElementById('messages').innerHTML = '';
  document.getElementById('login-error').style.display = 'none';
}

// ── ENVIAR MENSAJE ───────────────────────────────────────────────────
function sendMessage() {
  const input = document.getElementById('msg-input');
  const content = input.value.trim();
  if (!content || !ws || ws.readyState !== WebSocket.OPEN) return;

  ws.send(JSON.stringify({ content }));
  input.value = '';
  input.focus();
}

// ── MANEJAR MENSAJES DEL SERVIDOR ────────────────────────────────────
function handleMessage(msg) {
  switch(msg.type) {
    case 'chat':
      addChatMessage(msg);
      break;
    case 'join':
    case 'leave':
      addSystemMessage(msg.content);
      if (msg.users) updateUserList(msg.users);
      break;
    case 'users':
      updateUserList(msg.users);
      break;
  }
}

// ── RENDERIZAR BURBUJA DE CHAT ───────────────────────────────────────
function addChatMessage(msg) {
  const isOwn = msg.username === myUsername;
  const color = getColor(msg.username);
  const time = new Date(msg.timestamp).toLocaleTimeString('es', {
    hour: '2-digit', minute: '2-digit'
  });

  const div = document.createElement('div');
  div.className = `msg ${isOwn ? 'own' : 'other'}`;
  div.innerHTML = `
    ${!isOwn ? `
      <div class="msg-meta">
        <div class="msg-avatar" style="background:${color}20;color:${color}">
          ${msg.username[0].toUpperCase()}
        </div>
        <span style="color:${color};font-weight:600">${msg.username}</span>
      </div>
    ` : ''}
    <div class="msg-bubble">${escapeHTML(msg.content)}</div>
    <div class="msg-time">${time}</div>
  `;

  document.getElementById('messages').appendChild(div);
  scrollToBottom();
}

// ── MENSAJE DEL SISTEMA ──────────────────────────────────────────────
function addSystemMessage(text) {
  const div = document.createElement('div');
  div.className = 'msg-system';
  div.innerHTML = `<span>${escapeHTML(text)}</span>`;
  document.getElementById('messages').appendChild(div);
  scrollToBottom();
}

// ── LISTA DE USUARIOS ────────────────────────────────────────────────
function updateUserList(users) {
  const list = document.getElementById('users-list');
  list.innerHTML = users.map(u => `
    <div class="user-item">
      <div class="user-dot"></div>
      <span style="color:${getColor(u)};font-weight:${u===myUsername?'700':'400'}">
        ${escapeHTML(u)} ${u === myUsername ? '(tú)' : ''}
      </span>
    </div>
  `).join('');
}

// ── HELPERS ──────────────────────────────────────────────────────────
function setStatus(connected) {
  const dot = document.getElementById('ws-dot');
  const text = document.getElementById('ws-status-text');
  dot.className = 'ws-dot' + (connected ? '' : ' off');
  text.textContent = connected ? 'Conectado' : 'Desconectado';
}

function scrollToBottom() {
  const el = document.getElementById('messages');
  el.scrollTop = el.scrollHeight;
}

function escapeHTML(str) {
  return str
    .replace(/&/g,'&amp;')
    .replace(/</g,'&lt;')
    .replace(/>/g,'&gt;')
    .replace(/"/g,'&quot;');
}

// Enter en el login
document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('input-room').addEventListener('keydown', e => {
    if (e.key === 'Enter') connect();
  });
  document.getElementById('input-username').addEventListener('keydown', e => {
    if (e.key === 'Enter') document.getElementById('input-room').focus();
  });
});
