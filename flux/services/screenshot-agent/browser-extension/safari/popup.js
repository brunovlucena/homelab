// Popup script for Safari extension (compatible with Chrome)

document.addEventListener('DOMContentLoaded', async () => {
  const captureBtn = document.getElementById('captureBtn');
  const statusDiv = document.getElementById('status');
  const optionsLink = document.getElementById('optionsLink');
  
  // Compatibilidade com browser/chrome API
  const runtime = typeof browser !== 'undefined' ? browser.runtime : chrome.runtime;
  const tabs = typeof browser !== 'undefined' ? browser.tabs : chrome.tabs;
  
  // Open options page
  optionsLink.addEventListener('click', (e) => {
    e.preventDefault();
    runtime.openOptionsPage();
  });
  
  // Capture screenshot button
  captureBtn.addEventListener('click', async () => {
    try {
      // Disable button and show loading
      captureBtn.disabled = true;
      showStatus('Capturando screenshot...', 'loading');
      
      // Get current tab
      const tabList = await new Promise((resolve) => {
        tabs.query({ active: true, currentWindow: true }, resolve);
      });
      const tab = tabList[0];
      
      if (!tab) {
        throw new Error('Não foi possível encontrar a aba ativa');
      }
      
      // Send message to background script
      const response = await new Promise((resolve) => {
        runtime.sendMessage({
          action: 'captureScreenshot',
          tabId: tab.id
        }, resolve);
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
    const response = await new Promise((resolve) => {
      runtime.sendMessage({ action: 'getConfig' }, resolve);
    });
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
