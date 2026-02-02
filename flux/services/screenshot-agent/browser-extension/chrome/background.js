// Background service worker para Chrome Extension

// Configuração padrão
const DEFAULT_CONFIG = {
  agentUrl: 'http://localhost:8080/api/v1/screenshots',
  format: 'png',
  quality: 100
};

// Função para carregar configuração do storage
async function loadConfig() {
  return new Promise((resolve) => {
    chrome.storage.sync.get(['agentUrl', 'format', 'quality'], (result) => {
      resolve({
        ...DEFAULT_CONFIG,
        ...result
      });
    });
  });
}

// Listen for extension installation
chrome.runtime.onInstalled.addListener(() => {
  console.log('Screenshot Agent extension installed');
});

// Listen for messages from popup or content scripts
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'captureScreenshot') {
    handleScreenshotCapture(request, sender)
      .then((result) => sendResponse({ success: true, data: result }))
      .catch((error) => sendResponse({ success: false, error: error.message }));
    return true; // Indicates we will send a response asynchronously
  }
  
  if (request.action === 'getConfig') {
    loadConfig()
      .then((config) => sendResponse({ success: true, config }))
      .catch((error) => sendResponse({ success: false, error: error.message }));
    return true;
  }
});

async function handleScreenshotCapture(request, sender) {
  try {
    const config = await loadConfig();
    const tabId = sender.tab?.id || request.tabId;
    
    if (!tabId) {
      throw new Error('Tab ID não encontrado');
    }
    
    // Capture screenshot using chrome.tabs.captureVisibleTab
    const dataUrl = await new Promise((resolve, reject) => {
      chrome.tabs.captureVisibleTab(null, { format: config.format }, (dataUrl) => {
        if (chrome.runtime.lastError) {
          reject(new Error(chrome.runtime.lastError.message));
        } else {
          resolve(dataUrl);
        }
      });
    });
    
    // Convert data URL to blob
    const blob = await dataURLToBlob(dataUrl);
    
    // Get tab info
    const tab = await new Promise((resolve, reject) => {
      chrome.tabs.get(tabId, (tab) => {
        if (chrome.runtime.lastError) {
          reject(new Error(chrome.runtime.lastError.message));
        } else {
          resolve(tab);
        }
      });
    });
    
    // Upload to agent
    const result = await uploadToAgent(blob, tab, config);
    
    return {
      screenshotUrl: dataUrl,
      uploadResult: result,
      tab: {
        url: tab.url,
        title: tab.title
      }
    };
  } catch (error) {
    console.error('Error capturing screenshot:', error);
    throw error;
  }
}

function dataURLToBlob(dataURL) {
  return new Promise((resolve, reject) => {
    try {
      const arr = dataURL.split(',');
      const mime = arr[0].match(/:(.*?);/)[1];
      const bstr = atob(arr[1]);
      let n = bstr.length;
      const u8arr = new Uint8Array(n);
      
      while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
      }
      
      resolve(new Blob([u8arr], { type: mime }));
    } catch (error) {
      reject(error);
    }
  });
}

async function uploadToAgent(blob, tab, config) {
  const formData = new FormData();
  formData.append('screenshot', blob, `screenshot-${Date.now()}.${config.format}`);
  formData.append('url', tab.url || '');
  formData.append('title', tab.title || '');
  formData.append('timestamp', new Date().toISOString());
  
  const response = await fetch(config.agentUrl, {
    method: 'POST',
    body: formData
  });
  
  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Upload failed: ${response.status} ${response.statusText} - ${errorText}`);
  }
  
  return await response.json();
}
