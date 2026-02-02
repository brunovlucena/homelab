// Speech Coach Pi Client - JavaScript

const API_URL = '/api';
let currentExerciseType = null;
let currentSessionId = null;
let isRecording = false;

// Elements
const chatMessages = document.getElementById('chatMessages');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');
const micButton = document.getElementById('micButton');
const themeSelect = document.getElementById('themeSelect');
const loadingIndicator = document.getElementById('loadingIndicator');
const progressDisplay = document.getElementById('progressDisplay');

// Theme management
function changeTheme(theme) {
    const link = document.querySelector('link[rel="stylesheet"]');
    const href = link.href.split('/');
    href[href.length - 1] = `${theme}.css`;
    link.href = href.join('/');
    localStorage.setItem('speechCoachTheme', theme);
}

// Load saved theme
const savedTheme = localStorage.getItem('speechCoachTheme') || 'default';
themeSelect.value = savedTheme;
changeTheme(savedTheme);

themeSelect.addEventListener('change', (e) => {
    changeTheme(e.target.value);
});

// Send message
async function sendMessage(message, exerciseType = null) {
    if (!message.trim()) return;
    
    // Add user message to chat
    addMessage(message, true);
    
    // Clear input
    messageInput.value = '';
    
    // Show loading
    showLoading(true);
    
    try {
        const response = await fetch(`${API_URL}/message`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message,
                exercise_type: exerciseType || currentExerciseType,
                session_id: currentSessionId,
            }),
        });
        
        const data = await response.json();
        
        if (data.error) {
            addMessage(`Erro: ${data.error}`, false);
        } else {
            // Update session ID if provided
            if (data.session_id) {
                currentSessionId = data.session_id;
            }
            
            // Add agent response
            addMessage(data.response || 'Sem resposta', false);
            
            // Show exercise if provided
            if (data.exercise) {
                showExercise(data.exercise);
            }
            
            // Update progress
            if (data.progress) {
                updateProgress(data.progress);
            }
            
            // Show suggestions
            if (data.suggestions && data.suggestions.length > 0) {
                showSuggestions(data.suggestions);
            }
        }
    } catch (error) {
        addMessage(`Erro de conexão: ${error.message}`, false);
    } finally {
        showLoading(false);
    }
}

// Add message to chat
function addMessage(content, isFromUser) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${isFromUser ? 'user' : 'agent'}`;
    messageDiv.textContent = content;
    
    // Remove welcome message if exists
    const welcomeMsg = chatMessages.querySelector('.welcome-message');
    if (welcomeMsg) {
        welcomeMsg.remove();
    }
    
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Show loading indicator
function showLoading(show) {
    loadingIndicator.style.display = show ? 'flex' : 'none';
}

// Show exercise
function showExercise(exercise) {
    if (!exercise) return;
    
    const exerciseDiv = document.createElement('div');
    exerciseDiv.className = 'exercise-info';
    exerciseDiv.innerHTML = `
        <h4>📝 Exercício: ${exercise.title}</h4>
        <p>${exercise.description}</p>
        <p><strong>Instruções:</strong> ${exercise.instructions}</p>
    `;
    chatMessages.appendChild(exerciseDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Show suggestions
function showSuggestions(suggestions) {
    const suggestionsDiv = document.createElement('div');
    suggestionsDiv.className = 'suggestions';
    suggestionsDiv.innerHTML = `
        <h4>💡 Sugestões:</h4>
        <ul>
            ${suggestions.map(s => `<li>${s}</li>`).join('')}
        </ul>
    `;
    chatMessages.appendChild(suggestionsDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Update progress display
function updateProgress(progress) {
    if (!progress) return;
    
    progressDisplay.innerHTML = `
        <div class="progress-item">
            <strong>Sessões:</strong> ${progress.total_sessions || 0}
        </div>
        <div class="progress-item">
            <strong>Exercícios:</strong> ${progress.completed_exercises || 0}
        </div>
        <div class="progress-item">
            <strong>Pontos:</strong> ${progress.total_points || 0}
        </div>
        <div class="progress-item">
            <strong>Sequência:</strong> ${progress.current_streak || 0} dias
        </div>
    `;
}

// Event listeners
sendButton.addEventListener('click', () => {
    const message = messageInput.value.trim();
    if (message) {
        sendMessage(message);
    }
});

messageInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        const message = messageInput.value.trim();
        if (message) {
            sendMessage(message);
        }
    }
});

// Quick action buttons
document.querySelectorAll('.btn-action').forEach(button => {
    button.addEventListener('click', () => {
        const exerciseType = button.dataset.exercise;
        currentExerciseType = exerciseType;
        const exerciseName = button.textContent.trim();
        sendMessage(`Quero fazer o exercício: ${exerciseName}`, exerciseType);
    });
});

// Speech recognition (if available)
if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    const recognition = new SpeechRecognition();
    recognition.lang = 'pt-BR';
    recognition.continuous = false;
    recognition.interimResults = false;
    
    micButton.addEventListener('click', () => {
        if (isRecording) {
            recognition.stop();
            micButton.classList.remove('recording');
            isRecording = false;
        } else {
            recognition.start();
            micButton.classList.add('recording');
            isRecording = true;
        }
    });
    
    recognition.onresult = (event) => {
        const transcript = event.results[0][0].transcript;
        messageInput.value = transcript;
        sendMessage(transcript);
        micButton.classList.remove('recording');
        isRecording = false;
    };
    
    recognition.onerror = () => {
        micButton.classList.remove('recording');
        isRecording = false;
        addMessage('Erro ao reconhecer voz. Tente novamente.', false);
    };
    
    recognition.onend = () => {
        micButton.classList.remove('recording');
        isRecording = false;
    };
} else {
    micButton.style.display = 'none';
}

// Check health on load
fetch(`${API_URL}/health`)
    .then(res => res.json())
    .then(data => {
        if (!data.agent_connected) {
            addMessage('⚠️ Não conectado ao agente. Verifique a configuração.', false);
        }
    });
