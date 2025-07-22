let chat_id;

window.onload = async () => {
  const tg = window.Telegram.WebApp;
  const initData = tg.initDataUnsafe;

  tg.expand(); // Maximize UI in Mini App

  // Validate we are inside a group context
  if (!initData || !initData.chat || !initData.chat.id) {
    document.body.innerHTML = "<h3>❌ This app can only be used inside a group.</h3>";
    return;
  }

  chat_id = initData.chat.id;

  try {
    await Promise.all([
      loadBioMode(chat_id),
      loadEcho(chat_id),
      loadLinkFilter(chat_id)
    ]);

    document.getElementById("loading").style.display = "none";
    document.getElementById("settings-container").style.display = "block";
  } catch (err) {
    document.body.innerHTML = `<h3>❗ Failed to load settings. ${err}</h3>`;
  }
}

function loadBioMode(chat_id) {
  return fetch(`/api/biomode?chat_id=${chat_id}`)
    .then(res => res.json())
    .then(data => {
      document.getElementById('biomode-switch').checked = data.enabled;
    });
}

function loadEcho(chat_id) {
  return fetch(`/api/echo?chat_id=${chat_id}`)
    .then(res => res.json())
    .then(data => {
      document.getElementById('echo-input').value = data.echo_text || "";
      document.getElementById('longmode-select').value = data.long_mode || "automatic";
      document.getElementById('longlimit-input').value = data.long_limit || 800;
    });
}

function loadLinkFilter(chat_id) {
  return fetch(`/api/linkfilter?chat_id=${chat_id}`)
    .then(res => res.json())
    .then(data => {
      document.getElementById('linkfilter-switch').checked = data.enabled;
      const body = document.getElementById('allowed-links-body');
      body.innerHTML = '';
      (data.allowed_domains || []).forEach(domain => {
        const row = document.createElement('tr');
        row.innerHTML = `<td>${domain}</td><td><button onclick="removeDomain('${domain}')">❌</button></td>`;
        body.appendChild(row);
      });
    });
}

function removeDomain(domain) {
  const rows = document.querySelectorAll(`#allowed-links-body tr`);
  rows.forEach(row => {
    if (row.children[0].innerText === domain) row.remove();
  });
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input');
  const domain = input.value.trim();
  if (domain) {
    const body = document.getElementById('allowed-links-body');
    const row = document.createElement('tr');
    row.innerHTML = `<td>${domain}</td><td><button onclick="removeDomain('${domain}')">❌</button></td>`;
    body.appendChild(row);
    input.value = '';
  }
};

document.getElementById('save-all').onclick = async () => {
  await Promise.all([
    saveBioMode(),
    saveEcho(),
    saveLinkFilter()
  ]);
  alert("✅ Settings saved for this group.");
};

function saveBioMode() {
  return fetch(`/api/biomode?chat_id=${chat_id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled: document.getElementById('biomode-switch').checked })
  });
}

function saveEcho() {
  return fetch(`/api/echo?chat_id=${chat_id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      echo_text: document.getElementById('echo-input').value,
      long_mode: document.getElementById('longmode-select').value,
      long_limit: parseInt(document.getElementById('longlimit-input').value)
    })
  });
}

function saveLinkFilter() {
  const rows = document.querySelectorAll('#allowed-links-body tr');
  const domains = Array.from(rows).map(row => row.children[0].innerText);
  return fetch(`/api/linkfilter?chat_id=${chat_id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      enabled: document.getElementById('linkfilter-switch').checked,
      allowed_domains: domains
    })
  });
}
