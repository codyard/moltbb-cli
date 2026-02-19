const state = {
  diaries: [],
  prompts: [],
  currentDiaryId: null,
  currentPromptId: null,
  activePromptId: null,
};

const el = (id) => document.getElementById(id);
const mountPath = (() => {
  const path = window.location.pathname || '/';
  if (path.endsWith('/')) {
    return path;
  }
  const last = path.slice(path.lastIndexOf('/') + 1);
  if (last.includes('.')) {
    return path.slice(0, path.lastIndexOf('/') + 1) || '/';
  }
  return `${path}/`;
})();
const apiBase = `${mountPath}api`;

function apiPath(path) {
  if (!path) {
    return apiBase;
  }
  if (path.startsWith('/')) {
    return `${apiBase}${path}`;
  }
  return `${apiBase}/${path}`;
}

async function api(path, options = {}) {
  const res = await fetch(apiPath(path), {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  });
  let data = null;
  try {
    data = await res.json();
  } catch {
    data = null;
  }
  if (!res.ok) {
    const msg = data?.error || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

function setStatus(msg, isError = false) {
  const target = el('globalStatus');
  target.textContent = msg;
  target.style.color = isError ? 'var(--coral)' : '';
}

function escapeHtml(str) {
  return (str || '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function switchTab(name) {
  document.querySelectorAll('.tab-btn').forEach((btn) => {
    btn.classList.toggle('active', btn.dataset.tab === name);
  });
  document.querySelectorAll('.tab-panel').forEach((panel) => {
    panel.classList.toggle('active', panel.id === `tab-${name}`);
  });
}

async function loadState() {
  const data = await api('/state');
  el('statDiaries').textContent = String(data.diaryCount);
  el('statPrompts').textContent = String(data.promptCount);
  el('statActive').textContent = data.activePrompt || '-';
  state.activePromptId = data.activePrompt || '';
  if (!el('genOutput').value) {
    el('genOutput').value = data.defaultOutput || '';
  }
}

function renderDiaryList(items) {
  const container = el('diaryList');
  if (!items.length) {
    container.innerHTML = '<div class="muted">No diary files found.</div>';
    return;
  }
  container.innerHTML = items
    .map((item) => {
      const active = item.id === state.currentDiaryId ? 'active' : '';
      return `
        <article class="item ${active}" data-id="${escapeHtml(item.id)}">
          <h3>${escapeHtml(item.title || item.filename)}</h3>
          <p>${escapeHtml(item.preview || '')}</p>
          <div class="meta">${escapeHtml(item.date || 'n/a')} · ${escapeHtml(item.filename)}</div>
        </article>
      `;
    })
    .join('');

  container.querySelectorAll('.item').forEach((node) => {
    node.addEventListener('click', () => loadDiaryDetail(node.dataset.id));
  });
}

async function loadDiaries() {
  const q = el('diarySearch').value.trim();
  const data = await api(`/diaries?limit=200&q=${encodeURIComponent(q)}`);
  state.diaries = data.items || [];
  if (!state.currentDiaryId && state.diaries[0]) {
    state.currentDiaryId = state.diaries[0].id;
  }
  renderDiaryList(state.diaries);
  if (state.currentDiaryId) {
    await loadDiaryDetail(state.currentDiaryId, false);
  } else {
    el('diaryContent').textContent = 'Select a diary from the left list.';
    el('diaryMeta').textContent = '';
  }
}

async function loadDiaryDetail(id, rerender = true) {
  try {
    const data = await api(`/diaries/${encodeURIComponent(id)}`);
    state.currentDiaryId = data.id;
    if (rerender) {
      renderDiaryList(state.diaries);
    }
    el('diaryMeta').textContent = `${data.date || 'n/a'} · ${data.filename} · ${data.modifiedAt}`;
    el('diaryContent').textContent = data.content || '';
  } catch (err) {
    setStatus(`Load diary failed: ${err.message}`, true);
  }
}

function renderPromptList() {
  const container = el('promptList');
  if (!state.prompts.length) {
    container.innerHTML = '<div class="muted">No prompt templates.</div>';
    return;
  }
  container.innerHTML = state.prompts
    .map((item) => {
      const active = item.id === state.currentPromptId ? 'active' : '';
      const marker = item.active ? 'ACTIVE' : item.enabled ? 'ENABLED' : 'DISABLED';
      return `
        <article class="item ${active}" data-id="${escapeHtml(item.id)}">
          <h3>${escapeHtml(item.name)}</h3>
          <p>${escapeHtml(item.description || '(no description)')}</p>
          <div class="meta">${escapeHtml(marker)} · ${escapeHtml(item.id)}</div>
        </article>
      `;
    })
    .join('');

  container.querySelectorAll('.item').forEach((node) => {
    node.addEventListener('click', () => loadPromptDetail(node.dataset.id));
  });
}

function fillPromptSelector() {
  const target = el('genPrompt');
  target.innerHTML = state.prompts
    .map((item) => {
      const selected = item.active ? 'selected' : '';
      const status = item.enabled ? '' : ' (disabled)';
      return `<option value="${escapeHtml(item.id)}" ${selected}>${escapeHtml(item.name + status)}</option>`;
    })
    .join('');
}

async function loadPrompts() {
  const data = await api('/prompts');
  state.prompts = data.items || [];
  state.activePromptId = data.activePromptId || '';
  if (!state.currentPromptId) {
    state.currentPromptId = state.activePromptId || state.prompts[0]?.id || null;
  }
  renderPromptList();
  fillPromptSelector();
  if (state.currentPromptId) {
    await loadPromptDetail(state.currentPromptId, false);
  }
}

async function loadPromptDetail(id, rerender = true) {
  const data = await api(`/prompts/${encodeURIComponent(id)}`);
  state.currentPromptId = data.id;
  if (rerender) {
    renderPromptList();
  }
  el('promptName').value = data.name || '';
  el('promptDesc').value = data.description || '';
  el('promptContent').value = data.content || '';
  el('promptEnabled').checked = !!data.enabled;
  el('promptMeta').textContent = `${data.id} · ${data.builtin ? 'builtin' : 'custom'} · updated ${data.updatedAt}`;
  el('btnDeletePrompt').disabled = !!data.builtin;
}

async function createBlankPrompt() {
  state.currentPromptId = null;
  renderPromptList();
  el('promptName').value = '';
  el('promptDesc').value = '';
  el('promptContent').value = '';
  el('promptEnabled').checked = true;
  el('promptMeta').textContent = 'Creating new prompt';
  el('btnDeletePrompt').disabled = true;
}

async function savePrompt(event) {
  event.preventDefault();
  const payload = {
    name: el('promptName').value,
    description: el('promptDesc').value,
    content: el('promptContent').value,
    enabled: el('promptEnabled').checked,
  };

  if (!payload.name.trim() || !payload.content.trim()) {
    setStatus('Name and content are required.', true);
    return;
  }

  if (!state.currentPromptId) {
    await api('/prompts', { method: 'POST', body: JSON.stringify(payload) });
    setStatus('Prompt created.');
  } else {
    await api(`/prompts/${encodeURIComponent(state.currentPromptId)}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
    setStatus('Prompt updated.');
  }

  await loadPrompts();
  await loadState();
}

async function activatePrompt() {
  if (!state.currentPromptId) {
    setStatus('Select a prompt first.', true);
    return;
  }
  await api(`/prompts/${encodeURIComponent(state.currentPromptId)}/activate`, { method: 'POST' });
  setStatus(`Activated prompt: ${state.currentPromptId}`);
  await loadPrompts();
  await loadState();
}

async function deletePrompt() {
  if (!state.currentPromptId) {
    return;
  }
  const yes = confirm(`Delete prompt ${state.currentPromptId}?`);
  if (!yes) {
    return;
  }
  await api(`/prompts/${encodeURIComponent(state.currentPromptId)}`, { method: 'DELETE' });
  state.currentPromptId = null;
  setStatus('Prompt deleted.');
  await loadPrompts();
  await loadState();
}

async function generatePacket(event) {
  event.preventDefault();
  const hints = el('genHints').value
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);

  const payload = {
    date: el('genDate').value || '',
    hostname: el('genHost').value || '',
    promptId: el('genPrompt').value || '',
    outputDir: el('genOutput').value || '',
    logSourceHints: hints,
  };

  const data = await api('/generate-packet', {
    method: 'POST',
    body: JSON.stringify(payload),
  });

  el('generateResult').textContent = JSON.stringify(data, null, 2);
  setStatus(`Prompt packet generated: ${data.packetPath}`);
}

async function reindex() {
  const data = await api('/diaries/reindex', { method: 'POST' });
  setStatus(`Reindex completed: ${data.diaryCount} diaries.`);
  await loadDiaries();
  await loadState();
}

function bindEvents() {
  document.querySelectorAll('.tab-btn').forEach((btn) => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
  });

  el('btnReload').addEventListener('click', async () => {
    await bootstrap();
  });

  el('diarySearch').addEventListener('input', () => {
    clearTimeout(bindEvents.diaryTimer);
    bindEvents.diaryTimer = setTimeout(() => {
      loadDiaries().catch((err) => setStatus(`Load diaries failed: ${err.message}`, true));
    }, 220);
  });

  el('btnReindex').addEventListener('click', () => {
    reindex().catch((err) => setStatus(`Reindex failed: ${err.message}`, true));
  });

  el('promptForm').addEventListener('submit', (event) => {
    savePrompt(event).catch((err) => setStatus(`Save prompt failed: ${err.message}`, true));
  });

  el('btnActivatePrompt').addEventListener('click', () => {
    activatePrompt().catch((err) => setStatus(`Activate failed: ${err.message}`, true));
  });

  el('btnDeletePrompt').addEventListener('click', () => {
    deletePrompt().catch((err) => setStatus(`Delete failed: ${err.message}`, true));
  });

  el('btnNewPrompt').addEventListener('click', () => {
    createBlankPrompt().catch((err) => setStatus(`Init prompt failed: ${err.message}`, true));
  });

  el('generateForm').addEventListener('submit', (event) => {
    generatePacket(event).catch((err) => setStatus(`Generate failed: ${err.message}`, true));
  });
}

async function bootstrap() {
  setStatus('Loading...');
  await loadState();
  await loadDiaries();
  await loadPrompts();
  setStatus('Ready.');
}

function initDateDefault() {
  const today = new Date().toISOString().slice(0, 10);
  el('genDate').value = today;
}

document.addEventListener('DOMContentLoaded', async () => {
  bindEvents();
  initDateDefault();
  try {
    await bootstrap();
  } catch (err) {
    setStatus(`Initialization failed: ${err.message}`, true);
  }
});
