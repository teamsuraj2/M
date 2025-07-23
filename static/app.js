const tg = window.Telegram.WebApp;
let chat_id;

window.onload = async () => {
  const initData = tg?.initDataUnsafe;
  const user = initData?.user ?? null;
  const chat = initData?.chat ?? null;
  const initRaw = tg?.initData;

  if (!chat?.id) {
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

    document.body.innerHTML = `
      <div class="container">
        <h3>❌ This app must be opened from a Telegram <b>group chat</b>.</h3>
        <p>Please open your group > Menu > Mini Apps > Settings.</p>
      </div>
    `;
    return;
  }

  chat_id = chat.id;
  tg.expand();

  try {
    await Promise.all([
      loadBioMode(),
      loadEchoSettings(),
      loadLinkFilter()
    ]);
    document.getElementById("loading").style.display = "none";
    document.getElementById("settings-container").style.display = "block";
  } catch (e) {
    document.body.innerHTML = `<h3>❗ Failed to load settings: ${e?.message ?? e}</h3>`;
  }
};

// ----------------------- BioMode -----------------------
async function loadBioMode() {
  const res = await fetch(`/api/biomode?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load BioMode");
  const enabled = await res.json();
  document.getElementById('biomode-switch').checked = !!enabled;
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
  const res = await fetch(`/api/echo?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load Echo Settings");

  const data = await res.json();
  document.getElementById('longmode-select').value = data?.long_mode ?? 'automatic';
  document.getElementById('longlimit-input').value = data?.long_limit ?? 800;
}

function saveEchoSettings() {
  const long_mode = document.getElementById('longmode-select').value;
  let long_limit = parseInt(document.getElementById('longlimit-input').value, 10);
  if (isNaN(long_limit) || long_limit < 200 || long_limit > 4000) {
    alert("⚠️ Long limit must be between 200 and 4000");
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
  const res = await fetch(`/api/linkfilter?chat_id=${chat_id}`);
  if (!res.ok) throw new Error("Could not load LinkFilter");

  const data = await res.json();
  document.getElementById('linkfilter-switch').checked = !!data?.enabled;
  const domains = data?.allowed_domains ?? [];

  const tbody = document.getElementById('allowed-links-body');
  tbody.innerHTML = '';
  domains.forEach(domain => addDomainRow(domain));
}

function addDomainRow(domain) {
  const tbody = document.getElementById('allowed-links-body');
  if ([...tbody.children].some(tr => tr.children[0].textContent === domain)) return;

  const tr = document.createElement('tr');
  tr.innerHTML = `
    <td>${domain}</td>
    <td><button class="mdui-btn mdui-btn-icon" aria-label="Remove">❌</button></td>
  `;
  tr.querySelector('button').onclick = () => tbody.removeChild(tr);
  tbody.appendChild(tr);
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input');
  const domain = input?.value?.trim()?.toLowerCase();
  if (domain) {
    addDomainRow(domain);
    input.value = '';
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

// ----------------------- Save All -----------------------
document.getElementById('save-all').onclick = async () => {
  try {
    await saveBioMode();
    await saveEchoSettings();
    await saveLinkFilter();
    alert("✅ Settings saved successfully!");
  } catch (error) {
    alert("❌ Failed to save: " + (error?.message ?? error));
  }
};
