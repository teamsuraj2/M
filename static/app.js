const tg = window.Telegram.WebApp;
let chat_id;

window.onload = async () => {
  const initData = tg.initDataUnsafe;

  if (!initData || !initData.chat || !initData.chat.id) {
    document.body.innerHTML = '<h3>❌ This app can only be used inside a Telegram group.</h3>';
    return;
  }
  chat_id = initData.chat.id;
  tg.expand();

  try {
    await Promise.all([loadBioMode(), loadEchoSettings(), loadLinkFilter()]);
    document.getElementById("loading").style.display = "none";
    document.getElementById("settings-container").style.display = "block";
  } catch (e) {
    document.body.innerHTML = `<h3>❗ Failed to load settings: ${e.message}</h3>`;
  }
};

// Load BioMode state (boolean)
async function loadBioMode() {
  const res = await fetch(`/api/biomode?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load BioMode");
  const enabled = await res.json();
  document.getElementById('biomode-switch').checked = enabled;
}

function saveBioMode() {
  const enabled = document.getElementById('biomode-switch').checked;
  return fetch(`/api/biomode?chat_id=${chat_id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(enabled),
  });
}

// Load Echo Settings (long_mode, long_limit)
async function loadEchoSettings() {
  const res = await fetch(`/api/echo?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load Echo Settings");
  const data = await res.json();

  document.getElementById('longmode-select').value = data.long_mode || 'automatic';
  document.getElementById('longlimit-input').value = data.long_limit || 800;
}

function saveEchoSettings() {
  const long_mode = document.getElementById('longmode-select').value;
  let long_limit = parseInt(document.getElementById('longlimit-input').value, 10);
  if (isNaN(long_limit) || long_limit < 200 || long_limit > 4000) {
    alert("Long limit must be between 200 and 4000");
    return Promise.reject("Invalid long limit");
  }
  return fetch(`/api/echo?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {'Content-Type': 'application/json' },
    body: JSON.stringify({ long_mode, long_limit }),
  });
}

// Load LinkFilter (enabled + allowed domains)
async function loadLinkFilter() {
  const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load LinkFilter");
  const data = await res.json();

  document.getElementById('linkfilter-switch').checked = data.enabled;
  const tbody = document.getElementById('allowed-links-body');
  tbody.innerHTML = '';
  data.allowed_domains.forEach(domain => {
    addDomainRow(domain);
  });
}

function addDomainRow(domain) {
  const tbody = document.getElementById('allowed-links-body');
  if ([...tbody.children].some(tr => tr.children[0].textContent === domain)) {
    return; // skip duplicates
  }
  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td>${domain}</td>
    <td><button class="mdui-btn mdui-btn-icon" type="button" aria-label="Remove" title="Remove domain">❌</button></td>
  `;
  tr.querySelector('button').onclick = () => {
    tbody.removeChild(tr);
  };
  tbody.appendChild(tr);
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input');
  const domain = input.value.trim().toLowerCase();
  if (domain) {
    addDomainRow(domain);
    input.value = '';
  }
};

async function saveLinkFilter() {
  const enabled = document.getElementById('linkfilter-switch').checked;
  const tbody = document.getElementById('allowed-links-body');
  const domains = Array.from(tbody.children).map(tr => tr.children[0].textContent);

  // First, update enabled state
  let res = await fetch(`/api/linkfilter?chat_id=${chat_id}`, {
    method: 'POST',
    headers: {'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled }),
  });
  if (!res.ok) throw new Error("Failed to update Link Filter enabled state");

  // Then, sync domains individually to backend
  // To mirror your command logic, add/remove domains one by one:
  // Load current domains to detect changes
  res = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Failed to reload Link Filter");
  const current = await res.json();
  const currentDomains = new Set(current.allowed_domains);

  // Add new domains
  for (const d of domains) {
    if (!currentDomains.has(d)) {
      const resAdd = await fetch(`/api/linkfilter/allow?chat_id=${chat_id}`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: d }),
      });
      if (!resAdd.ok) throw new Error("Failed to add domain " + d);
    }
  }
  // Remove domains no longer in the UI list
  for (const d of currentDomains) {
    if (!domains.includes(d)) {
      const resRm = await fetch(`/api/linkfilter/remove?chat_id=${chat_id}`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json' },
        body: JSON.stringify({ domain: d }),
      });
      if (!resRm.ok) throw new Error("Failed to remove domain " + d);
    }
  }
}

document.getElementById('save-all').onclick = async () => {
  try {
    await saveBioMode();
    await saveEchoSettings();
    await saveLinkFilter();
    alert("✅ Settings saved successfully!");
  } catch (error) {
    alert("❌ Failed to save settings: " + error);
  }
};
