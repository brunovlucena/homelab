// Configuração padrão do agente
const DEFAULT_CONFIG = {
  agentUrl: 'http://localhost:8080/api/v1/screenshots',
  // Para homelab, use algo como:
  // agentUrl: 'https://api.lucena.cloud/api/v1/screenshots',
  // agentUrl: 'http://homelab-api.local:8080/api/v1/screenshots',
  format: 'png',
  quality: 100,
  fullPage: false,
  timeout: 30000
};

// Função para carregar configuração do storage
async function loadConfig() {
  return new Promise((resolve) => {
    if (typeof chrome !== 'undefined' && chrome.storage) {
      chrome.storage.sync.get(['agentUrl', 'format', 'quality'], (result) => {
        resolve({
          ...DEFAULT_CONFIG,
          ...result
        });
      });
    } else if (typeof browser !== 'undefined' && browser.storage) {
      browser.storage.sync.get(['agentUrl', 'format', 'quality']).then((result) => {
        resolve({
          ...DEFAULT_CONFIG,
          ...result
        });
      });
    } else {
      resolve(DEFAULT_CONFIG);
    }
  });
}

// Função para salvar configuração
async function saveConfig(config) {
  return new Promise((resolve) => {
    const toSave = {
      agentUrl: config.agentUrl,
      format: config.format,
      quality: config.quality
    };
    
    if (typeof chrome !== 'undefined' && chrome.storage) {
      chrome.storage.sync.set(toSave, () => resolve());
    } else if (typeof browser !== 'undefined' && browser.storage) {
      browser.storage.sync.set(toSave).then(() => resolve());
    } else {
      resolve();
    }
  });
}
