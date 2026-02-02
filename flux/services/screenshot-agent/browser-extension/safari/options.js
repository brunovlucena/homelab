// Options page script (Safari compatible)

// Compatibilidade com browser/chrome API
const storage = typeof browser !== 'undefined' ? browser.storage : chrome.storage;

document.addEventListener('DOMContentLoaded', async () => {
  const form = document.getElementById('configForm');
  const statusDiv = document.getElementById('status');
  
  // Load current config
  await loadConfig();
  
  // Save config on form submit
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const config = {
      agentUrl: document.getElementById('agentUrl').value,
      format: document.getElementById('format').value
    };
    
    try {
      await new Promise((resolve) => {
        storage.sync.set(config, resolve);
      });
      showStatus('✅ Configurações salvas com sucesso!', 'success');
    } catch (error) {
      showStatus(`❌ Erro ao salvar: ${error.message}`, 'error');
    }
  });
});

async function loadConfig() {
  try {
    const result = await new Promise((resolve) => {
      storage.sync.get(['agentUrl', 'format'], resolve);
    });
    
    if (result.agentUrl) {
      document.getElementById('agentUrl').value = result.agentUrl;
    }
    
    if (result.format) {
      document.getElementById('format').value = result.format;
    }
  } catch (error) {
    console.error('Error loading config:', error);
  }
}

function showStatus(message, type) {
  const statusDiv = document.getElementById('status');
  statusDiv.textContent = message;
  statusDiv.className = `status ${type}`;
  
  // Clear status after 3 seconds
  setTimeout(() => {
    statusDiv.className = '';
    statusDiv.textContent = '';
  }, 3000);
}
