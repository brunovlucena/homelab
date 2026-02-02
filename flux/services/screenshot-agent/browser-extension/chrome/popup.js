// Popup script for Chrome extension

document.addEventListener('DOMContentLoaded', async () => {
  const captureBtn = document.getElementById('captureBtn');
  const statusDiv = document.getElementById('status');
  const optionsLink = document.getElementById('optionsLink');
  
  // Open options page
  optionsLink.addEventListener('click', (e) => {
    e.preventDefault();
    chrome.runtime.openOptionsPage();
  });
  
  // Capture screenshot button
  captureBtn.addEventListener('click', async () => {
    try {
      // Disable button and show loading
      captureBtn.disabled = true;
      showStatus('Capturando screenshot...', 'loading');
      
      // Get current tab
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      
      if (!tab) {
        throw new Error('Não foi possível encontrar a aba ativa');
      }
      
      // Send message to background script
      const response = await chrome.runtime.sendMessage({
        action: 'captureScreenshot',
        tabId: tab.id
      });
      
      if (response.success) {
        showStatus('✅ Screenshot enviado com sucesso!', 'success');
        document.getElementById('info').textContent = `Enviado de: ${response.data.tab.title || response.data.tab.url}`;
      } else {
        throw new Error(response.error || 'Erro ao capturar screenshot');
      }
    } catch (error) {
      console.error('Error:', error);
      showStatus(`❌ Erro: ${error.message}`, 'error');
    } finally {
      captureBtn.disabled = false;
      
      // Clear status after 3 seconds
      setTimeout(() => {
        statusDiv.innerHTML = '';
        statusDiv.className = '';
      }, 3000);
    }
  });
  
  // Load config to verify connection
  try {
    const response = await chrome.runtime.sendMessage({ action: 'getConfig' });
    if (response.success && response.config) {
      const agentUrl = response.config.agentUrl || 'Não configurado';
      if (agentUrl === 'http://localhost:8080/api/v1/screenshots') {
        document.getElementById('info').textContent = '⚠️ Configure a URL do agente nas opções';
      }
    }
  } catch (error) {
    console.error('Error loading config:', error);
  }
});

function showStatus(message, type) {
  const statusDiv = document.getElementById('status');
  statusDiv.textContent = message;
  statusDiv.className = `status ${type}`;
}
