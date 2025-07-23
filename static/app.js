const tg = window.Telegram.WebApp;
let chat_id;

window.onload = async () => {
  tg.ready();
  const initData = tg?.initDataUnsafe;
  const user = initData?.user ?? null;
  const chat = initData?.chat ?? null;
  const initRaw = tg?.initData;

  /*if (!chat?.id) {
    // ❌ Not launched in a group
    // Report to backend for debugging
    try {
      await fetch("/report-unauthorized", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          user: user,
          chat: chat,
          initData: initRaw ?? "",
        }),
      });
    } catch (err) {
      console.warn("Debug report failed:", err);
    }

    document.body.innerHTML = '<h3>❌ This app must be opened from a Telegram <b>group chat</b>.</h3>';
    return;
  }

  chat_id = chat.id;*/

  chat_id = 2867211623;
  tg.expand();

  try {
    await Promise.all([
      loadBioMode(),
      loadEchoSettings(),
      loadLinkFilter()
    ]);
    document.getElementById("loading").style.display = "none";
    document.getElementById("settings-container").style.display = "block";

    // Initialize the first section as expanded
    toggleSection('biomode');
  } catch (e) {
    showErrorPage(e?.message ?? e);
  }
};

// ----------------------- Section Toggle Functionality -----------------------
function toggleSection(sectionId) {
  const content = document.getElementById(`${sectionId}-content`);
  const icon = document.getElementById(`${sectionId}-icon`);

  // Close all other sections
  const allSections = ['biomode', 'echo', 'linkfilter'];
  allSections.forEach(id => {
    if (id !== sectionId) {
      const otherContent = document.getElementById(`${id}-content`);
      const otherIcon = document.getElementById(`${id}-icon`);
      otherContent.classList.remove('expanded');
      otherIcon.classList.remove('expanded');
    }
  });

  // Toggle current section
  content.classList.toggle('expanded');
  icon.classList.toggle('expanded');
}

// ----------------------- Error Handling -----------------------
function showErrorPage(error) {
  document.getElementById("loading").style.display = "none";
  document.body.innerHTML = `
    <div class="error-container">
      <div class="error-card">
        <div class="error-icon">⚠️</div>
        <h2 class="error-title">Settings Unavailable</h2>
        <p class="error-message">Unable to connect to the settings API</p>
        <div class="error-details">
          <p><strong>Error:</strong> ${error}</p>
          <p>Please check that your backend API is running and accessible.</p>
        </div>
        <div class="error-actions">
          <button onclick="location.reload()" class="retry-btn">🔄 Retry</button>
        </div>
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
    document.getElementById('biomode-switch').checked = !!enabled;
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
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(enabled),
  });
}

// ----------------------- Echo Settings -----------------------
async function loadEchoSettings() {
  try {
    const res = await fetch(`/api/echo?chat_id=${chat_id}`);
    if (!res.ok) throw new Error("API not available");
    const data = await res.json();
    document.getElementById('longmode-select').value = data?.long_mode ?? 'automatic';
    document.getElementById('longlimit-input').value = data?.long_limit ?? 800;
  } catch (e) {
    // Fallback to demo mode
    document.getElementById('longmode-select').value = 'automatic';
    document.getElementById('longlimit-input').value = 800;
    throw new Error("Could not load Echo Settings - API endpoint not found");
  }
}

function saveEchoSettings() {
  const long_mode = document.getElementById('longmode-select').value;
  let long_limit = parseInt(document.getElementById('longlimit-input').value, 10);
  if (isNaN(long_limit) || long_limit < 200 || long_limit > 4000) {
    showToast("⚠️ Long limit must be between 200 and 4000", "warning");
    return Promise.reject("Invalid long limit");
  }

  return fetch(`/api/echo?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {'Content-Type': 'application/json' },
    body: JSON.stringify({ long_mode, long_limit }),
  });
}

// ----------------------- Link Filter -----------------------
async function loadLinkFilter() {
  try {
    const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
    if (!res.ok) throw new Error("API not available");
    const data = await res.json();
    document.getElementById('linkfilter-switch').checked = !!data?.enabled;
    const domains = data?.allowed_domains ?? [];

    const tbody = document.getElementById('allowed-links-body');
    tbody.innerHTML = '';
    domains.forEach(domain => addDomainRow(domain));
  } catch (e) {
    // Fallback to demo mode
    document.getElementById('linkfilter-switch').checked = false;
    const tbody = document.getElementById('allowed-links-body');
    tbody.innerHTML = '';
    throw new Error("Could not load LinkFilter - API endpoint not found");
  }
}

function addDomainRow(domain) {
  const tbody = document.getElementById('allowed-links-body');
  if ([...tbody.children].some(tr => tr.children[0].textContent === domain)) return;

  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td>${domain}</td>
    <td><button class="remove-btn" aria-label="Remove domain">Remove</button></td>
  `;
  tr.querySelector('button').onclick = () => {
    tbody.removeChild(tr);
    showToast(`Domain "${domain}" removed`, "info");
  };
  tbody.appendChild(tr);
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input');
  const domain = input?.value?.trim()?.toLowerCase();
  if (domain) {
    addDomainRow(domain);
    input.value = '';
    showToast(`Domain "${domain}" added`, "success");
  } else {
    showToast("Please enter a valid domain", "warning");
  }
};

async function saveLinkFilter() {
  const enabled = document.getElementById('linkfilter-switch').checked;

  // Toggle on/off first
  const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled }),
  });
  if (!res.ok) throw new Error("Failed to update LinkFilter state");

  // Get full list from backend to find deltas
  const response = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
  const data = await response.json();
  const oldDomains = new Set(data?.allowed_domains ?? []);

  const tbody = document.getElementById('allowed-links-body');
  const newDomains = Array.from(tbody.children).map(row => row.children[0].textContent);

  for (const d of newDomains) {
    if (!oldDomains.has(d)) {
      const resAdd = await fetch(`/api/linkfilter/allow?chat_id=${chat_id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: d }),
      });
      if (!resAdd.ok) throw new Error("Failed to add domain: " + d);
    }
  }

  // Remove domains that were removed in the UI
  for (const d of oldDomains) {
    if (!newDomains.includes(d)) {
      const resRm = await fetch(`/api/linkfilter/remove?chat_id=${chat_id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: d }),
      });
      if (!resRm.ok) throw new Error("Failed to remove domain: " + d);
    }
  }
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
    background: ${type === 'success' ? '#48bb78' : type === 'warning' ? '#ed8936' : type === 'error' ? '#e53e3e' : '#4299e1'};
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

// ----------------------- Save All -----------------------
document.getElementById('save-all').onclick = async () => {
  const saveButton = document.getElementById('save-all');
  const originalText = saveButton.textContent;

  try {
    saveButton.textContent = '💾 Saving...';
    saveButton.disabled = true;

    await saveBioMode();
    await saveEchoSettings();
    await saveLinkFilter();

    showToast("✅ All settings saved successfully!", "success");

    // Provide haptic feedback if available
    if (tg.HapticFeedback) {
      tg.HapticFeedback.notificationOccurred('success');
    }
  } catch (error) {
    showToast("❌ Failed to save: " + (error?.message ?? error), "error");

    // Provide haptic feedback if available
    if (tg.HapticFeedback) {
      tg.HapticFeedback.notificationOccurred('error');
    }
  } finally {
    saveButton.textContent = originalText;
    saveButton.disabled = false;
  }
};