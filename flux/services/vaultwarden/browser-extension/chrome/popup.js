// Popup script

const API_BASE_URL = "https://vaultwarden.lucena.cloud";

// Check if already logged in
chrome.storage.local.get(["accessToken"], (result) => {
  if (result.accessToken) {
    showVaultView();
  } else {
    showLoginView();
  }
});

// Login form handler
document.getElementById("loginForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;
  const errorDiv = document.getElementById("error");
  
  try {
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
    await chrome.storage.local.set({ accessToken: data.access_token });
    
    showVaultView();
  } catch (error) {
    errorDiv.textContent = "Login failed. Please check your credentials.";
  }
});

// Show vault view
async function showVaultView() {
  document.getElementById("loginView").style.display = "none";
  document.getElementById("vaultView").style.display = "block";
  
  await loadPasswords();
}

// Show login view
function showLoginView() {
  document.getElementById("loginView").style.display = "block";
  document.getElementById("vaultView").style.display = "none";
}

// Load passwords
async function loadPasswords() {
  const result = await chrome.storage.local.get(["accessToken"]);
  const token = result.accessToken;
  
  if (!token) {
    showLoginView();
    return;
  }
  
  try {
    const response = await fetch(`${API_BASE_URL}/api/ciphers`, {
      headers: {
        "Authorization": `Bearer ${token}`
      }
    });
    
    if (!response.ok) {
      throw new Error("Failed to fetch passwords");
    }
    
    const data = await response.json();
    displayPasswords(data.Data || []);
  } catch (error) {
    console.error("Error loading passwords:", error);
    showLoginView();
  }
}

// Display passwords
function displayPasswords(ciphers) {
  const listDiv = document.getElementById("passwordList");
  listDiv.innerHTML = "";
  
  if (ciphers.length === 0) {
    listDiv.innerHTML = "<div style='padding: 16px; text-align: center; color: #666;'>No passwords stored</div>";
    return;
  }
  
  ciphers.forEach(cipher => {
    const item = document.createElement("div");
    item.className = "password-item";
    item.innerHTML = `
      <div class="password-name">${cipher.name || "Untitled"}</div>
      <div class="password-username">${cipher.login?.username || ""}</div>
    `;
    item.addEventListener("click", () => {
      // Copy password to clipboard
      if (cipher.login?.password) {
        navigator.clipboard.writeText(cipher.login.password);
      }
    });
    listDiv.appendChild(item);
  });
}
