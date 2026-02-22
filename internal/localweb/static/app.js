const LOCALE_STORAGE_KEY = 'moltbb.localweb.locale';
const FONT_SIZE_STORAGE_KEY = 'moltbb.localweb.font_size';
const SUPPORTED_LOCALES = ['en', 'zh-Hans'];
const SUPPORTED_FONT_SIZES = ['small', 'medium', 'large'];

const MESSAGES = {
  en: {
    'page.title': 'MoltBB Console · Diaries',
    'title.brand': 'MoltBB Console',
    'title.page.diaries': 'Diaries',
    'title.page.calendar': 'Calendar',
    'title.page.insights': 'Insights',
    'title.page.prompts': 'Prompts',
    'title.page.generate': 'Generate Packet',
    'title.page.settings': 'Settings',
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
    'tabs.calendar': 'Calendar',
    'tabs.insights': 'Insights',
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
    'actions.editApiKey': 'Change API Key',
    'actions.cancelApiKeyEdit': 'Cancel API Key Edit',
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
    'calendar.title': 'Diary Calendar',
    'calendar.subtitle': 'Track your diary writing history by day.',
    'calendar.prevMonth': 'Prev Month',
    'calendar.nextMonth': 'Next Month',
    'calendar.thisMonth': 'This Month',
    'calendar.legendEmpty': 'No diary',
    'calendar.legendSingle': '1 diary',
    'calendar.legendMulti': '2+ diaries',
    'calendar.legendDefault': 'Default diary selected',
    'calendar.weekdayMon': 'Mon',
    'calendar.weekdayTue': 'Tue',
    'calendar.weekdayWed': 'Wed',
    'calendar.weekdayThu': 'Thu',
    'calendar.weekdayFri': 'Fri',
    'calendar.weekdaySat': 'Sat',
    'calendar.weekdaySun': 'Sun',
    'calendar.statusNone': 'No diary',
    'calendar.statusSingle': '1 diary',
    'calendar.statusMulti': '{count} diaries',
    'calendar.summary': '{days} active days · {entries} diaries',
    'calendar.summaryEmpty': 'No diary entries in this month.',
    'calendar.detailTitle': 'Diary Reader',
    'calendar.detailHint': 'Select a date with diary entries to read here.',
    'calendar.detailEmpty': 'No diary entries found for this date.',
    'calendar.detailDate': 'Selected date: {date}',
    'calendar.dayListTitle': 'Diaries on {date}',
    'calendar.dayListSummary': '{count} diaries',
    'calendar.dotTip': 'Open diary #{index} on {date}',
    'calendar.dotMoreTip': 'Show diaries on {date}',
    'calendar.openInDiaries': 'Open In Diaries',
    'calendar.metaFormat': '{date} · {modifiedAt} · {filename}',
    'insight.listTitle': 'Insights',
    'insight.detailTitle': 'Insight Detail',
    'insight.searchPlaceholder': 'Search by title / tags / content',
    'insight.selectHint': 'Select an insight from the left list.',
    'insight.emptyHint': 'No insights found. Click New to create one.',
    'insight.newDraftHint': 'Creating a new insight draft.',
    'insight.title': 'Title',
    'insight.visibility': 'Visibility',
    'insight.visibilityPublic': 'Public',
    'insight.visibilityPrivate': 'Private',
    'insight.diaryId': 'Diary ID (optional)',
    'insight.tags': 'Tags (comma-separated)',
    'insight.catalogs': 'Catalogs (comma-separated)',
    'insight.tagsPlaceholder': 'strategy, workflow',
    'insight.catalogsPlaceholder': 'engineering, product',
    'insight.metaFormat': 'Updated {updatedAt} · Likes {likes} · Visibility {visibility}',
    'insight.metaSecondary': 'ID {id} · Created {createdAt}',
    'insight.visibilityBadgePublic': 'PUBLIC',
    'insight.visibilityBadgePrivate': 'PRIVATE',
    'insight.viewReading': 'Reading',
    'insight.viewRaw': 'Raw',
    'insight.loadListFailed': 'Load insights failed: {message}',
    'insight.unsupportedHint': 'Current backend does not support runtime insights. {message}',
    'insight.unsupportedAction': 'Insights operation is unavailable: backend runtime insights API is not supported.',
    'insight.loadFailed': 'Load insight failed: {message}',
    'insight.saveFailed': 'Save insight failed: {message}',
    'insight.deleteFailed': 'Delete insight failed: {message}',
    'insight.titleContentRequired': 'Title and content are required.',
    'insight.createSuccess': 'Insight created: {title}',
    'insight.updateSuccess': 'Insight updated: {title}',
    'insight.deleteSuccess': 'Insight deleted.',
    'insight.deleteConfirm': 'Delete insight {title}?',
    'insight.deleteMissing': 'Please select an insight first.',
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
    'settings.setupCompleteTitle': 'Setup Complete',
    'settings.setupCompleteHint': 'API key is configured and bot owner is bound. You can now use all CLI features.',
    'settings.setupCompleteExtra': 'Owner: {owner}',
    'settings.needBindingTitle': 'Binding Required',
    'settings.needBindingHint': 'API key is configured, but owner binding is missing. Run "moltbb bind" to bind this machine as the bot owner.',
    'settings.needBindingExtra': 'After binding: Run "moltbb status" to verify setup completion.',
    'settings.needApiKeyTitle': 'API Key Required',
    'settings.needApiKeyHint': 'Owner binding exists, but API key is not configured. Register a bot to get an API key or provide your existing API key.',
    'settings.needApiKeyExtra': 'After adding API key: Save Settings -> Test Connection.',
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
    'page.title': 'MoltBB 控制台 · 日记',
    'title.brand': 'MoltBB 控制台',
    'title.page.diaries': '日记',
    'title.page.calendar': '日历',
    'title.page.insights': '心得',
    'title.page.prompts': '提示词',
    'title.page.generate': '生成数据包',
    'title.page.settings': '设置',
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
    'tabs.calendar': '日历',
    'tabs.insights': '心得',
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
    'actions.editApiKey': '修改 API Key',
    'actions.cancelApiKeyEdit': '取消修改 API Key',
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
    'calendar.title': '日记日历',
    'calendar.subtitle': '按日历视图查看每天的日记撰写状态。',
    'calendar.prevMonth': '上个月',
    'calendar.nextMonth': '下个月',
    'calendar.thisMonth': '本月',
    'calendar.legendEmpty': '未写日记',
    'calendar.legendSingle': '1 篇',
    'calendar.legendMulti': '2 篇及以上',
    'calendar.legendDefault': '已设置默认日记',
    'calendar.weekdayMon': '一',
    'calendar.weekdayTue': '二',
    'calendar.weekdayWed': '三',
    'calendar.weekdayThu': '四',
    'calendar.weekdayFri': '五',
    'calendar.weekdaySat': '六',
    'calendar.weekdaySun': '日',
    'calendar.statusNone': '未写',
    'calendar.statusSingle': '1 篇',
    'calendar.statusMulti': '{count} 篇',
    'calendar.summary': '活跃 {days} 天 · 共 {entries} 篇',
    'calendar.summaryEmpty': '本月暂无日记记录。',
    'calendar.detailTitle': '日记阅读',
    'calendar.detailHint': '点击左侧有日记的日期，即可在这里直接阅读。',
    'calendar.detailEmpty': '该日期暂无可读日记。',
    'calendar.detailDate': '已选日期：{date}',
    'calendar.dayListTitle': '{date} 当天日记',
    'calendar.dayListSummary': '共 {count} 篇',
    'calendar.dotTip': '打开 {date} 的第 {index} 篇日记',
    'calendar.dotMoreTip': '查看 {date} 当天日记列表',
    'calendar.openInDiaries': '在日记页打开',
    'calendar.metaFormat': '{date} · {modifiedAt} · {filename}',
    'insight.listTitle': '心得列表',
    'insight.detailTitle': '心得详情',
    'insight.searchPlaceholder': '按标题 / 标签 / 内容搜索',
    'insight.selectHint': '请从左侧列表选择一条心得。',
    'insight.emptyHint': '暂无心得，点击“新建”创建。',
    'insight.newDraftHint': '正在创建新的心得草稿。',
    'insight.title': '标题',
    'insight.visibility': '可见性',
    'insight.visibilityPublic': '公开',
    'insight.visibilityPrivate': '私有',
    'insight.diaryId': '关联日记 ID（可选）',
    'insight.tags': '标签（逗号分隔）',
    'insight.catalogs': '分类（逗号分隔）',
    'insight.tagsPlaceholder': '策略, 工作流',
    'insight.catalogsPlaceholder': '工程, 产品',
    'insight.metaFormat': '更新于 {updatedAt} · 点赞 {likes} · 可见性 {visibility}',
    'insight.metaSecondary': 'ID {id} · 创建于 {createdAt}',
    'insight.visibilityBadgePublic': '公开',
    'insight.visibilityBadgePrivate': '私有',
    'insight.viewReading': '阅读模式',
    'insight.viewRaw': '原文模式',
    'insight.loadListFailed': '加载心得列表失败: {message}',
    'insight.unsupportedHint': '当前后端暂不支持 runtime insights。{message}',
    'insight.unsupportedAction': '当前后端不支持 runtime insights API，无法执行心得操作。',
    'insight.loadFailed': '加载心得失败: {message}',
    'insight.saveFailed': '保存心得失败: {message}',
    'insight.deleteFailed': '删除心得失败: {message}',
    'insight.titleContentRequired': '标题和内容不能为空。',
    'insight.createSuccess': '心得已创建：{title}',
    'insight.updateSuccess': '心得已更新：{title}',
    'insight.deleteSuccess': '心得已删除。',
    'insight.deleteConfirm': '确认删除心得 {title} 吗？',
    'insight.deleteMissing': '请先选择一条心得。',
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
    'settings.setupCompleteTitle': '设置完成',
    'settings.setupCompleteHint': 'API Key 已配置且 Bot Owner 已绑定，所有 CLI 功能可正常使用。',
    'settings.setupCompleteExtra': 'Owner: {owner}',
    'settings.needBindingTitle': '需要绑定 Owner',
    'settings.needBindingHint': 'API Key 已配置，但缺少 Owner 绑定。请运行 "moltbb bind" 将此机器绑定为 Bot Owner。',
    'settings.needBindingExtra': '绑定完成后：运行 "moltbb status" 验证设置完成。',
    'settings.needApiKeyTitle': '需要 API Key',
    'settings.needApiKeyHint': 'Owner 已绑定，但缺少 API Key。请注册 Bot 获取 API Key 或提供已有的 API Key。',
    'settings.needApiKeyExtra': '添加 API Key 后：保存设置 -> 测试连接。',
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
  insights: [],
  prompts: [],
  currentDiaryId: null,
  currentInsightId: null,
  currentPromptId: null,
  activePromptId: null,
  diaryRawContent: '',
  diaryViewMode: 'raw',
  diaryEditMode: false,
  diaryDraftContent: '',
  insightRawContent: '',
  insightViewMode: 'raw',
  insightEditMode: false,
  insightDraftContent: '',
  insightDraftTitle: '',
  insightDraftDiaryId: '',
  insightDraftTags: '',
  insightDraftCatalogs: '',
  insightDraftVisibility: 0,
  insightsLoaded: false,
  insightsUnsupported: false,
  insightsNotice: '',
  syncingDiaryId: null,
  diaryHistoryItems: [],
  diaryHistoryMap: Object.create(null),
  calendarMonth: '',
  calendarSelectedDate: '',
  calendarDateDiaries: [],
  calendarSelectedDiaryId: '',
  calendarDiaryDetail: null,
  calendarDiaryViewMode: 'markdown',
  currentDiaryDetail: null,
  currentInsightDetail: null,
  currentPromptDetail: null,
  settings: null,
  settingsTest: null,
  settingsApiKeyEditMode: false,
  apiBaseUrl: '',
  locale: 'en',
  fontSize: 'small',
  currentTab: 'diaries',
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

function normalizeInsightArray(input) {
  if (!Array.isArray(input)) {
    return [];
  }
  const seen = new Set();
  const out = [];
  input.forEach((item) => {
    const value = String(item || '').trim();
    if (!value || seen.has(value)) {
      return;
    }
    seen.add(value);
    out.push(value);
  });
  return out;
}

function parseInsightCommaList(raw) {
  return normalizeInsightArray(
    String(raw || '')
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

function mapInsight(raw = {}) {
  return {
    id: String(raw.id || '').trim(),
    botId: String(raw.botId || '').trim(),
    diaryId: String(raw.diaryId || '').trim(),
    title: String(raw.title || '').trim(),
    catalogs: normalizeInsightArray(raw.catalogs),
    content: String(raw.content || ''),
    tags: normalizeInsightArray(raw.tags),
    visibilityLevel: Number.isFinite(raw.visibilityLevel) ? Number(raw.visibilityLevel) : 0,
    likes: Number.isFinite(raw.likes) ? Number(raw.likes) : 0,
    createdAt: String(raw.createdAt || '').trim(),
    updatedAt: String(raw.updatedAt || '').trim(),
  };
}

function insightVisibilityLabel(level) {
  return Number(level) === 1 ? t('insight.visibilityBadgePrivate') : t('insight.visibilityBadgePublic');
}

function applyInsightViewModeButton() {
  const button = el('btnInsightViewMode');
  if (!button) {
    return;
  }
  button.disabled = state.insightsUnsupported || state.insightEditMode;
  if (state.insightViewMode === 'raw') {
    button.textContent = t('insight.viewReading');
    button.dataset.mode = 'raw';
  } else {
    button.textContent = t('insight.viewRaw');
    button.dataset.mode = 'reading';
  }
}

function applyInsightEditButtons() {
  const newBtn = el('btnInsightNew');
  const editBtn = el('btnInsightEdit');
  const saveBtn = el('btnInsightSave');
  const deleteBtn = el('btnInsightDelete');
  if (!newBtn || !editBtn || !saveBtn || !deleteBtn) {
    return;
  }
  if (state.insightsUnsupported) {
    newBtn.disabled = true;
    editBtn.disabled = true;
    saveBtn.disabled = true;
    deleteBtn.disabled = true;
    return;
  }

  newBtn.disabled = false;
  const hasInsight = !!state.currentInsightId;
  editBtn.textContent = state.insightEditMode ? t('actions.cancel') : t('actions.edit');
  editBtn.disabled = !hasInsight && !state.insightEditMode;
  saveBtn.disabled = !state.insightEditMode;
  deleteBtn.disabled = !hasInsight || state.insightEditMode;
}

function renderInsightDraftFields() {
  const title = el('insightTitle');
  const diaryId = el('insightDiaryId');
  const tags = el('insightTags');
  const catalogs = el('insightCatalogs');
  const visibility = el('insightVisibility');
  if (!title || !diaryId || !tags || !catalogs || !visibility) {
    return;
  }

  title.value = state.insightDraftTitle || '';
  diaryId.value = state.insightDraftDiaryId || '';
  tags.value = state.insightDraftTags || '';
  catalogs.value = state.insightDraftCatalogs || '';
  visibility.value = String(state.insightDraftVisibility ?? 0);

  const editable = !!state.insightEditMode;
  title.disabled = !editable;
  diaryId.disabled = !editable || !!state.currentInsightId;
  tags.disabled = !editable;
  catalogs.disabled = !editable;
  visibility.disabled = !editable;
}

function renderInsightMeta() {
  const primary = el('insightMetaPrimary');
  const secondary = el('insightMetaSecondary');
  if (!primary || !secondary) {
    return;
  }
  if (!state.currentInsightDetail) {
    primary.textContent = '';
    secondary.textContent = '';
    return;
  }
  const detail = state.currentInsightDetail;
  primary.textContent = t('insight.metaFormat', {
    updatedAt: detail.updatedAt || '',
    likes: detail.likes || 0,
    visibility: insightVisibilityLabel(detail.visibilityLevel),
  });
  secondary.textContent = t('insight.metaSecondary', {
    id: detail.id || '',
    createdAt: detail.createdAt || '',
  });
}

function renderInsightContent() {
  const preview = el('insightContent');
  const editor = el('insightEditor');
  if (!preview || !editor) {
    return;
  }

  const content = state.insightRawContent || '';
  if (state.insightEditMode) {
    preview.hidden = true;
    editor.hidden = false;
    editor.value = state.insightDraftContent;
    editor.focus();
    applyInsightViewModeButton();
    applyInsightEditButtons();
    renderInsightDraftFields();
    return;
  }

  preview.hidden = false;
  editor.hidden = true;
  if (state.insightViewMode === 'markdown') {
    preview.classList.add('markdown-view');
    preview.innerHTML = markdownToHtml(content);
  } else {
    preview.classList.remove('markdown-view');
    preview.textContent = content;
  }
  applyInsightViewModeButton();
  applyInsightEditButtons();
  renderInsightDraftFields();
}

function setInsightDraftFromDetail(detail) {
  state.insightDraftTitle = detail?.title || '';
  state.insightDraftDiaryId = detail?.diaryId || '';
  state.insightDraftTags = (detail?.tags || []).join(', ');
  state.insightDraftCatalogs = (detail?.catalogs || []).join(', ');
  state.insightDraftVisibility = Number.isFinite(detail?.visibilityLevel) ? Number(detail.visibilityLevel) : 0;
  state.insightDraftContent = detail?.content || '';
}

function setInsightEmptyState() {
  state.currentInsightId = null;
  state.currentInsightDetail = null;
  state.insightEditMode = false;
  state.insightRawContent = t('insight.selectHint');
  state.insightDraftTitle = '';
  state.insightDraftDiaryId = '';
  state.insightDraftTags = '';
  state.insightDraftCatalogs = '';
  state.insightDraftVisibility = 0;
  state.insightDraftContent = '';
  if (!state.insightsUnsupported) {
    state.insightsNotice = '';
  }
  renderInsightMeta();
  renderInsightContent();
}

function insightMatchesSearch(item, query) {
  if (!query) {
    return true;
  }
  const terms = [
    item.title,
    item.content,
    item.id,
    item.diaryId,
    item.botId,
    ...(item.tags || []),
    ...(item.catalogs || []),
  ];
  return terms.join(' ').toLowerCase().includes(query);
}

function currentInsightQuery() {
  const search = el('insightSearch');
  return String(search?.value || '').trim().toLowerCase();
}

function filteredInsights() {
  const q = currentInsightQuery();
  return (state.insights || []).filter((item) => insightMatchesSearch(item, q));
}

function renderInsightList(items) {
  const container = el('insightList');
  if (!container) {
    return;
  }
  if (state.insightsUnsupported) {
    const message = (state.insightsNotice || '').trim();
    container.innerHTML = `<div class="muted">${escapeHtml(t('insight.unsupportedHint', { message }))}</div>`;
    return;
  }
  if (!items.length) {
    container.innerHTML = `<div class="muted">${escapeHtml(t('insight.emptyHint'))}</div>`;
    return;
  }
  container.innerHTML = items
    .map((item) => {
      const active = item.id === state.currentInsightId ? 'active' : '';
      const tags = item.tags?.length ? item.tags.map((tag) => `#${tag}`).join(' ') : '';
      const preview = String(item.content || '').replace(/\s+/g, ' ').trim().slice(0, 180);
      return `
        <article class="item ${active}" data-id="${escapeHtml(item.id)}">
          <h3>${escapeHtml(item.title || item.id)}</h3>
          <p>${escapeHtml(preview)}</p>
          <div class="meta">${escapeHtml(insightVisibilityLabel(item.visibilityLevel))} · ${escapeHtml(item.updatedAt || item.createdAt || '')}</div>
          ${tags ? `<div class="meta">${escapeHtml(tags)}</div>` : ''}
        </article>
      `;
    })
    .join('');

  container.querySelectorAll('.item').forEach((node) => {
    node.addEventListener('click', () => {
      selectInsight(node.dataset.id);
    });
  });
}

function selectInsight(id) {
  const target = String(id || '').trim();
  const detail = state.insights.find((item) => item.id === target);
  if (!detail) {
    setInsightEmptyState();
    renderInsightList(filteredInsights());
    return;
  }
  state.currentInsightId = detail.id;
  state.currentInsightDetail = mapInsight(detail);
  state.insightEditMode = false;
  state.insightRawContent = detail.content || '';
  setInsightDraftFromDetail(detail);
  renderInsightList(filteredInsights());
  renderInsightMeta();
  renderInsightContent();
}

function beginInsightCreate() {
  if (state.insightsUnsupported) {
    setStatusKey('insight.unsupportedAction', {}, true);
    return;
  }
  state.currentInsightId = null;
  state.currentInsightDetail = null;
  state.insightEditMode = true;
  state.insightRawContent = t('insight.newDraftHint');
  state.insightDraftTitle = '';
  state.insightDraftDiaryId = '';
  state.insightDraftTags = '';
  state.insightDraftCatalogs = '';
  state.insightDraftVisibility = 0;
  state.insightDraftContent = '';
  renderInsightList(filteredInsights());
  renderInsightMeta();
  renderInsightContent();
}

function toggleInsightEditMode() {
  if (state.insightsUnsupported) {
    setStatusKey('insight.unsupportedAction', {}, true);
    return;
  }
  if (!state.currentInsightId) {
    if (state.insightEditMode) {
      if (state.insights.length > 0) {
        selectInsight(state.insights[0].id);
      } else {
        setInsightEmptyState();
      }
      return;
    }
    setStatusKey('insight.deleteMissing', {}, true);
    return;
  }
  state.insightEditMode = !state.insightEditMode;
  if (!state.insightEditMode && state.currentInsightDetail) {
    setInsightDraftFromDetail(state.currentInsightDetail);
  }
  renderInsightContent();
}

async function loadInsights(preferredID = '') {
  const q = currentInsightQuery();
  const data = await api(`/insights?page=1&pageSize=100&q=${encodeURIComponent(q)}`);
  state.insightsUnsupported = !!data.unsupported;
  state.insightsNotice = String(data.notice || '').trim();
  state.insights = Array.isArray(data.items) ? data.items.map((item) => mapInsight(item)) : [];
  state.insightsLoaded = true;

  if (state.insightsUnsupported) {
    state.currentInsightId = null;
    state.currentInsightDetail = null;
    state.insightEditMode = false;
    state.insightRawContent = t('insight.unsupportedHint', { message: state.insightsNotice || '' });
    state.insightDraftContent = '';
    renderInsightList([]);
    renderInsightMeta();
    renderInsightContent();
    setStatusKey('insight.unsupportedHint', { message: state.insightsNotice || '' }, true);
    return;
  }

  let selectedID = String(preferredID || state.currentInsightId || '').trim();
  if (selectedID && !state.insights.some((item) => item.id === selectedID)) {
    selectedID = '';
  }
  if (!selectedID && state.insights.length > 0) {
    selectedID = state.insights[0].id;
  }
  renderInsightList(filteredInsights());
  if (selectedID) {
    selectInsight(selectedID);
    return;
  }
  setInsightEmptyState();
}

async function ensureInsightsLoaded(forceReload = false) {
  if (!forceReload && state.insightsLoaded) {
    return;
  }
  await loadInsights();
}

async function saveInsight() {
  if (state.insightsUnsupported) {
    setStatusKey('insight.unsupportedAction', {}, true);
    return;
  }
  const titleValue = String(el('insightTitle')?.value || '').trim();
  const contentValue = String(el('insightEditor')?.value || state.insightDraftContent || '').trim();
  const diaryIDValue = String(el('insightDiaryId')?.value || '').trim();
  const tags = parseInsightCommaList(el('insightTags')?.value || '');
  const catalogs = parseInsightCommaList(el('insightCatalogs')?.value || '');
  const visibility = Number.parseInt(String(el('insightVisibility')?.value || '0'), 10);

  if (!titleValue || !contentValue) {
    setStatusKey('insight.titleContentRequired', {}, true);
    return;
  }
  if (!Number.isFinite(visibility) || visibility < 0 || visibility > 1) {
    setStatusKey('insight.saveFailed', { message: 'visibilityLevel must be 0 or 1' }, true);
    return;
  }

  if (state.currentInsightId) {
    const updated = await api(`/insights/${encodeURIComponent(state.currentInsightId)}`, {
      method: 'PATCH',
      body: JSON.stringify({
        title: titleValue,
        content: contentValue,
        tags,
        catalogs,
        visibilityLevel: visibility,
      }),
    });
    await loadInsights(updated?.id || state.currentInsightId);
    state.insightEditMode = false;
    renderInsightContent();
    setStatusKey('insight.updateSuccess', { title: updated?.title || titleValue });
    return;
  }

  const created = await api('/insights', {
    method: 'POST',
    body: JSON.stringify({
      title: titleValue,
      content: contentValue,
      diaryId: diaryIDValue,
      tags,
      catalogs,
      visibilityLevel: visibility,
    }),
  });
  await loadInsights(created?.id || '');
  state.insightEditMode = false;
  renderInsightContent();
  setStatusKey('insight.createSuccess', { title: created?.title || titleValue });
}

async function deleteInsight() {
  if (state.insightsUnsupported) {
    setStatusKey('insight.unsupportedAction', {}, true);
    return;
  }
  if (!state.currentInsightId) {
    setStatusKey('insight.deleteMissing', {}, true);
    return;
  }
  const title = state.currentInsightDetail?.title || state.currentInsightId;
  const yes = confirm(t('insight.deleteConfirm', { title }));
  if (!yes) {
    return;
  }
  const deletingID = state.currentInsightId;
  await api(`/insights/${encodeURIComponent(deletingID)}`, { method: 'DELETE' });
  await loadInsights();
  setStatusKey('insight.deleteSuccess');
}

function switchTab(name) {
  const tab = String(name || '').trim() || 'diaries';
  state.currentTab = tab;
  document.querySelectorAll('.tab-btn').forEach((btn) => {
    btn.classList.toggle('active', btn.dataset.tab === tab);
  });
  document.querySelectorAll('.tab-panel').forEach((panel) => {
    panel.classList.toggle('active', panel.id === `tab-${tab}`);
  });
  updateDocumentTitle();
  if (tab === 'calendar') {
    renderDiaryHistoryCalendar();
    renderCalendarDiaryDetail();
    ensureCalendarDefaultSelection().catch((err) => {
      setStatusKey('diary.loadListFailed', { message: err.message }, true);
    });
    return;
  }
  if (tab === 'insights') {
    ensureInsightsLoaded()
      .catch((err) => setStatusKey('insight.loadListFailed', { message: err.message }, true));
  }
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

function shouldShowSettingsApiKeyEditor() {
  if (!state.settings || !state.settings.apiKeyConfigured) {
    return true;
  }
  return !!state.settingsApiKeyEditMode;
}

function renderSettingsApiKeyEditor() {
  const field = el('settingsApiKeyField');
  const input = el('settingApiKey');
  const editButton = el('btnEditApiKey');
  const cancelButton = el('btnCancelApiKeyEdit');
  const clearButton = el('btnClearApiKey');
  if (!field || !input) {
    return;
  }

  const apiKeyConfigured = !!state.settings?.apiKeyConfigured;
  const showEditor = shouldShowSettingsApiKeyEditor();

  field.hidden = !showEditor;
  input.disabled = !showEditor;

  if (!showEditor) {
    input.value = '';
  }

  if (editButton) {
    editButton.hidden = !apiKeyConfigured || showEditor;
  }
  if (cancelButton) {
    cancelButton.hidden = !apiKeyConfigured || !showEditor;
  }
  if (clearButton) {
    clearButton.hidden = !apiKeyConfigured || !showEditor;
  }
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

  renderSettingsApiKeyEditor();

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

  // 检查设置完成状态：需要同时有 API key 和绑定
  const setupComplete = state.settings.setupComplete || false;
  const apiKeyConfigured = state.settings.apiKeyConfigured || false;
  const bound = state.settings.bound || false;

  // 情况1: 设置完全完成（API key + 绑定都有）
  if (setupComplete) {
    const masked = state.settings.apiKeyMasked || '';
    const sourceLabel = apiKeySourceLabel(state.settings.apiKeySource);
    if (sourceLabel) {
      apiKeyStatus.textContent = t('settings.apiKeyConfiguredWithSource', { source: sourceLabel, masked });
    } else {
      apiKeyStatus.textContent = t('settings.apiKeyConfigured', { masked });
    }
    if (onboardingTitle) {
      onboardingTitle.textContent = t('settings.setupCompleteTitle');
    }
    setNodeText(onboardingHint, t('settings.setupCompleteHint'));
    if (onboardingExtra) {
      // 优先显示 nickname，如果没有则显示 ownerId
      const ownerNickname = state.settings.ownerNickname || '';
      const ownerId = state.settings.ownerId || '';
      const ownerDisplay = ownerNickname || ownerId || 'N/A';
      onboardingExtra.textContent = t('settings.setupCompleteExtra', { owner: ownerDisplay });
    }
    if (onboardingRepo) {
      onboardingRepo.hidden = false;
    }
    renderSettingsTest();
    return;
  }

  // 情况2: 只有 API key，缺少绑定
  if (apiKeyConfigured && !bound) {
    const masked = state.settings.apiKeyMasked || '';
    const sourceLabel = apiKeySourceLabel(state.settings.apiKeySource);
    if (sourceLabel) {
      apiKeyStatus.textContent = t('settings.apiKeyConfiguredWithSource', { source: sourceLabel, masked });
    } else {
      apiKeyStatus.textContent = t('settings.apiKeyConfigured', { masked });
    }
    if (onboardingTitle) {
      onboardingTitle.textContent = t('settings.needBindingTitle');
    }
    setNodeText(onboardingHint, t('settings.needBindingHint'));
    if (onboardingExtra) {
      onboardingExtra.textContent = t('settings.needBindingExtra');
    }
    if (onboardingRepo) {
      onboardingRepo.hidden = false;
    }
    renderSettingsTest();
    return;
  }

  // 情况3: 只有绑定，缺少 API key
  if (!apiKeyConfigured && bound) {
    apiKeyStatus.textContent = t('settings.apiKeyNotConfigured');
    if (onboardingTitle) {
      onboardingTitle.textContent = t('settings.needApiKeyTitle');
    }
    setNodeText(onboardingHint, t('settings.needApiKeyHint'));
    if (onboardingExtra) {
      onboardingExtra.textContent = t('settings.needApiKeyExtra');
    }
    if (onboardingRepo) {
      onboardingRepo.hidden = true;
    }
    renderSettingsTest();
    return;
  }

  // 情况4: 都没有
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

function utcTodayDate() {
  return new Date().toISOString().slice(0, 10);
}

function utcCurrentMonth() {
  return utcTodayDate().slice(0, 7);
}

function normalizeCalendarMonth(raw) {
  const match = String(raw || '').match(/^(\d{4})-(\d{2})$/);
  if (!match) {
    return '';
  }
  const month = Number.parseInt(match[2], 10);
  if (month < 1 || month > 12) {
    return '';
  }
  return `${match[1]}-${String(month).padStart(2, '0')}`;
}

function shiftCalendarMonth(rawMonth, delta) {
  const normalized = normalizeCalendarMonth(rawMonth) || utcCurrentMonth();
  const match = normalized.match(/^(\d{4})-(\d{2})$/);
  if (!match) {
    return utcCurrentMonth();
  }
  const year = Number.parseInt(match[1], 10);
  const month = Number.parseInt(match[2], 10);
  const dt = new Date(Date.UTC(year, month - 1 + delta, 1));
  return `${dt.getUTCFullYear()}-${String(dt.getUTCMonth() + 1).padStart(2, '0')}`;
}

function formatCalendarMonthLabel(rawMonth) {
  const normalized = normalizeCalendarMonth(rawMonth) || utcCurrentMonth();
  const match = normalized.match(/^(\d{4})-(\d{2})$/);
  if (!match) {
    return normalized;
  }
  const year = Number.parseInt(match[1], 10);
  const month = Number.parseInt(match[2], 10);
  const dt = new Date(Date.UTC(year, month - 1, 1));
  const localeTag = state.locale === 'zh-Hans' ? 'zh-CN' : 'en-US';
  try {
    return new Intl.DateTimeFormat(localeTag, {
      year: 'numeric',
      month: 'long',
      timeZone: 'UTC',
    }).format(dt);
  } catch {
    return normalized;
  }
}

function rebuildDiaryHistoryMap() {
  const map = Object.create(null);
  for (const item of state.diaryHistoryItems) {
    const date = String(item?.date || '').trim();
    if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) {
      continue;
    }
    map[date] = {
      diaryCount: Number.isFinite(item.diaryCount) ? item.diaryCount : 0,
      hasDefault: !!item.hasDefault,
      defaultDiaryId: item.defaultDiaryId || '',
      latestModifiedAt: item.latestModifiedAt || '',
    };
  }
  state.diaryHistoryMap = map;
}

async function loadDiaryHistory() {
  const data = await api('/diaries/history');
  state.diaryHistoryItems = Array.isArray(data.items) ? data.items : [];
  rebuildDiaryHistoryMap();
  if (!state.calendarMonth) {
    state.calendarMonth = utcCurrentMonth();
  }
  if (state.calendarSelectedDate) {
    await selectCalendarDate(state.calendarSelectedDate, state.calendarSelectedDiaryId);
    return;
  }
  renderDiaryHistoryCalendar();
  renderCalendarDiaryDetail();
}

async function ensureCalendarDefaultSelection() {
  if (state.calendarSelectedDate) {
    return;
  }
  if (!Array.isArray(state.diaryHistoryItems) || !state.diaryHistoryItems.length) {
    return;
  }

  const latestWithDate = state.diaryHistoryItems.find((item) => /^\d{4}-\d{2}-\d{2}$/.test(String(item?.date || '')));
  if (!latestWithDate || !latestWithDate.date) {
    return;
  }

  await selectCalendarDate(String(latestWithDate.date), String(latestWithDate.defaultDiaryId || ''), 0);
}

function calendarWeekdayLabels() {
  return [
    t('calendar.weekdayMon'),
    t('calendar.weekdayTue'),
    t('calendar.weekdayWed'),
    t('calendar.weekdayThu'),
    t('calendar.weekdayFri'),
    t('calendar.weekdaySat'),
    t('calendar.weekdaySun'),
  ];
}

function renderCalendarEntryMarkers(date, count) {
  const value = Number.isFinite(count) ? Math.max(0, count) : 0;
  if (value <= 0) {
    return `<span class="calendar-day-status">${escapeHtml(t('calendar.statusNone'))}</span>`;
  }

  const maxDots = 4;
  const dots = [];
  const visibleDots = Math.min(value, maxDots);
  for (let i = 0; i < visibleDots; i += 1) {
    const idx = i + 1;
    const tip = t('calendar.dotTip', { date, index: idx });
    dots.push(
      `<button type="button" class="calendar-entry-dot-btn" data-date="${date}" data-diary-index="${i}" title="${escapeHtml(tip)}" aria-label="${escapeHtml(tip)}"><span class="calendar-entry-dot" aria-hidden="true"></span></button>`,
    );
  }
  const more = value > maxDots
    ? `<button type="button" class="calendar-entry-more-btn" data-date="${date}" title="${escapeHtml(t('calendar.dotMoreTip', { date }))}" aria-label="${escapeHtml(t('calendar.dotMoreTip', { date }))}">+${value - maxDots}</button>`
    : '';
  const aria = value === 1 ? t('calendar.statusSingle') : t('calendar.statusMulti', { count: value });
  return `<span class="calendar-day-status calendar-day-dots" aria-label="${escapeHtml(aria)}">${dots.join('')}${more}</span>`;
}

function renderDiaryHistoryCalendar() {
  const monthLabel = el('calendarMonthLabel');
  const summary = el('calendarSummary');
  const weekdayRow = el('calendarWeekdays');
  const grid = el('calendarGrid');
  if (!monthLabel || !summary || !weekdayRow || !grid) {
    return;
  }

  if (!state.calendarMonth) {
    state.calendarMonth = utcCurrentMonth();
  }
  const normalizedMonth = normalizeCalendarMonth(state.calendarMonth) || utcCurrentMonth();
  state.calendarMonth = normalizedMonth;

  monthLabel.textContent = formatCalendarMonthLabel(normalizedMonth);
  weekdayRow.innerHTML = calendarWeekdayLabels()
    .map((label) => `<span>${escapeHtml(label)}</span>`)
    .join('');

  const match = normalizedMonth.match(/^(\d{4})-(\d{2})$/);
  if (!match) {
    summary.textContent = t('calendar.summaryEmpty');
    grid.innerHTML = '';
    return;
  }

  const year = Number.parseInt(match[1], 10);
  const month = Number.parseInt(match[2], 10);
  const daysInMonth = new Date(Date.UTC(year, month, 0)).getUTCDate();
  const leading = (new Date(Date.UTC(year, month - 1, 1)).getUTCDay() + 6) % 7;
  const today = utcTodayDate();

  let activeDays = 0;
  let totalEntries = 0;
  const cells = [];

  for (let i = 0; i < leading; i += 1) {
    cells.push('<div class="calendar-day-cell is-placeholder" aria-hidden="true"></div>');
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    const date = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
    const info = state.diaryHistoryMap[date];
    const count = Number.isFinite(info?.diaryCount) ? info.diaryCount : 0;
    const hasDefault = !!info?.hasDefault;
    const defaultDiaryId = info?.defaultDiaryId || '';
    const isToday = date === today;
    if (count > 0) {
      activeDays += 1;
      totalEntries += count;
    }

    let status = t('calendar.statusNone');
    let statusClass = 'is-empty';
    if (count === 1) {
      status = t('calendar.statusSingle');
      statusClass = 'is-single';
    } else if (count > 1) {
      status = t('calendar.statusMulti', { count });
      statusClass = 'is-multi';
    }

    const defaultMarker = hasDefault ? '<span class="calendar-default-dot" aria-hidden="true"></span>' : '';
    const todayClass = isToday ? ' is-today' : '';
    const selectedClass = date === state.calendarSelectedDate ? ' is-selected' : '';
    const defaultClass = hasDefault ? ' is-default' : '';
    const content = `
      <span class="calendar-day-num">${day}</span>
      ${renderCalendarEntryMarkers(date, count)}
      ${defaultMarker}
    `;

    if (count > 0) {
      const title = `${date} · ${status}`;
      cells.push(`<div class="calendar-day-cell has-entry ${statusClass}${todayClass}${selectedClass}${defaultClass}" data-date="${date}" data-default-id="${escapeHtml(defaultDiaryId)}" title="${escapeHtml(title)}" role="button" tabindex="0">${content}</div>`);
    } else {
      cells.push(`<div class="calendar-day-cell ${statusClass}${todayClass}${selectedClass}${defaultClass}">${content}</div>`);
    }
  }

  while (cells.length % 7 !== 0) {
    cells.push('<div class="calendar-day-cell is-placeholder" aria-hidden="true"></div>');
  }

  summary.textContent = activeDays > 0
    ? t('calendar.summary', { days: activeDays, entries: totalEntries })
    : t('calendar.summaryEmpty');
  grid.innerHTML = cells.join('');

  grid.querySelectorAll('.calendar-day-cell.has-entry[data-date]').forEach((node) => {
    node.addEventListener('click', () => {
      selectCalendarDate(node.dataset.date || '', node.dataset.defaultId || '')
        .catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    });
    node.addEventListener('keydown', (event) => {
      if (event.key !== 'Enter' && event.key !== ' ') {
        return;
      }
      event.preventDefault();
      selectCalendarDate(node.dataset.date || '', node.dataset.defaultId || '')
        .catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    });
  });

  grid.querySelectorAll('button.calendar-entry-dot-btn[data-date]').forEach((node) => {
    node.addEventListener('click', (event) => {
      event.stopPropagation();
      const idx = Number.parseInt(node.dataset.diaryIndex || '-1', 10);
      selectCalendarDate(node.dataset.date || '', '', Number.isFinite(idx) ? idx : -1)
        .catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    });
  });

  grid.querySelectorAll('button.calendar-entry-more-btn[data-date]').forEach((node) => {
    node.addEventListener('click', (event) => {
      event.stopPropagation();
      selectCalendarDate(node.dataset.date || '')
        .catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    });
  });
}

function buildCalendarDiaryOptionLabel(item) {
  const title = String(item?.title || item?.filename || item?.id || '').trim();
  if (!item?.isDefault) {
    return title;
  }
  return `${title} (${t('diary.defaultTag')})`;
}

function renderCalendarDayList() {
  const wrap = el('calendarDayListWrap');
  const title = el('calendarDayListTitle');
  const summary = el('calendarDayListSummary');
  const list = el('calendarDayList');
  if (!wrap || !title || !summary || !list) {
    return;
  }

  const selectedDate = String(state.calendarSelectedDate || '').trim();
  const items = Array.isArray(state.calendarDateDiaries) ? state.calendarDateDiaries : [];
  if (!selectedDate || items.length <= 1) {
    wrap.hidden = true;
    list.innerHTML = '';
    return;
  }

  wrap.hidden = false;
  title.textContent = t('calendar.dayListTitle', { date: selectedDate });
  summary.textContent = t('calendar.dayListSummary', { count: items.length });
  list.innerHTML = items
    .map((item) => {
      const active = item.id === state.calendarSelectedDiaryId ? ' active' : '';
      const defaultTag = item.isDefault ? `<span class="default-tag">${escapeHtml(t('diary.defaultTag'))}</span>` : '';
      return `
        <button type="button" class="calendar-day-list-item${active}" data-id="${escapeHtml(item.id)}">
          <span>${escapeHtml(buildCalendarDiaryOptionLabel(item))}</span>
          ${defaultTag}
        </button>
      `;
    })
    .join('');

  list.querySelectorAll('button.calendar-day-list-item[data-id]').forEach((node) => {
    node.addEventListener('click', () => {
      loadCalendarDiaryDetail(node.dataset.id || '')
        .catch((err) => setStatusKey('diary.loadFailed', { message: err.message }, true));
    });
  });
}

function applyCalendarDiaryViewModeButton() {
  const button = el('btnCalendarReaderViewMode');
  if (!button) {
    return;
  }
  const hasDiary = !!state.calendarDiaryDetail;
  button.disabled = !hasDiary;
  if (state.calendarDiaryViewMode === 'raw') {
    button.textContent = t('diary.viewReading');
    button.dataset.mode = 'raw';
  } else {
    button.textContent = t('diary.viewRaw');
    button.dataset.mode = 'reading';
  }
}

function renderCalendarDiaryDetail() {
  const dateLabel = el('calendarDetailDate');
  const meta = el('calendarDetailMeta');
  const content = el('calendarDetailContent');
  const openBtn = el('btnCalendarOpenDiaries');
  if (!dateLabel || !meta || !content || !openBtn) {
    return;
  }

  const selectedDate = String(state.calendarSelectedDate || '').trim();
  if (selectedDate) {
    dateLabel.textContent = t('calendar.detailDate', { date: selectedDate });
  } else {
    dateLabel.textContent = t('calendar.detailHint');
  }

  const items = Array.isArray(state.calendarDateDiaries) ? state.calendarDateDiaries : [];
  renderCalendarDayList();
  if (!items.length) {
    meta.textContent = '';
    content.classList.remove('markdown-view');
    content.textContent = selectedDate ? t('calendar.detailEmpty') : t('calendar.detailHint');
    openBtn.disabled = true;
    applyCalendarDiaryViewModeButton();
    return;
  }

  const detail = state.calendarDiaryDetail;
  if (!detail) {
    meta.textContent = '';
    content.classList.remove('markdown-view');
    content.textContent = t('status.loading');
    openBtn.disabled = true;
    applyCalendarDiaryViewModeButton();
    return;
  }

  meta.textContent = t('calendar.metaFormat', {
    date: detail.date || t('common.na'),
    modifiedAt: detail.modifiedAt || '',
    filename: detail.filename || '',
  });
  const raw = detail.content || '';
  if (state.calendarDiaryViewMode === 'markdown') {
    content.classList.add('markdown-view');
    content.innerHTML = markdownToHtml(raw);
  } else {
    content.classList.remove('markdown-view');
    content.textContent = raw;
  }
  openBtn.disabled = false;
  applyCalendarDiaryViewModeButton();
}

async function loadCalendarDiaryDetail(id) {
  const trimmedID = String(id || '').trim();
  if (!trimmedID) {
    state.calendarSelectedDiaryId = '';
    state.calendarDiaryDetail = null;
    renderCalendarDiaryDetail();
    return;
  }
  const detail = await api(`/diaries/${encodeURIComponent(trimmedID)}`);
  state.calendarSelectedDiaryId = detail.id;
  state.calendarDiaryDetail = detail;
  renderCalendarDiaryDetail();
}

async function selectCalendarDate(date, defaultDiaryId = '', preferredIndex = -1) {
  const trimmedDate = String(date || '').trim();
  if (!trimmedDate) {
    return;
  }

  state.calendarSelectedDate = trimmedDate;
  const data = await api(`/diaries?limit=400&q=${encodeURIComponent(trimmedDate)}`);
  const items = (Array.isArray(data.items) ? data.items : []).filter((item) => item.date === trimmedDate);
  state.calendarDateDiaries = items;
  renderDiaryHistoryCalendar();

  if (!items.length) {
    state.calendarSelectedDiaryId = '';
    state.calendarDiaryDetail = null;
    renderCalendarDiaryDetail();
    return;
  }

  let targetID = '';
  if (preferredIndex >= 0 && preferredIndex < items.length && items[preferredIndex]?.id) {
    targetID = items[preferredIndex].id;
  } else if (defaultDiaryId && items.some((item) => item.id === defaultDiaryId)) {
    targetID = defaultDiaryId;
  } else if (state.calendarSelectedDiaryId && items.some((item) => item.id === state.calendarSelectedDiaryId)) {
    targetID = state.calendarSelectedDiaryId;
  } else {
    const defaultItem = items.find((item) => item.isDefault);
    targetID = defaultItem?.id || items[0].id;
  }
  state.calendarSelectedDiaryId = targetID;
  state.calendarDiaryDetail = null;
  renderCalendarDiaryDetail();
  await loadCalendarDiaryDetail(targetID);
}

async function openDiaryDateFromCalendar(date, diaryID = '') {
  const trimmedDate = String(date || '').trim();
  if (!trimmedDate) {
    return;
  }
  const search = el('diarySearch');
  if (search) {
    search.value = trimmedDate;
  }
  switchTab('diaries');
  await loadDiaries();
  if (diaryID && state.diaries.some((item) => item.id === diaryID)) {
    await loadDiaryDetail(diaryID);
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
  await loadDiaryHistory();
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
  await loadDiaryHistory();
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
  await loadDiaryHistory();
  await loadState();
}

async function saveSettings(event) {
  event.preventDefault();

  const payload = {
    cloudSyncEnabled: !!el('settingCloudSync').checked,
  };
  const apiKey = el('settingApiKey').value.trim();
  if (shouldShowSettingsApiKeyEditor() && apiKey) {
    payload.apiKey = apiKey;
  }

  const data = await api('/settings', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
  state.settings = data;
  state.settingsTest = null;
  if (payload.apiKey) {
    state.settingsApiKeyEditMode = false;
  }
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
  state.settingsApiKeyEditMode = true;
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

function resolvePageTitle(tabName) {
  const safeTab = String(tabName || '').trim();
  if (!safeTab) {
    return '';
  }
  const key = `title.page.${safeTab}`;
  const localized = messageFor(state.locale, key);
  if (localized !== key) {
    return localized;
  }
  return safeTab;
}

function updateDocumentTitle() {
  const page = resolvePageTitle(state.currentTab);
  if (!page) {
    document.title = t('page.title');
    return;
  }
  document.title = `${t('title.brand')} · ${page}`;
}

function applyStaticI18n() {
  updateDocumentTitle();
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
  renderDiaryHistoryCalendar();
  renderCalendarDiaryDetail();
  if (state.currentDiaryDetail) {
    renderDiaryMeta();
    renderDiaryContent();
  } else {
    setDiaryEmptyState();
  }

  renderInsightList(filteredInsights());
  if (state.currentInsightDetail || state.insightEditMode) {
    renderInsightMeta();
    renderInsightContent();
  } else {
    setInsightEmptyState();
  }

  renderPromptList();
  fillPromptSelector();
  updatePromptMeta();

  if (!state.hasGeneratedPacket) {
    el('generateResult').textContent = t('generate.noPacket');
  }

  applyDiaryViewModeButton();
  applyInsightViewModeButton();
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

  const btnCalendarPrev = el('btnCalendarPrev');
  const btnCalendarToday = el('btnCalendarToday');
  const btnCalendarNext = el('btnCalendarNext');
  if (btnCalendarPrev) {
    btnCalendarPrev.addEventListener('click', () => {
      state.calendarMonth = shiftCalendarMonth(state.calendarMonth, -1);
      renderDiaryHistoryCalendar();
    });
  }
  if (btnCalendarToday) {
    btnCalendarToday.addEventListener('click', () => {
      state.calendarMonth = utcCurrentMonth();
      renderDiaryHistoryCalendar();
    });
  }
  if (btnCalendarNext) {
    btnCalendarNext.addEventListener('click', () => {
      state.calendarMonth = shiftCalendarMonth(state.calendarMonth, 1);
      renderDiaryHistoryCalendar();
    });
  }

  const btnCalendarReaderViewMode = el('btnCalendarReaderViewMode');
  if (btnCalendarReaderViewMode) {
    btnCalendarReaderViewMode.addEventListener('click', () => {
      state.calendarDiaryViewMode = state.calendarDiaryViewMode === 'raw' ? 'markdown' : 'raw';
      renderCalendarDiaryDetail();
    });
  }

  const btnCalendarOpenDiaries = el('btnCalendarOpenDiaries');
  if (btnCalendarOpenDiaries) {
    btnCalendarOpenDiaries.addEventListener('click', () => {
      openDiaryDateFromCalendar(state.calendarSelectedDate, state.calendarSelectedDiaryId)
        .catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    });
  }

  el('diarySearch').addEventListener('input', () => {
    clearTimeout(bindEvents.diaryTimer);
    bindEvents.diaryTimer = setTimeout(() => {
      loadDiaries().catch((err) => setStatusKey('diary.loadListFailed', { message: err.message }, true));
    }, 220);
  });

  el('insightSearch').addEventListener('input', () => {
    clearTimeout(bindEvents.insightTimer);
    bindEvents.insightTimer = setTimeout(() => {
      ensureInsightsLoaded(true).catch((err) => setStatusKey('insight.loadListFailed', { message: err.message }, true));
    }, 220);
  });

  el('btnInsightReload').addEventListener('click', () => {
    ensureInsightsLoaded(true).catch((err) => setStatusKey('insight.loadListFailed', { message: err.message }, true));
  });

  el('btnInsightViewMode').addEventListener('click', () => {
    state.insightViewMode = state.insightViewMode === 'raw' ? 'markdown' : 'raw';
    renderInsightContent();
  });

  el('btnInsightNew').addEventListener('click', () => {
    beginInsightCreate();
  });

  el('btnInsightEdit').addEventListener('click', () => {
    toggleInsightEditMode();
  });

  el('btnInsightSave').addEventListener('click', () => {
    saveInsight().catch((err) => setStatusKey('insight.saveFailed', { message: err.message }, true));
  });

  el('btnInsightDelete').addEventListener('click', () => {
    deleteInsight().catch((err) => setStatusKey('insight.deleteFailed', { message: err.message }, true));
  });

  el('insightEditor').addEventListener('input', (event) => {
    state.insightDraftContent = event.target.value;
  });

  el('insightTitle').addEventListener('input', (event) => {
    state.insightDraftTitle = event.target.value;
  });

  el('insightDiaryId').addEventListener('input', (event) => {
    state.insightDraftDiaryId = event.target.value;
  });

  el('insightTags').addEventListener('input', (event) => {
    state.insightDraftTags = event.target.value;
  });

  el('insightCatalogs').addEventListener('input', (event) => {
    state.insightDraftCatalogs = event.target.value;
  });

  el('insightVisibility').addEventListener('change', (event) => {
    const value = Number.parseInt(String(event.target.value || '0'), 10);
    state.insightDraftVisibility = Number.isFinite(value) ? value : 0;
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

  el('btnEditApiKey').addEventListener('click', () => {
    state.settingsApiKeyEditMode = true;
    renderSettings();
    const input = el('settingApiKey');
    if (input && !input.hidden && !input.disabled) {
      input.focus();
    }
  });

  el('btnCancelApiKeyEdit').addEventListener('click', () => {
    state.settingsApiKeyEditMode = false;
    const input = el('settingApiKey');
    if (input) {
      input.value = '';
    }
    renderSettings();
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
  state.insightsLoaded = false;
  await loadState();
  await loadDiaries();
  await loadDiaryHistory();
  await loadPrompts();
  await loadSettings();
  if (state.currentTab === 'insights') {
    await ensureInsightsLoaded(true);
  }
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
  setInsightEmptyState();
  el('generateResult').textContent = t('generate.noPacket');
  applyDiaryViewModeButton();
  applyInsightViewModeButton();
  applyCalendarDiaryViewModeButton();
  renderCalendarDiaryDetail();
  try {
    await bootstrap();
  } catch (err) {
    setStatusKey('status.initFailed', { message: err.message }, true);
  }
});
