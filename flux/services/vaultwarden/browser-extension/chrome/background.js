// Background service worker for Chrome extension

const API_BASE_URL = "https://vaultwarden.lucena.cloud";

// Initialize extension
chrome.runtime.onInstalled.addListener(() => {
  console.log("Vaultwarden extension installed");
});

// Handle messages from content scripts and popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === "getAccessToken") {
    getAccessToken().then(token => {
      sendResponse({ token });
    });
    return true; // Keep channel open for async response
  }
  
  if (request.action === "fetchCiphers") {
    getAccessToken().then(token => {
      return fetch(`${API_BASE_URL}/api/ciphers`, {
        headers: {
          "Authorization": `Bearer ${token}`
        }
      });
    })
    .then(response => response.json())
    .then(data => {
      sendResponse({ ciphers: data.Data || [] });
    })
    .catch(error => {
      sendResponse({ error: error.message });
    });
    return true;
  }
});

// Get access token from storage
async function getAccessToken() {
  const result = await chrome.storage.local.get(["accessToken"]);
  return result.accessToken;
}

// Store access token
async function setAccessToken(token) {
  await chrome.storage.local.set({ accessToken: token });
}

// Login function
async function login(email, password) {
  const formData = new URLSearchParams();
  formData.append("grant_type", "password");
  formData.append("username", email);
  formData.append("password", password);
  
  const response = await fetch(`${API_BASE_URL}/api/identity/connect/token`, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded"
    },
    body: formData
  });
  
  if (!response.ok) {
    throw new Error("Login failed");
  }
  
  const data = await response.json();
  await setAccessToken(data.access_token);
  return data;
}

// Export for use in popup/options
if (typeof module !== "undefined" && module.exports) {
  module.exports = { login, getAccessToken, setAccessToken };
}
