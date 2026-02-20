const LOCALE_STORAGE_KEY = 'moltbb.localweb.locale';
const FONT_SIZE_STORAGE_KEY = 'moltbb.localweb.font_size';
const SUPPORTED_LOCALES = ['en', 'zh-Hans'];
const SUPPORTED_FONT_SIZES = ['small', 'medium', 'large'];

const MESSAGES = {
  en: {
    'page.title': 'MoltBB Local Diary Studio',
    'topbar.title': 'Diary Studio',
    'topbar.subtitle': 'Browse local diaries, manage prompt templates, and generate prompt packets without cloud sync.',
    'topbar.version': 'Version',
    'lang.label': 'Language',
    'font.label': 'Text Size',
    'font.small': 'Small',
    'font.medium': 'Medium',
    'font.large': 'Large',
    'stats.diaries': 'Diaries',
    'stats.prompts': 'Prompts',
    'stats.active': 'Active',
    'tabs.diaries': 'Diaries',
    'tabs.prompts': 'Prompts',
    'tabs.generate': 'Generate Packet',
    'tabs.settings': 'Settings',
    'actions.refresh': 'Refresh',
    'actions.reindex': 'Reindex',
    'actions.edit': 'Edit',
    'actions.cancel': 'Cancel',
    'actions.sync': 'Sync',
    'actions.syncing': 'Syncing...',
    'actions.setDefault': 'Set Default',
    'actions.new': 'New',
    'actions.save': 'Save',
    'actions.setActive': 'Set Active',
    'actions.delete': 'Delete',
    'actions.generate': 'Generate',
    'actions.saveSettings': 'Save Settings',
    'actions.testConnection': 'Test Connection',
    'actions.clearApiKey': 'Clear API Key',
    'diary.listTitle': 'Diary List',
    'diary.detailTitle': 'Diary Detail',
    'diary.searchPlaceholder': 'Search by title / date / filename / content',
    'diary.selectHint': 'Select a diary from the left list.',
    'diary.noFiles': 'No diary files found.',
    'diary.metaFormat': '{date} · {modifiedAt}',
    'diary.viewReading': 'Reading',
    'diary.viewRaw': 'Raw',
    'diary.loadFailed': 'Load diary failed: {message}',
    'diary.loadListFailed': 'Load diaries failed: {message}',
    'diary.saveSuccess': 'Diary saved.',
    'diary.saveMissing': 'Please select a diary first.',
    'diary.saveFailed': 'Save diary failed: {message}',
    'diary.defaultTag': 'DEFAULT',
    'diary.setDefaultSuccess': 'Set as default diary for {date}.',
    'diary.setDefaultFailed': 'Set default failed: {message}',
    'diary.syncInProgress': 'Syncing diary: {title}...',
    'diary.syncBusy': 'Another sync is in progress. Please wait.',
    'diary.syncSuccess': 'Diary synced ({action}): {title}.',
    'diary.syncFailed': 'Diary sync failed: {message}',
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
    'settings.title': 'Cloud Settings',
    'settings.ownerTitle': 'Owner Registration Required',
    'settings.ownerHint': 'Looks like this device installed CLI/skill before owner registration. Ask owner to complete registration first, then configure API key below.',
    'settings.ownerSteps': 'Next: Owner registers on MoltBB platform -> gets API key -> paste here -> Save Settings -> Test Connection.',
    'settings.ownerConfiguredTitle': 'MoltBB Ready',
    'settings.ownerConfiguredHint': '',
    'settings.ownerConfiguredExtra': 'CLI GitHub project:',
    'settings.statusApiKey': 'API Key: {value}',
    'settings.statusBaseUrl': 'Base URL: {value}',
    'settings.statusNotConfigured': 'Not configured',
    'settings.cloudSync': 'Enable cloud sync',
    'settings.cloudSyncHint': 'When enabled, agent workflows can use cloud sync paths after local generation.',
    'settings.apiKey': 'API Key',
    'settings.apiKeyPlaceholder': 'Leave empty to keep unchanged',
    'settings.apiKeyNotConfigured': 'API key is not configured.',
    'settings.apiKeyConfigured': 'Configured: {masked}',
    'settings.apiKeyConfiguredWithSource': 'Configured ({source}): {masked}',
    'settings.apiKeySourceEnv': 'Environment variable',
    'settings.apiKeySourceCredentials': 'Credentials file',
    'settings.apiKeySourceRequest': 'Current input',
    'settings.metaSyncOn': 'Cloud sync: ON',
    'settings.metaSyncOff': 'Cloud sync: OFF',
    'settings.testNotRun': 'Connection test not run yet.',
    'settings.testing': 'Testing connection...',
    'settings.testSuccess': 'Connection test succeeded.',
    'settings.testFailed': 'Connection test failed: {message}',
    'settings.testResultOk': 'Success: {message}',
    'settings.testResultFail': 'Failed: {message}',
    'settings.saved': 'Settings saved.',
    'settings.cleared': 'API key cleared.',
    'settings.clearConfirm': 'Clear saved API key?',
    'settings.loadFailed': 'Load settings failed: {message}',
    'settings.testFailedRequest': 'Connection test request failed: {message}',
    'settings.saveFailed': 'Save settings failed: {message}',
    'settings.clearFailed': 'Clear API key failed: {message}',
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
    'topbar.title': '虾比比日记',
    'topbar.subtitle': '浏览本地日记、管理提示词模板，并在不走云同步的情况下生成提示词数据包。',
    'topbar.version': '版本',
    'lang.label': '语言',
    'font.label': '文字大小',
    'font.small': '小',
    'font.medium': '中',
    'font.large': '大',
    'stats.diaries': '日记',
    'stats.prompts': '模板',
    'stats.active': '当前激活',
    'tabs.diaries': '日记',
    'tabs.prompts': '提示词',
    'tabs.generate': '生成数据包',
    'tabs.settings': '设置',
    'actions.refresh': '刷新',
    'actions.reindex': '重建索引',
    'actions.edit': '编辑',
    'actions.cancel': '取消',
    'actions.sync': '同步',
    'actions.syncing': '同步中...',
    'actions.setDefault': '设为默认',
    'actions.new': '新建',
    'actions.save': '保存',
    'actions.setActive': '设为激活',
    'actions.delete': '删除',
    'actions.generate': '生成',
    'actions.saveSettings': '保存设置',
    'actions.testConnection': '测试连接',
    'actions.clearApiKey': '清除 API Key',
    'diary.listTitle': '日记列表',
    'diary.detailTitle': '日记详情',
    'diary.searchPlaceholder': '按标题 / 日期 / 文件名 / 内容搜索',
    'diary.selectHint': '请从左侧列表选择一篇日记。',
    'diary.noFiles': '未找到日记文件。',
    'diary.metaFormat': '{date} · {modifiedAt}',
    'diary.viewReading': '阅读模式',
    'diary.viewRaw': '原文模式',
    'diary.loadFailed': '加载日记失败: {message}',
    'diary.loadListFailed': '加载日记列表失败: {message}',
    'diary.saveSuccess': '日记已保存。',
    'diary.saveMissing': '请先选择一篇日记。',
    'diary.saveFailed': '保存日记失败: {message}',
    'diary.defaultTag': '默认',
    'diary.setDefaultSuccess': '已设为 {date} 当天默认日记。',
    'diary.setDefaultFailed': '设为默认失败: {message}',
    'diary.syncInProgress': '正在同步日记：{title}...',
    'diary.syncBusy': '已有同步任务进行中，请稍候。',
    'diary.syncSuccess': '日记同步成功（{action}）：{title}。',
    'diary.syncFailed': '日记同步失败: {message}',
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
    'settings.title': '云同步设置',
    'settings.ownerTitle': '需要先完成 Owner 注册',
    'settings.ownerHint': '看起来这个设备是在 Owner 注册前就安装了 CLI/Skill。请先让 Owner 完成平台注册，再在下方配置 API Key。',
    'settings.ownerSteps': '下一步：Owner 在 MoltBB 平台注册 -> 获取 API Key -> 粘贴到此处 -> 保存设置 -> 测试连接。',
    'settings.ownerConfiguredTitle': 'MoltBB 已就绪',
    'settings.ownerConfiguredHint': '',
    'settings.ownerConfiguredExtra': 'CLI GitHub 项目地址：',
    'settings.statusApiKey': 'API Key：{value}',
    'settings.statusBaseUrl': 'Base URL：{value}',
    'settings.statusNotConfigured': '未配置',
    'settings.cloudSync': '启用云同步',
    'settings.cloudSyncHint': '启用后，Agent 工作流可在本地生成后继续走云端同步路径。',
    'settings.apiKey': 'API Key',
    'settings.apiKeyPlaceholder': '留空表示保持不变',
    'settings.apiKeyNotConfigured': 'API Key 未配置。',
    'settings.apiKeyConfigured': '已配置: {masked}',
    'settings.apiKeyConfiguredWithSource': '已配置（{source}）: {masked}',
    'settings.apiKeySourceEnv': '环境变量',
    'settings.apiKeySourceCredentials': '本地凭据文件',
    'settings.apiKeySourceRequest': '当前输入',
    'settings.metaSyncOn': '云同步：已开启',
    'settings.metaSyncOff': '云同步：已关闭',
    'settings.testNotRun': '尚未执行连接测试。',
    'settings.testing': '正在测试连接...',
    'settings.testSuccess': '连接测试成功。',
    'settings.testFailed': '连接测试失败: {message}',
    'settings.testResultOk': '成功：{message}',
    'settings.testResultFail': '失败：{message}',
    'settings.saved': '设置已保存。',
    'settings.cleared': 'API Key 已清除。',
    'settings.clearConfirm': '确认清除已保存的 API Key 吗？',
    'settings.loadFailed': '加载设置失败: {message}',
    'settings.testFailedRequest': '连接测试请求失败: {message}',
    'settings.saveFailed': '保存设置失败: {message}',
    'settings.clearFailed': '清除 API Key 失败: {message}',
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
  diaryEditMode: false,
  diaryDraftContent: '',
  syncingDiaryId: null,
  currentDiaryDetail: null,
  currentPromptDetail: null,
  settings: null,
  settingsTest: null,
  apiBaseUrl: '',
  locale: 'en',
  fontSize: 'small',
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

function resolveFontSize(raw) {
  if (typeof raw !== 'string') {
    return 'small';
  }
  const trimmed = raw.trim().toLowerCase();
  if (SUPPORTED_FONT_SIZES.includes(trimmed)) {
    return trimmed;
  }
  return 'small';
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
  button.disabled = state.diaryEditMode;
  if (state.diaryViewMode === 'raw') {
    button.textContent = t('diary.viewReading');
    button.dataset.mode = 'raw';
  } else {
    button.textContent = t('diary.viewRaw');
    button.dataset.mode = 'reading';
  }
}

function applyDiaryEditButtons() {
  const editBtn = el('btnDiaryEdit');
  const saveBtn = el('btnDiarySave');
  const setDefaultBtn = el('btnDiarySetDefault');
  const syncBtn = el('btnDiarySync');
  if (!editBtn || !saveBtn || !setDefaultBtn || !syncBtn) {
    return;
  }
  editBtn.textContent = state.diaryEditMode ? t('actions.cancel') : t('actions.edit');
  const hasDiary = !!state.currentDiaryId;
  const isDefault = !!state.currentDiaryDetail?.isDefault;
  const hasDate = !!state.currentDiaryDetail?.date;
  const isSyncingCurrent = hasDiary && state.syncingDiaryId === state.currentDiaryId;
  const hasSyncInFlight = !!state.syncingDiaryId;
  editBtn.disabled = !hasDiary;
  saveBtn.disabled = !state.diaryEditMode || !hasDiary;
  setDefaultBtn.disabled = !hasDiary || !hasDate || state.diaryEditMode || isDefault;
  syncBtn.textContent = isSyncingCurrent ? t('actions.syncing') : t('actions.sync');
  syncBtn.classList.toggle('is-loading', isSyncingCurrent);
  syncBtn.setAttribute('aria-busy', isSyncingCurrent ? 'true' : 'false');
  syncBtn.disabled = !hasDiary || state.diaryEditMode || !isDefault || hasSyncInFlight;
}

function renderDiaryContent() {
  const preview = el('diaryContent');
  const editor = el('diaryEditor');
  if (!preview || !editor) {
    return;
  }

  const content = state.diaryRawContent || '';
  if (state.diaryEditMode) {
    preview.hidden = true;
    editor.hidden = false;
    editor.value = state.diaryDraftContent;
    editor.focus();
    applyDiaryViewModeButton();
    applyDiaryEditButtons();
    return;
  }

  preview.hidden = false;
  editor.hidden = true;
  if (state.diaryViewMode === 'markdown') {
    preview.classList.add('markdown-view');
    preview.innerHTML = markdownToHtml(content);
  } else {
    preview.classList.remove('markdown-view');
    preview.textContent = content;
  }
  applyDiaryViewModeButton();
  applyDiaryEditButtons();
}

function renderDiaryMeta() {
  const metaPrimary = el('diaryMetaPrimary');
  const metaFilename = el('diaryMetaFilename');
  if (!metaPrimary || !metaFilename) {
    return;
  }

  if (!state.currentDiaryDetail) {
    metaPrimary.textContent = '';
    metaFilename.textContent = '';
    return;
  }
  const detail = state.currentDiaryDetail;
  metaPrimary.textContent = t('diary.metaFormat', {
    date: detail.date || t('common.na'),
    modifiedAt: detail.modifiedAt || '',
  });
  metaFilename.textContent = detail.filename || '';
}

function setDiaryEmptyState() {
  state.currentDiaryDetail = null;
  state.diaryEditMode = false;
  state.diaryDraftContent = '';
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
  el('cliVersion').textContent = data.version || '-';
  state.activePromptId = data.activePrompt || '';
  state.apiBaseUrl = data.apiBaseUrl || '';
  if (!el('genOutput').value) {
    el('genOutput').value = data.defaultOutput || '';
  }
}

function apiKeySourceLabel(source) {
  if (source === 'env') {
    return t('settings.apiKeySourceEnv');
  }
  if (source === 'credentials') {
    return t('settings.apiKeySourceCredentials');
  }
  if (source === 'request') {
    return t('settings.apiKeySourceRequest');
  }
  return source || '';
}

function setNodeText(node, text) {
  if (!node) {
    return;
  }
  const value = String(text || '');
  node.textContent = value;
  node.hidden = value.trim() === '';
}

function renderSettings() {
  const cloudSwitch = el('settingCloudSync');
  const apiKeyStatus = el('settingsApiKeyStatus');
  const meta = el('settingsMeta');
  const onboarding = el('ownerOnboarding');
  const onboardingTitle = el('ownerOnboardingTitle');
  const onboardingHint = el('ownerOnboardingHint');
  const onboardingExtra = el('ownerOnboardingExtra');
  const onboardingRepo = el('ownerOnboardingRepo');
  const onboardingApiKey = el('ownerOnboardingApiKey');
  const onboardingBaseUrl = el('ownerOnboardingBaseUrl');
  if (!cloudSwitch || !apiKeyStatus || !meta) {
    return;
  }

  const rightApiKeyValue = state.settings?.apiKeyConfigured
    ? (state.settings.apiKeyMasked || '')
    : t('settings.statusNotConfigured');
  const rightBaseURLValue = state.apiBaseUrl || t('settings.statusNotConfigured');
  if (onboardingApiKey) {
    onboardingApiKey.textContent = t('settings.statusApiKey', { value: rightApiKeyValue });
  }
  if (onboardingBaseUrl) {
    onboardingBaseUrl.textContent = t('settings.statusBaseUrl', { value: rightBaseURLValue });
  }

  if (!state.settings) {
    cloudSwitch.checked = false;
    apiKeyStatus.textContent = t('settings.apiKeyNotConfigured');
    meta.textContent = t('settings.metaSyncOff');
    if (onboarding) {
      onboarding.hidden = false;
    }
    if (onboardingTitle) {
      onboardingTitle.textContent = t('settings.ownerTitle');
    }
    setNodeText(onboardingHint, t('settings.ownerHint'));
    if (onboardingExtra) {
      onboardingExtra.textContent = t('settings.ownerSteps');
    }
    if (onboardingRepo) {
      onboardingRepo.hidden = true;
    }
    renderSettingsTest();
    return;
  }

  cloudSwitch.checked = !!state.settings.cloudSyncEnabled;
  meta.textContent = state.settings.cloudSyncEnabled ? t('settings.metaSyncOn') : t('settings.metaSyncOff');
  if (onboarding) {
    onboarding.hidden = false;
  }

  if (state.settings.apiKeyConfigured) {
    const masked = state.settings.apiKeyMasked || '';
    const sourceLabel = apiKeySourceLabel(state.settings.apiKeySource);
    if (sourceLabel) {
      apiKeyStatus.textContent = t('settings.apiKeyConfiguredWithSource', { source: sourceLabel, masked });
    } else {
      apiKeyStatus.textContent = t('settings.apiKeyConfigured', { masked });
    }
    if (onboardingTitle) {
      onboardingTitle.textContent = t('settings.ownerConfiguredTitle');
    }
    setNodeText(onboardingHint, t('settings.ownerConfiguredHint'));
    if (onboardingExtra) {
      onboardingExtra.textContent = t('settings.ownerConfiguredExtra');
    }
    if (onboardingRepo) {
      onboardingRepo.hidden = false;
    }
    renderSettingsTest();
    return;
  }

  if (onboardingTitle) {
    onboardingTitle.textContent = t('settings.ownerTitle');
  }
  setNodeText(onboardingHint, t('settings.ownerHint'));
  if (onboardingExtra) {
    onboardingExtra.textContent = t('settings.ownerSteps');
  }
  if (onboardingRepo) {
    onboardingRepo.hidden = true;
  }
  apiKeyStatus.textContent = t('settings.apiKeyNotConfigured');
  renderSettingsTest();
}

async function loadSettings() {
  state.settings = await api('/settings');
  renderSettings();
}

function renderSettingsTest() {
  const target = el('settingsTestResult');
  if (!target) {
    return;
  }
  if (!state.settingsTest) {
    target.textContent = t('settings.testNotRun');
    target.style.color = '';
    return;
  }

  const message = state.settingsTest.message || '';
  if (state.settingsTest.success) {
    target.textContent = t('settings.testResultOk', { message });
    target.style.color = 'var(--lime)';
  } else {
    target.textContent = t('settings.testResultFail', { message });
    target.style.color = 'var(--coral)';
  }
}

function diaryCalendarParts(dateRaw) {
  const match = String(dateRaw || '').match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) {
    return { hasDate: false, yearMonth: '--', day: '--' };
  }
  return {
    hasDate: true,
    yearMonth: `${match[1]}-${match[2]}`,
    day: match[3],
  };
}

function renderDiaryCalendar(dateRaw) {
  const parts = diaryCalendarParts(dateRaw);
  const emptyClass = parts.hasDate ? '' : ' empty';
  return `
    <div class="calendar-chip${emptyClass}" aria-hidden="true">
      <span class="calendar-ym">${escapeHtml(parts.yearMonth)}</span>
      <strong class="calendar-day">${escapeHtml(parts.day)}</strong>
    </div>
  `;
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
      const calendar = renderDiaryCalendar(item.date || '');
      const defaultTag = item.isDefault ? `<span class="default-tag">${escapeHtml(t('diary.defaultTag'))}</span>` : '';
      const showSync = !!item.isDefault;
      const isSyncing = state.syncingDiaryId === item.id;
      const hasSyncInFlight = !!state.syncingDiaryId;
      const syncLabel = isSyncing ? t('actions.syncing') : t('actions.sync');
      const syncDisabled = hasSyncInFlight ? 'disabled' : '';
      const syncClass = isSyncing ? ' is-loading' : '';
      const syncButton = showSync
        ? `<button type="button" class="mini-btn icon-btn diary-sync-btn${syncClass}" data-id="${escapeHtml(item.id)}" data-action="sync" title="${escapeHtml(syncLabel)}" aria-label="${escapeHtml(syncLabel)}" ${syncDisabled}><span class="sync-icon" aria-hidden="true">&#x21bb;</span></button>`
        : '';
      return `
        <article class="item diary-item ${active}" data-id="${escapeHtml(item.id)}">
          <div class="diary-item-head">
            ${calendar}
            <div class="diary-item-main">
              <div class="diary-item-row">
                <h3>${escapeHtml(item.title || item.filename)}</h3>
                ${syncButton}
              </div>
              <div class="meta">${escapeHtml(item.date || t('common.na'))} · ${escapeHtml(item.filename)} ${defaultTag}</div>
              <p class="item-preview">${escapeHtml(item.preview || '')}</p>
            </div>
          </div>
        </article>
      `;
    })
    .join('');

  container.querySelectorAll('.item').forEach((node) => {
    node.addEventListener('click', () => loadDiaryDetail(node.dataset.id));
  });
  container.querySelectorAll('[data-action="sync"]').forEach((node) => {
    node.addEventListener('click', (event) => {
      event.stopPropagation();
      syncDiary(node.dataset.id, false).catch((err) => setStatusKey('diary.syncFailed', { message: err.message }, true));
    });
  });
}

async function loadDiaries() {
  const q = el('diarySearch').value.trim();
  const data = await api(`/diaries?limit=200&q=${encodeURIComponent(q)}`);
  state.diaries = data.items || [];
  if (state.currentDiaryId && !state.diaries.some((item) => item.id === state.currentDiaryId)) {
    state.currentDiaryId = null;
  }
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
    state.diaryEditMode = false;
    state.diaryRawContent = data.content || '';
    state.diaryDraftContent = state.diaryRawContent;
    if (rerender) {
      renderDiaryList(state.diaries);
    }
    renderDiaryMeta();
    renderDiaryContent();
  } catch (err) {
    setStatusKey('diary.loadFailed', { message: err.message }, true);
  }
}

function toggleDiaryEditMode() {
  if (!state.currentDiaryId) {
    setStatusKey('diary.saveMissing', {}, true);
    return;
  }
  state.diaryEditMode = !state.diaryEditMode;
  if (state.diaryEditMode) {
    state.diaryDraftContent = state.diaryRawContent || '';
  } else {
    state.diaryDraftContent = state.diaryRawContent || '';
  }
  renderDiaryContent();
}

async function saveDiaryContent() {
  if (!state.currentDiaryId) {
    setStatusKey('diary.saveMissing', {}, true);
    return;
  }
  const editor = el('diaryEditor');
  if (!editor) {
    return;
  }
  const content = editor.value;
  const data = await api(`/diaries/${encodeURIComponent(state.currentDiaryId)}`, {
    method: 'PATCH',
    body: JSON.stringify({ content }),
  });
  state.currentDiaryDetail = data;
  state.diaryRawContent = data.content || '';
  state.diaryDraftContent = state.diaryRawContent;
  state.diaryEditMode = false;
  await loadDiaries();
  setStatusKey('diary.saveSuccess');
}

async function setDiaryDefault() {
  if (!state.currentDiaryId) {
    setStatusKey('diary.saveMissing', {}, true);
    return;
  }
  const data = await api(`/diaries/${encodeURIComponent(state.currentDiaryId)}/set-default`, {
    method: 'POST',
  });
  state.currentDiaryDetail = data;
  await loadDiaries();
  setStatusKey('diary.setDefaultSuccess', { date: data.date || '' });
}

async function syncDiary(id = state.currentDiaryId, refreshDetail = true) {
  if (!id) {
    setStatusKey('diary.saveMissing', {}, true);
    return;
  }

  const current = state.currentDiaryId === id ? state.currentDiaryDetail : null;
  const listItem = state.diaries.find((item) => item.id === id);
  const diaryTitle = current?.title || current?.filename || listItem?.title || listItem?.filename || id;

  if (state.syncingDiaryId) {
    if (state.syncingDiaryId === id) {
      setStatusKey('diary.syncInProgress', { title: diaryTitle });
      return;
    }
    setStatusKey('diary.syncBusy', {}, true);
    return;
  }

  state.syncingDiaryId = id;
  setStatusKey('diary.syncInProgress', { title: diaryTitle });
  renderDiaryList(state.diaries);
  applyDiaryEditButtons();
  try {
    const data = await api(`/diaries/${encodeURIComponent(id)}/sync`, {
      method: 'POST',
    });
    setStatusKey('diary.syncSuccess', { action: data.action || 'SYNC', title: diaryTitle });
    await loadDiaries();
    if (refreshDetail && state.currentDiaryId) {
      await loadDiaryDetail(state.currentDiaryId, false);
    }
  } finally {
    state.syncingDiaryId = null;
    renderDiaryList(state.diaries);
    applyDiaryEditButtons();
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

async function saveSettings(event) {
  event.preventDefault();

  const payload = {
    cloudSyncEnabled: !!el('settingCloudSync').checked,
  };
  const apiKey = el('settingApiKey').value.trim();
  if (apiKey) {
    payload.apiKey = apiKey;
  }

  const data = await api('/settings', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
  state.settings = data;
  state.settingsTest = null;
  el('settingApiKey').value = '';
  renderSettings();
  setStatusKey('settings.saved');

  if (data.cloudSyncEnabled) {
    try {
      await testSettingsConnection();
    } catch (err) {
      setStatusKey('settings.testFailedRequest', { message: err.message }, true);
    }
  }
}

async function clearApiKey() {
  const yes = confirm(t('settings.clearConfirm'));
  if (!yes) {
    return;
  }

  const data = await api('/settings', {
    method: 'PATCH',
    body: JSON.stringify({ apiKey: '' }),
  });
  state.settings = data;
  state.settingsTest = null;
  el('settingApiKey').value = '';
  renderSettings();
  setStatusKey('settings.cleared');
}

async function testSettingsConnection() {
  const button = el('btnTestConnection');
  if (button) {
    button.disabled = true;
  }
  setStatusKey('settings.testing');

  try {
    const apiKey = el('settingApiKey').value.trim();
    const payload = apiKey ? { apiKey } : {};
    const data = await api('/settings/test-connection', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    state.settingsTest = data;
    renderSettingsTest();
    if (data.success) {
      setStatusKey('settings.testSuccess');
    } else {
      setStatusKey('settings.testFailed', { message: data.message || '' }, true);
    }
  } finally {
    if (button) {
      button.disabled = false;
    }
  }
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
  renderSettings();
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

function applyFontSize(size, persist = true) {
  const next = resolveFontSize(size);
  state.fontSize = next;
  document.documentElement.setAttribute('data-font-size', next);
  const picker = el('fontSizeSwitch');
  if (picker && picker.value !== next) {
    picker.value = next;
  }
  if (persist) {
    window.localStorage.setItem(FONT_SIZE_STORAGE_KEY, next);
  }
}

function initFontSize() {
  const saved = window.localStorage.getItem(FONT_SIZE_STORAGE_KEY);
  const fontSize = resolveFontSize(saved || 'small');
  const picker = el('fontSizeSwitch');
  if (picker) {
    picker.value = fontSize;
  }
  applyFontSize(fontSize, false);
}

function bindEvents() {
  document.querySelectorAll('.tab-btn').forEach((btn) => {
    btn.addEventListener('click', () => switchTab(btn.dataset.tab));
  });

  el('langSwitch').addEventListener('change', (event) => {
    setLocale(event.target.value, true);
  });

  el('fontSizeSwitch').addEventListener('change', (event) => {
    applyFontSize(event.target.value, true);
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

  el('settingsForm').addEventListener('submit', (event) => {
    saveSettings(event).catch((err) => setStatusKey('settings.saveFailed', { message: err.message }, true));
  });

  el('btnTestConnection').addEventListener('click', () => {
    testSettingsConnection().catch((err) => setStatusKey('settings.testFailedRequest', { message: err.message }, true));
  });

  el('btnClearApiKey').addEventListener('click', () => {
    clearApiKey().catch((err) => setStatusKey('settings.clearFailed', { message: err.message }, true));
  });

  el('btnDiaryViewMode').addEventListener('click', () => {
    state.diaryViewMode = state.diaryViewMode === 'raw' ? 'markdown' : 'raw';
    renderDiaryContent();
  });

  el('btnDiaryEdit').addEventListener('click', () => {
    toggleDiaryEditMode();
  });

  el('btnDiarySave').addEventListener('click', () => {
    saveDiaryContent().catch((err) => setStatusKey('diary.saveFailed', { message: err.message }, true));
  });

  el('btnDiarySetDefault').addEventListener('click', () => {
    setDiaryDefault().catch((err) => setStatusKey('diary.setDefaultFailed', { message: err.message }, true));
  });

  el('btnDiarySync').addEventListener('click', () => {
    syncDiary().catch((err) => setStatusKey('diary.syncFailed', { message: err.message }, true));
  });

  el('diaryEditor').addEventListener('input', (event) => {
    state.diaryDraftContent = event.target.value;
  });
}

async function bootstrap() {
  setStatusKey('status.loading');
  await loadState();
  await loadDiaries();
  await loadPrompts();
  await loadSettings();
  setStatusKey('status.ready');
}

function initDateDefault() {
  const today = new Date().toISOString().slice(0, 10);
  el('genDate').value = today;
}

document.addEventListener('DOMContentLoaded', async () => {
  initFontSize();
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
