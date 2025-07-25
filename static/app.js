const tg = window.Telegram.WebApp;
let chat_id;
tg.ready();

window.onload = async () => {
  // tg.ready();
  applyTheme();
  tg.onEvent("themeChanged", applyTheme);

  if (!tg || !tg.initDataUnsafe || !tg.initDataUnsafe.user) {
    showErrorPage("This page must be opened inside Telegram.", {
      title: "WebApp Only",
      message: "This tool can only be used from within the Telegram WebApp.",
      showRetry: false
    });
    return;
  }

  const initData = tg.initDataUnsafe;
  const user = initData.user;

  // ✅ Extract access_key from start_param
  const startParam = initData.start_param;

  if (!startParam || !startParam.startsWith("access_key")) {
    showErrorPage("Missing or invalid 'access_key' in Telegram WebApp", {
      title: "Invalid Request",
      message: "This page requires a valid access_key to function.",
      showRetry: false
    });
    return;
  }

  const access_key = startParam.replace("access_key", "");

  try {
    chat_id = decodeDigits(access_key);
  } catch (e) {
    showErrorPage(e?.message ?? e, {
      title: "Settings Unavailable",
      message: "Looks like your access_key is wrong?"
    });
    return;
  }

  tg.expand();
  tg.enableClosingConfirmation();

  try {
    await Promise.all([
      loadBioMode(),
      loadEchoSettings(),
      loadLinkFilter()
    ]);

    document.getElementById("loading").style.display = "none";
    document.getElementById("main-content").style.display = "block";

  } catch (e) {
    showErrorPage(e?.message ?? e, {
      title: "Settings Unavailable",
      message: "Unable to connect to the Backend API"
    });
  }
};


// ----------------------- access_key to chat_id -----------------------

const digitMap = "adefjtghkz"; // must match Go

function decodeDigits(str) {
  if (str.length !== 10) {
    throw new Error("Invalid input length: expected 10 characters");
  }

  let result = "";
  for (let ch of str) {
    const idx = digitMap.indexOf(ch);
    if (idx === -1) throw new Error("Invalid encoded character: " + ch);
    result += idx.toString();
  }
  return parseInt(result, 10);
}
// ----------------------- Theme Support -----------------------

function applyTheme() {
  let theme = "light"; // default fallback
  if (tg && typeof tg.colorScheme === "string") {
    theme = tg.colorScheme === "dark" ? "dark": "light";
  }
  document.body.setAttribute("data-theme", theme);
}



// ----------------------- Validation Functions -----------------------
function extractHostname(input) {
  input = input.trim();
  if (!input.startsWith("http://") && !input.startsWith("https://")) {
    input = "http://" + input;
  }

  try {
    const urlObj = new URL(input);
    return urlObj.hostname.toLowerCase();
  } catch {
    return "";
  }
}

function isValidDomain(hostname) {
  if (!hostname) return false;

  const ipv4Pattern = /^(?:\d{1,3}\.){3}\d{1,3}$/;
  const ipv6Pattern = /^\[?[a-fA-F0-9:]+\]?$/;
  const domainPattern = /^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  return (
    ipv4Pattern.test(hostname) ||
    ipv6Pattern.test(hostname) ||
    domainPattern.test(hostname)
  );
}

function validateLongLimit(value) {
  const num = parseInt(value, 10);
  return !isNaN(num) && num >= 200 && num <= 4000;
}

// ----------------------- Section Toggle Functionality -----------------------
function toggleSection(sectionId) {
  const content = document.getElementById(`${sectionId}-content`);
  const icon = document.getElementById(`${sectionId}-icon`);

  /* // Close all other sections
  const allSections = ['biomode', 'echo', 'linkfilter'];
  allSections.forEach(id => {
    if (id !== sectionId) {
      const otherContent = document.getElementById(`${id}-content`);
      const otherIcon = document.getElementById(`${id}-icon`);
      otherContent.classList.remove('expanded');
      otherIcon.classList.remove('expanded');
    }
  });*/

  // Toggle current section
  content.classList.toggle('expanded');
  icon.classList.toggle('expanded');
}

// ----------------------- Error Handling -----------------------
function showErrorPage(error, options = {}) {
  const {
    title = "Something Went Wrong",
    message = "An unexpected error occurred.",
    showRetry = true
  } = options;

  document.getElementById("loading")?.remove();
  document.body.innerHTML = `
  <div class="error-container">
  <div class="error-card">
  <div class="error-icon">⚠️</div>
  <h2 class="error-title">${title}</h2>
  <p class="error-message">${message}</p>
  <div class="error-details">
  <p><strong>Error:</strong> ${error ?? "Unknown error"}</p>
  </div>
  ${
  showRetry
  ? `<div class="error-actions">
  <button onclick="location.reload()" class="retry-btn">🔄 Retry</button>
  </div>`: ""
  }
  </div>
  </div>
  `;
}

// ----------------------- BioMode -----------------------
async function loadBioMode() {
  try {
    const res = await fetch(`/api/biomode?chat_id=${chat_id}`);
    if (!res.ok) throw new Error("API not available");
    const enabled = await res.json();
    const switchEl = document.getElementById('biomode-switch');
    switchEl.checked = !!enabled;

    // Add live update event listener
    switchEl.addEventListener('sl-change', () => {
      saveBioMode().catch(err => {
        showToast(`❌ Failed to update BioMode:  ${err?.message || err || "Unknown error"}`, "error");
      });
    });
  } catch (e) {
    // Fallback to demo mode
    document.getElementById('biomode-switch').checked = false;
    throw new Error("Could not load BioMode - API endpoint not found");
  }
}

function saveBioMode() {
  const enabled = document.getElementById('biomode-switch').checked;
  return fetch(`/api/biomode?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(enabled),
  });
}

// ----------------------- Echo Settings -----------------------
async function loadEchoSettings() {
  try {
    const res = await fetch(`/api/echo?chat_id=${chat_id}`);
    if (!res.ok) throw new Error("API not available");
    const data = await res.json();

    const selectEl = document.getElementById('longmode-select');
    const inputEl = document.getElementById('longlimit-input');
    const saveBtn = document.getElementById('save-echo-btn');

    selectEl.value = data?.long_mode ?? 'automatic';
    inputEl.value = data?.long_limit ?? 800;

    // ✅ Set click event on save button
    saveBtn.addEventListener('click', () => {
      if (!validateLongLimit(inputEl.value)) {
        showToast("⚠️ Long limit must be between 200 and 4000", "warning");
        inputEl.value = data?.long_limit ?? 800;
        return;
      }

      saveEchoSettings().catch(err => {
        showToast(`❌ Failed to save echo settings:  ${err?.message || err || "Unknown error"}`, "error");
      });
    });

  } catch (e) {
    document.getElementById('longmode-select').value = 'automatic';
    document.getElementById('longlimit-input').value = 800;
    throw new Error("Could not load Echo Settings - API endpoint not found");
  }
}

function saveEchoSettings() {
  const long_mode = document.getElementById('longmode-select').value;
  let long_limit = parseInt(document.getElementById('longlimit-input').value,
    10);
  if (isNaN(long_limit) || long_limit < 200 || long_limit > 4000) {
    showToast("⚠️ Long limit must be between 200 and 4000", "warning");
    return Promise.reject("Invalid long limit");
  }

  return fetch(`/api/echo?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      long_mode, long_limit
    }),
  });
}

// ----------------------- Link Filter -----------------------
async function loadLinkFilter() {
  try {
    const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
    if (!res.ok) throw new Error("API not available");
    const data = await res.json();

    const switchEl = document.getElementById('linkfilter-switch');
    switchEl.checked = !!data?.enabled;
    const domains = data?.allowed_domains ?? [];

    const listEl = document.getElementById('allowed-links-list');
    listEl.innerHTML = '';
    domains.forEach(domain => addDomainItem(domain));

    // Add live update event listener for switch
    switchEl.addEventListener('sl-change', () => {
      saveLinkFilterState().catch(err => {
        showToast(`❌ Failed to update link filter:  ${err?.message || err || "Unknown error"}`, "error");
      });
    });
  } catch (e) {
    // Fallback to demo mode
    document.getElementById('linkfilter-switch').checked = false;
    const listEl = document.getElementById('allowed-links-list');
    listEl.innerHTML = '';
    throw new Error("Could not load LinkFilter - API endpoint not found");
  }
}

function addDomainItem(domain) {
  const listEl = document.getElementById('allowed-links-list');
  if ([...listEl.children].some(item => item.querySelector('.domain-name').textContent === domain)) return;

  const item = document.createElement('div');
  item.className = 'domain-item';
  item.innerHTML = `
    <span class="domain-name">${domain}</span>
    <button class="remove-btn" aria-label="Remove domain">🗑️</button>
  `;

  const removeBtn = item.querySelector('button');
  removeBtn.onclick = (e) => {
    e.preventDefault();
    e.stopPropagation();
    item.remove();
    saveDomainRemove(domain).catch(err => {
      showToast(`❌ Failed to remove domain: ${err?.message || err || "Unknown error"}`, "error");
      // Re-add the item if API call fails
      listEl.insertBefore(item, listEl.firstChild);
    });
  };
  listEl.insertBefore(item, listEl.firstChild);
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input');
  const rawValue = input?.value ?? '';
  const hostname = extractHostname(rawValue); // extract and normalize

  if (hostname && isValidDomain(hostname)) {
    addDomainItem(hostname);
    input.value = '';
    saveDomainAdd(hostname)
    .then(() => {
      showToast("✅ Domain added successfully!", "success");
    })
    .catch(err => {
      showToast(`❌ Failed to add domain:  ${err?.message || err || "Unknown error"}`, "error");
    });
  } else {
    showToast("⚠️ Please enter a valid domain (e.g., example.com)", "warning");
  }
};

// Add enter key support for domain input
document.getElementById('allow-domain-input').addEventListener('keypress', (e) => {
  if (e.key === 'Enter') {
    document.getElementById('allow-btn').click();
  }
});

// Simplified functions for live updating
async function saveLinkFilterState() {
  const enabled = document.getElementById('linkfilter-switch').checked;
  const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        enabled
      }),
    });
  if (!res.ok) throw new Error("Failed to update LinkFilter state");
}

async function saveDomainAdd(domain) {
  const res = await fetch(`/api/linkfilter/allow?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      domain
    }),
  });
  if (!res.ok) throw new Error("Failed to add domain");
}

async function saveDomainRemove(domain) {
  const res = await fetch(`/api/linkfilter/remove?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      domain
    }),
  });
  if (!res.ok) throw new Error("Failed to remove domain");
}

// ----------------------- Toast Notifications -----------------------
function showToast(message, type = "info") {
  // Create toast element
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.textContent = message;

  // Style the toast
  toast.style.cssText = `
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: ${type === 'success' ? '#48bb78': type === 'warning' ? '#ed8936': type === 'error' ? '#e53e3e': '#4299e1'};
  color: white;
  padding: 12px 20px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  font-weight: 500;
  opacity: 0;
  transition: opacity 0.3s ease;
  `;

  document.body.appendChild(toast);

  // Animate in
  setTimeout(() => toast.style.opacity = '1', 100);

  // Remove after 3 seconds
  setTimeout(() => {
    toast.style.opacity = '0';
    setTimeout(() => document.body.removeChild(toast), 300);
  }, 3000);
}