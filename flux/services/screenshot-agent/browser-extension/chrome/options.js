// Options page script

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
      await chrome.storage.sync.set(config);
      showStatus('✅ Configurações salvas com sucesso!', 'success');
    } catch (error) {
      showStatus(`❌ Erro ao salvar: ${error.message}`, 'error');
    }
  });
});

async function loadConfig() {
  try {
    const result = await chrome.storage.sync.get(['agentUrl', 'format']);
    
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
