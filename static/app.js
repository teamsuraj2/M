window.onload = () => {
  loadBioMode()
  loadEcho()
  loadLinkFilter()
}

function loadBioMode() {
  fetch('/api/biomode')
    .then(res => res.json())
    .then(data => {
      document.getElementById('biomode-switch').checked = data.enabled
    })
}

function loadEcho() {
  fetch('/api/echo')
    .then(res => res.json())
    .then(data => {
      document.getElementById('echo-input').value = data.echo_text
      document.getElementById('longmode-select').value = data.long_mode
      document.getElementById('longlimit-input').value = data.long_limit
    })
}

function loadLinkFilter() {
  fetch('/api/linkfilter')
    .then(res => res.json())
    .then(data => {
      document.getElementById('linkfilter-switch').checked = data.enabled
      const body = document.getElementById('allowed-links-body')
      body.innerHTML = ''
      data.allowed_domains.forEach(domain => {
        const row = document.createElement('tr')
        row.innerHTML = `<td>${domain}</td><td><button onclick="removeDomain('${domain}')">X</button></td>`
        body.appendChild(row)
      })
    })
}

function removeDomain(domain) {
  const rows = document.querySelectorAll(`#allowed-links-body tr`)
  rows.forEach(row => {
    if (row.children[0].innerText === domain) row.remove()
  })
}

document.getElementById('allow-btn').onclick = () => {
  const input = document.getElementById('allow-domain-input')
  const domain = input.value.trim()
  if (domain) {
    const body = document.getElementById('allowed-links-body')
    const row = document.createElement('tr')
    row.innerHTML = `<td>${domain}</td><td><button onclick="removeDomain('${domain}')">X</button></td>`
    body.appendChild(row)
    input.value = ''
  }
}

document.getElementById('save-all').onclick = () => {
  saveBioMode()
  saveEcho()
  saveLinkFilter()
  .then(() => alert("Settings saved"))
}

function saveBioMode() {
  return fetch('/api/biomode', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      enabled: document.getElementById('biomode-switch').checked
    })
  })
}

function saveEcho() {
  return fetch('/api/echo', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      echo_text: document.getElementById('echo-input').value,
      long_mode: document.getElementById('longmode-select').value,
      long_limit: parseInt(document.getElementById('longlimit-input').value)
    })
  })
}

function saveLinkFilter() {
  const rows = document.querySelectorAll('#allowed-links-body tr')
  const domains = Array.from(rows).map(row => row.children[0].innerText)
  return fetch('/api/linkfilter', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      enabled: document.getElementById('linkfilter-switch').checked,
      allowed_domains: domains
    })
  })
}
