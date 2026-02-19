const LOCALE_STORAGE_KEY = 'moltbb.localweb.locale';
const SUPPORTED_LOCALES = ['en', 'zh-Hans'];

const MESSAGES = {
  en: {
    'page.title': 'MoltBB Local Diary Studio',
    'topbar.title': 'Diary Studio',
    'topbar.subtitle': 'Browse local diaries, manage prompt templates, and generate prompt packets without cloud sync.',
    'lang.label': 'Language',
    'stats.diaries': 'Diaries',
    'stats.prompts': 'Prompts',
    'stats.active': 'Active',
    'tabs.diaries': 'Diaries',
    'tabs.prompts': 'Prompts',
    'tabs.generate': 'Generate Packet',
    'actions.refresh': 'Refresh',
    'actions.reindex': 'Reindex',
    'actions.new': 'New',
    'actions.save': 'Save',
    'actions.setActive': 'Set Active',
    'actions.delete': 'Delete',
    'actions.generate': 'Generate',
    'diary.listTitle': 'Diary List',
    'diary.detailTitle': 'Diary Detail',
    'diary.searchPlaceholder': 'Search by title / date / filename',
    'diary.selectHint': 'Select a diary from the left list.',
    'diary.noFiles': 'No diary files found.',
    'diary.metaFormat': '{date} · {filename} · {modifiedAt}',
    'diary.viewReading': 'Reading',
    'diary.viewRaw': 'Raw',
    'diary.loadFailed': 'Load diary failed: {message}',
    'diary.loadListFailed': 'Load diaries failed: {message}',
    'prompt.listTitle': 'Prompt Templates',
    'prompt.editorTitle': 'Prompt Editor',
    'prompt.name': 'Name',
    'prompt.description': 'Description',
    'prompt.content': 'Content',
    'prompt.enabled': 'Enabled',
    'prompt.noTemplates': 'No prompt templates.',
    'prompt.noDescription': '(no description)',
    'prompt.markerActive': 'ACTIVE',
    'prompt.markerEnabled': 'ENABLED',
    'prompt.markerDisabled': 'DISABLED',
    'prompt.typeBuiltin': 'builtin',
    'prompt.typeCustom': 'custom',
    'prompt.metaFormat': '{id} · {kind} · updated {updatedAt}',
    'prompt.createNew': 'Creating new prompt',
    'prompt.disabledSuffix': ' (disabled)',
    'prompt.nameContentRequired': 'Name and content are required.',
    'prompt.created': 'Prompt created.',
    'prompt.updated': 'Prompt updated.',
    'prompt.selectFirst': 'Select a prompt first.',
    'prompt.activated': 'Activated prompt: {id}',
    'prompt.deleted': 'Prompt deleted.',
    'prompt.deleteConfirm': 'Delete prompt {id}?',
    'prompt.saveFailed': 'Save prompt failed: {message}',
    'prompt.activateFailed': 'Activate failed: {message}',
    'prompt.deleteFailed': 'Delete failed: {message}',
    'prompt.initFailed': 'Init prompt failed: {message}',
    'generate.title': 'Generate Prompt Packet',
    'generate.date': 'Date (UTC)',
    'generate.hostname': 'Hostname',
    'generate.prompt': 'Prompt',
    'generate.outputDir': 'Output Directory',
    'generate.hints': 'Log Source Hints (one per line)',
    'generate.outputPlaceholder': 'default diary dir',
    'generate.hintsPlaceholder': '~/.openclaw/logs/work.log',
    'generate.noPacket': 'No packet generated yet.',
    'generate.generated': 'Prompt packet generated: {path}',
    'generate.failed': 'Generate failed: {message}',
    'reindex.done': 'Reindex completed: {count} diaries.',
    'reindex.failed': 'Reindex failed: {message}',
    'status.loading': 'Loading...',
    'status.ready': 'Ready.',
    'status.initFailed': 'Initialization failed: {message}',
    'common.optional': 'optional',
    'common.na': 'n/a',
  },
  'zh-Hans': {
    'page.title': 'MoltBB 本地日记工作台',
    'topbar.title': '日记工作台',
    'topbar.subtitle': '浏览本地日记、管理提示词模板，并在不走云同步的情况下生成提示词数据包。',
    'lang.label': '语言',
    'stats.diaries': '日记',
    'stats.prompts': '模板',
    'stats.active': '当前激活',
    'tabs.diaries': '日记',
    'tabs.prompts': '提示词',
    'tabs.generate': '生成数据包',
    'actions.refresh': '刷新',
    'actions.reindex': '重建索引',
    'actions.new': '新建',
    'actions.save': '保存',
    'actions.setActive': '设为激活',
    'actions.delete': '删除',
    'actions.generate': '生成',
    'diary.listTitle': '日记列表',
    'diary.detailTitle': '日记详情',
    'diary.searchPlaceholder': '按标题 / 日期 / 文件名搜索',
    'diary.selectHint': '请从左侧列表选择一篇日记。',
    'diary.noFiles': '未找到日记文件。',
    'diary.metaFormat': '{date} · {filename} · {modifiedAt}',
    'diary.viewReading': '阅读模式',
    'diary.viewRaw': '原文模式',
    'diary.loadFailed': '加载日记失败: {message}',
    'diary.loadListFailed': '加载日记列表失败: {message}',
    'prompt.listTitle': '提示词模板',
    'prompt.editorTitle': '提示词编辑器',
    'prompt.name': '名称',
    'prompt.description': '描述',
    'prompt.content': '内容',
    'prompt.enabled': '启用',
    'prompt.noTemplates': '暂无提示词模板。',
    'prompt.noDescription': '（无描述）',
    'prompt.markerActive': '激活中',
    'prompt.markerEnabled': '已启用',
    'prompt.markerDisabled': '已禁用',
    'prompt.typeBuiltin': '内置',
    'prompt.typeCustom': '自定义',
    'prompt.metaFormat': '{id} · {kind} · 更新于 {updatedAt}',
    'prompt.createNew': '正在创建新模板',
    'prompt.disabledSuffix': '（已禁用）',
    'prompt.nameContentRequired': '名称和内容不能为空。',
    'prompt.created': '提示词已创建。',
    'prompt.updated': '提示词已更新。',
    'prompt.selectFirst': '请先选择一个提示词。',
    'prompt.activated': '已激活提示词: {id}',
    'prompt.deleted': '提示词已删除。',
    'prompt.deleteConfirm': '确认删除提示词 {id} 吗？',
    'prompt.saveFailed': '保存提示词失败: {message}',
    'prompt.activateFailed': '激活失败: {message}',
    'prompt.deleteFailed': '删除失败: {message}',
    'prompt.initFailed': '初始化提示词失败: {message}',
    'generate.title': '生成提示词数据包',
    'generate.date': '日期 (UTC)',
    'generate.hostname': '主机名',
    'generate.prompt': '提示词',
    'generate.outputDir': '输出目录',
    'generate.hints': '日志来源提示（每行一个）',
    'generate.outputPlaceholder': '默认日记目录',
    'generate.hintsPlaceholder': '~/.openclaw/logs/work.log',
    'generate.noPacket': '尚未生成数据包。',
    'generate.generated': '已生成提示词数据包: {path}',
    'generate.failed': '生成失败: {message}',
    'reindex.done': '索引重建完成: {count} 篇日记。',
    'reindex.failed': '重建索引失败: {message}',
    'status.loading': '加载中...',
    'status.ready': '就绪。',
    'status.initFailed': '初始化失败: {message}',
    'common.optional': '可选',
    'common.na': '无',
  },
};

const state = {
  diaries: [],
  prompts: [],
  currentDiaryId: null,
  currentPromptId: null,
  activePromptId: null,
  diaryRawContent: '',
  diaryViewMode: 'raw',
  currentDiaryDetail: null,
  currentPromptDetail: null,
  locale: 'en',
  hasGeneratedPacket: false,
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

function resolveLocale(raw) {
  if (typeof raw !== 'string') {
    return 'en';
  }
  const trimmed = raw.trim();
  if (!trimmed) {
    return 'en';
  }
  if (trimmed === 'zh-Hans' || trimmed === 'zh_CN' || trimmed === 'zh-CN' || trimmed.toLowerCase().startsWith('zh')) {
    return 'zh-Hans';
  }
  return 'en';
}

function messageFor(locale, key) {
  const pack = MESSAGES[locale] || MESSAGES.en;
  if (Object.prototype.hasOwnProperty.call(pack, key)) {
    return pack[key];
  }
  if (Object.prototype.hasOwnProperty.call(MESSAGES.en, key)) {
    return MESSAGES.en[key];
  }
  return key;
}

function t(key, params = {}) {
  const template = messageFor(state.locale, key);
  return template.replace(/\{(\w+)\}/g, (_, token) => {
    if (Object.prototype.hasOwnProperty.call(params, token)) {
      return String(params[token]);
    }
    return '';
  });
}

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

function setStatusKey(key, params = {}, isError = false) {
  setStatus(t(key, params), isError);
}

function escapeHtml(str) {
  return (str || '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function renderInlineMarkdown(text) {
  let out = escapeHtml(text || '');
  out = out.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>');
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  out = out.replace(/(^|[^*])\*([^*]+)\*(?=[^*]|$)/g, '$1<em>$2</em>');
  return out;
}

function markdownToHtml(markdown) {
  const lines = (markdown || '').replaceAll('\r\n', '\n').replaceAll('\r', '\n').split('\n');
  const html = [];
  let i = 0;
  let inCode = false;
  let codeLines = [];
  let paragraph = [];
  let listType = '';
  let listItems = [];
  let quoteLines = [];

  const flushParagraph = () => {
    if (!paragraph.length) {
      return;
    }
    html.push(`<p>${renderInlineMarkdown(paragraph.join(' '))}</p>`);
    paragraph = [];
  };

  const flushList = () => {
    if (!listItems.length) {
      return;
    }
    const tag = listType === 'ol' ? 'ol' : 'ul';
    const items = listItems.map((item) => `<li>${renderInlineMarkdown(item)}</li>`).join('');
    html.push(`<${tag}>${items}</${tag}>`);
    listType = '';
    listItems = [];
  };

  const flushQuote = () => {
    if (!quoteLines.length) {
      return;
    }
    html.push(`<blockquote>${renderInlineMarkdown(quoteLines.join(' '))}</blockquote>`);
    quoteLines = [];
  };

  while (i < lines.length) {
    const line = lines[i];
    const trimmed = line.trim();

    if (inCode) {
      if (trimmed.startsWith('```')) {
        const codeHtml = escapeHtml(codeLines.join('\n'));
        html.push(`<pre><code>${codeHtml}</code></pre>`);
        inCode = false;
        codeLines = [];
      } else {
        codeLines.push(line);
      }
      i += 1;
      continue;
    }

    if (trimmed === '') {
      flushParagraph();
      flushList();
      flushQuote();
      i += 1;
      continue;
    }

    if (trimmed.startsWith('```')) {
      flushParagraph();
      flushList();
      flushQuote();
      inCode = true;
      codeLines = [];
      i += 1;
      continue;
    }

    const heading = trimmed.match(/^(#{1,4})\s+(.+)$/);
    if (heading) {
      flushParagraph();
      flushList();
      flushQuote();
      const level = heading[1].length;
      html.push(`<h${level}>${renderInlineMarkdown(heading[2])}</h${level}>`);
      i += 1;
      continue;
    }

    const quote = trimmed.match(/^>\s?(.*)$/);
    if (quote) {
      flushParagraph();
      flushList();
      quoteLines.push(quote[1]);
      i += 1;
      continue;
    }

    const ul = trimmed.match(/^[-*+]\s+(.+)$/);
    if (ul) {
      flushParagraph();
      flushQuote();
      if (listType !== 'ul') {
        flushList();
        listType = 'ul';
      }
      listItems.push(ul[1]);
      i += 1;
      continue;
    }

    const ol = trimmed.match(/^\d+\.\s+(.+)$/);
    if (ol) {
      flushParagraph();
      flushQuote();
      if (listType !== 'ol') {
        flushList();
        listType = 'ol';
      }
      listItems.push(ol[1]);
      i += 1;
      continue;
    }

    if (/^---+$/.test(trimmed) || /^___+$/.test(trimmed)) {
      flushParagraph();
      flushList();
      flushQuote();
      html.push('<hr />');
      i += 1;
      continue;
    }

    flushList();
    flushQuote();
    paragraph.push(trimmed);
    i += 1;
  }

  if (inCode) {
    const codeHtml = escapeHtml(codeLines.join('\n'));
    html.push(`<pre><code>${codeHtml}</code></pre>`);
  }
  flushParagraph();
  flushList();
  flushQuote();

  return html.join('\n');
}

function applyDiaryViewModeButton() {
  const button = el('btnDiaryViewMode');
  if (!button) {
    return;
  }
  if (state.diaryViewMode === 'raw') {
    button.textContent = t('diary.viewReading');
    button.dataset.mode = 'raw';
  } else {
    button.textContent = t('diary.viewRaw');
    button.dataset.mode = 'reading';
  }
}

function renderDiaryContent() {
  const target = el('diaryContent');
  const content = state.diaryRawContent || '';
  if (state.diaryViewMode === 'markdown') {
    target.classList.add('markdown-view');
    target.innerHTML = markdownToHtml(content);
  } else {
    target.classList.remove('markdown-view');
    target.textContent = content;
  }
  applyDiaryViewModeButton();
}

function renderDiaryMeta() {
  if (!state.currentDiaryDetail) {
    el('diaryMeta').textContent = '';
    return;
  }
  const detail = state.currentDiaryDetail;
  el('diaryMeta').textContent = t('diary.metaFormat', {
    date: detail.date || t('common.na'),
    filename: detail.filename || '',
    modifiedAt: detail.modifiedAt || '',
  });
}

function setDiaryEmptyState() {
  state.currentDiaryDetail = null;
  state.diaryRawContent = t('diary.selectHint');
  renderDiaryContent();
  renderDiaryMeta();
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
    container.innerHTML = `<div class="muted">${escapeHtml(t('diary.noFiles'))}</div>`;
    return;
  }
  container.innerHTML = items
    .map((item) => {
      const active = item.id === state.currentDiaryId ? 'active' : '';
      return `
        <article class="item ${active}" data-id="${escapeHtml(item.id)}">
          <h3>${escapeHtml(item.title || item.filename)}</h3>
          <p>${escapeHtml(item.preview || '')}</p>
          <div class="meta">${escapeHtml(item.date || t('common.na'))} · ${escapeHtml(item.filename)}</div>
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
    setDiaryEmptyState();
  }
}

async function loadDiaryDetail(id, rerender = true) {
  try {
    const data = await api(`/diaries/${encodeURIComponent(id)}`);
    state.currentDiaryId = data.id;
    state.currentDiaryDetail = data;
    state.diaryRawContent = data.content || '';
    if (rerender) {
      renderDiaryList(state.diaries);
    }
    renderDiaryMeta();
    renderDiaryContent();
  } catch (err) {
    setStatusKey('diary.loadFailed', { message: err.message }, true);
  }
}

function promptMarker(item) {
  if (item.active) {
    return t('prompt.markerActive');
  }
  if (item.enabled) {
    return t('prompt.markerEnabled');
  }
  return t('prompt.markerDisabled');
}

function renderPromptList() {
  const container = el('promptList');
  if (!state.prompts.length) {
    container.innerHTML = `<div class="muted">${escapeHtml(t('prompt.noTemplates'))}</div>`;
    return;
  }
  container.innerHTML = state.prompts
    .map((item) => {
      const active = item.id === state.currentPromptId ? 'active' : '';
      const marker = promptMarker(item);
      return `
        <article class="item ${active}" data-id="${escapeHtml(item.id)}">
          <h3>${escapeHtml(item.name)}</h3>
          <p>${escapeHtml(item.description || t('prompt.noDescription'))}</p>
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
      const status = item.enabled ? '' : t('prompt.disabledSuffix');
      return `<option value="${escapeHtml(item.id)}" ${selected}>${escapeHtml(item.name + status)}</option>`;
    })
    .join('');
}

function updatePromptMeta() {
  if (!state.currentPromptDetail) {
    if (!state.currentPromptId) {
      el('promptMeta').textContent = t('prompt.createNew');
    }
    return;
  }
  const data = state.currentPromptDetail;
  const kind = data.builtin ? t('prompt.typeBuiltin') : t('prompt.typeCustom');
  el('promptMeta').textContent = t('prompt.metaFormat', {
    id: data.id,
    kind,
    updatedAt: data.updatedAt || '',
  });
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
  state.currentPromptDetail = data;
  if (rerender) {
    renderPromptList();
  }
  el('promptName').value = data.name || '';
  el('promptDesc').value = data.description || '';
  el('promptContent').value = data.content || '';
  el('promptEnabled').checked = !!data.enabled;
  updatePromptMeta();
  el('btnDeletePrompt').disabled = !!data.builtin;
}

async function createBlankPrompt() {
  state.currentPromptId = null;
  state.currentPromptDetail = null;
  renderPromptList();
  el('promptName').value = '';
  el('promptDesc').value = '';
  el('promptContent').value = '';
  el('promptEnabled').checked = true;
  el('promptMeta').textContent = t('prompt.createNew');
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
    setStatusKey('prompt.nameContentRequired', {}, true);
    return;
  }

  if (!state.currentPromptId) {
    await api('/prompts', { method: 'POST', body: JSON.stringify(payload) });
    setStatusKey('prompt.created');
  } else {
    await api(`/prompts/${encodeURIComponent(state.currentPromptId)}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
    setStatusKey('prompt.updated');
  }

  await loadPrompts();
  await loadState();
}

async function activatePrompt() {
  if (!state.currentPromptId) {
    setStatusKey('prompt.selectFirst', {}, true);
    return;
  }
  await api(`/prompts/${encodeURIComponent(state.currentPromptId)}/activate`, { method: 'POST' });
  setStatusKey('prompt.activated', { id: state.currentPromptId });
  await loadPrompts();
  await loadState();
}

async function deletePrompt() {
  if (!state.currentPromptId) {
    return;
  }
  const yes = confirm(t('prompt.deleteConfirm', { id: state.currentPromptId }));
  if (!yes) {
    return;
  }
  await api(`/prompts/${encodeURIComponent(state.currentPromptId)}`, { method: 'DELETE' });
  state.currentPromptId = null;
  state.currentPromptDetail = null;
  setStatusKey('prompt.deleted');
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
  state.hasGeneratedPacket = true;
  setStatusKey('generate.generated', { path: data.packetPath });
}

async function reindex() {
  const data = await api('/diaries/reindex', { method: 'POST' });
  setStatusKey('reindex.done', { count: data.diaryCount });
  await loadDiaries();
  await loadState();
}

function applyStaticI18n() {
  document.title = t('page.title');
  document.documentElement.lang = state.locale;

  document.querySelectorAll('[data-i18n]').forEach((node) => {
    const key = node.dataset.i18n;
    if (!key) {
      return;
    }
    node.textContent = t(key);
  });

  document.querySelectorAll('[data-i18n-placeholder]').forEach((node) => {
    const key = node.dataset.i18nPlaceholder;
    if (!key) {
      return;
    }
    node.setAttribute('placeholder', t(key));
  });
}

function refreshLocalizedDynamicText() {
  renderDiaryList(state.diaries);
  if (state.currentDiaryDetail) {
    renderDiaryMeta();
    renderDiaryContent();
  } else {
    setDiaryEmptyState();
  }

  renderPromptList();
  fillPromptSelector();
  updatePromptMeta();

  if (!state.hasGeneratedPacket) {
    el('generateResult').textContent = t('generate.noPacket');
  }

  applyDiaryViewModeButton();
}

function setLocale(locale, persist = true) {
  const next = resolveLocale(locale);
  state.locale = next;
  const picker = el('langSwitch');
  if (picker && picker.value !== next) {
    picker.value = next;
  }
  if (persist) {
    window.localStorage.setItem(LOCALE_STORAGE_KEY, next);
  }
  applyStaticI18n();
  refreshLocalizedDynamicText();
}

function initLocale() {
  const saved = window.localStorage.getItem(LOCALE_STORAGE_KEY);
  const locale = resolveLocale(saved || 'en');
  const picker = el('langSwitch');
  if (picker) {
    picker.value = SUPPORTED_LOCALES.includes(locale) ? locale : 'en';
  }
  setLocale(locale, false);
}

function bindEvents() {
  document.querySelectorAll('.tab-btn').forEach((btn) => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
  });

  el('langSwitch').addEventListener('change', (event) => {
    setLocale(event.target.value, true);
  });

  el('btnReload').addEventListener('click', async () => {
    await bootstrap();
  });

  el('diarySearch').addEventListener('input', () => {
    clearTimeout(bindEvents.diaryTimer);
    bindEvents.diaryTimer = setTimeout(() => {
      loadDiaries().catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    }, 220);
  });

  el('btnReindex').addEventListener('click', () => {
    reindex().catch((err) => setStatusKey('reindex.failed', { message: err.message }, true));
  });

  el('promptForm').addEventListener('submit', (event) => {
    savePrompt(event).catch((err) => setStatusKey('prompt.saveFailed', { message: err.message }, true));
  });

  el('btnActivatePrompt').addEventListener('click', () => {
    activatePrompt().catch((err) => setStatusKey('prompt.activateFailed', { message: err.message }, true));
  });

  el('btnDeletePrompt').addEventListener('click', () => {
    deletePrompt().catch((err) => setStatusKey('prompt.deleteFailed', { message: err.message }, true));
  });

  el('btnNewPrompt').addEventListener('click', () => {
    createBlankPrompt().catch((err) => setStatusKey('prompt.initFailed', { message: err.message }, true));
  });

  el('generateForm').addEventListener('submit', (event) => {
    generatePacket(event).catch((err) => setStatusKey('generate.failed', { message: err.message }, true));
  });

  el('btnDiaryViewMode').addEventListener('click', () => {
    state.diaryViewMode = state.diaryViewMode === 'raw' ? 'markdown' : 'raw';
    renderDiaryContent();
  });
}

async function bootstrap() {
  setStatusKey('status.loading');
  await loadState();
  await loadDiaries();
  await loadPrompts();
  setStatusKey('status.ready');
}

function initDateDefault() {
  const today = new Date().toISOString().slice(0, 10);
  el('genDate').value = today;
}

document.addEventListener('DOMContentLoaded', async () => {
  initLocale();
  bindEvents();
  initDateDefault();
  setDiaryEmptyState();
  el('generateResult').textContent = t('generate.noPacket');
  applyDiaryViewModeButton();
  try {
    await bootstrap();
  } catch (err) {
    setStatusKey('status.initFailed', { message: err.message }, true);
  }
});
