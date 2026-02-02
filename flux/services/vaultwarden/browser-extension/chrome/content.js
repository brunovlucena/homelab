// Content script for auto-fill functionality

(function() {
  'use strict';
  
  // Detect password fields on the page
  function detectPasswordFields() {
    const passwordFields = document.querySelectorAll('input[type="password"]');
    const usernameFields = document.querySelectorAll('input[type="email"], input[type="text"][name*="user"], input[type="text"][name*="email"]');
    
    if (passwordFields.length > 0) {
      // Send message to background script to get matching passwords
      chrome.runtime.sendMessage({
        action: "fetchCiphers"
      }, (response) => {
        if (response && response.ciphers) {
          // Find matching password for current domain
          const currentDomain = window.location.hostname;
          const matchingCipher = response.ciphers.find(cipher => {
            if (cipher.login && cipher.login.uris) {
              return cipher.login.uris.some(uri => {
                if (!uri.uri) return false;
                try {
                  const url = new URL(uri.uri);
                  return url.hostname === currentDomain || currentDomain.endsWith(url.hostname);
                } catch {
                  return uri.uri.includes(currentDomain);
                }
              });
            }
            return false;
          });
          
          if (matchingCipher) {
            showAutoFillButton(matchingCipher, passwordFields[0], usernameFields[0]);
          }
        }
      });
    }
  }
  
  // Show auto-fill button
  function showAutoFillButton(cipher, passwordField, usernameField) {
    // Create button element
    const button = document.createElement('div');
    button.id = 'vaultwarden-autofill';
    button.innerHTML = '🔐';
    button.style.cssText = `
      position: absolute;
      width: 30px;
      height: 30px;
      background: #175DDC;
      color: white;
      border-radius: 4px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 16px;
      z-index: 10000;
      box-shadow: 0 2px 4px rgba(0,0,0,0.2);
    `;
    
    // Position button next to password field
    const rect = passwordField.getBoundingClientRect();
    button.style.top = (rect.top + window.scrollY) + 'px';
    button.style.left = (rect.right + window.scrollX + 5) + 'px';
    
    // Add click handler
    button.addEventListener('click', () => {
      if (usernameField && cipher.login && cipher.login.username) {
        usernameField.value = cipher.login.username;
      }
      
      // Decrypt and fill password (in production, decrypt client-side)
      if (cipher.login && cipher.login.password) {
        passwordField.value = cipher.login.password;
      }
      
      button.remove();
    });
    
    document.body.appendChild(button);
  }
  
  // Run detection when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', detectPasswordFields);
  } else {
    detectPasswordFields();
  }
  
  // Also detect dynamically added fields
  const observer = new MutationObserver(() => {
    detectPasswordFields();
  });
  
  observer.observe(document.body, {
    childList: true,
    subtree: true
  });
})();
