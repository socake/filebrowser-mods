package fbhttp

const docsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Doc Portal</title>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
:root {
  --bg: #f8f9fa; --sidebar-bg: #fff; --card-bg: #fff;
  --text: #1a1a2e; --text-secondary: #6c757d; --border: #e9ecef;
  --primary: #4361ee; --primary-light: #eef0ff; --primary-dark: #3a56d4;
  --success: #2ec4b6; --warning: #ff9f1c; --hover: #f1f3f5;
  --radius: 8px; --shadow: 0 1px 3px rgba(0,0,0,.08);
  --shadow-hover: 0 4px 12px rgba(0,0,0,.12);
}
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: var(--bg); color: var(--text); display: flex; height: 100vh; overflow: hidden; }

/* Sidebar */
.sidebar { width: 280px; min-width: 280px; background: var(--sidebar-bg); border-right: 1px solid var(--border); display: flex; flex-direction: column; overflow: hidden; }
.sidebar-header { padding: 20px; border-bottom: 1px solid var(--border); }
.sidebar-header h1 { font-size: 18px; font-weight: 700; color: var(--primary); display: flex; align-items: center; gap: 8px; }
.sidebar-header h1::before { content: "📄"; }
.sidebar-search { padding: 12px 16px; border-bottom: 1px solid var(--border); }
.sidebar-search input { width: 100%; padding: 8px 12px; border: 1px solid var(--border); border-radius: var(--radius); font-size: 13px; outline: none; transition: border-color .2s; }
.sidebar-search input:focus { border-color: var(--primary); }
.sidebar-tree { flex: 1; overflow-y: auto; padding: 8px; }

.tree-item { padding: 7px 12px; border-radius: 6px; cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text); transition: background .15s; user-select: none; }
.tree-item:hover { background: var(--hover); }
.tree-item.active { background: var(--primary-light); color: var(--primary); font-weight: 500; }
.tree-item .icon { font-size: 15px; flex-shrink: 0; }
.tree-item .name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.tree-children { padding-left: 16px; }
.tree-toggle { font-size: 10px; color: var(--text-secondary); transition: transform .2s; width: 16px; text-align: center; flex-shrink: 0; }
.tree-toggle.open { transform: rotate(90deg); }

/* Main */
.main { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.topbar { padding: 16px 24px; background: var(--sidebar-bg); border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; gap: 16px; min-height: 60px; }
.breadcrumb { display: flex; align-items: center; gap: 4px; font-size: 13px; flex-wrap: wrap; }
.breadcrumb a { color: var(--text-secondary); text-decoration: none; }
.breadcrumb a:hover { color: var(--primary); }
.breadcrumb .sep { color: var(--border); }
.breadcrumb .current { color: var(--text); font-weight: 500; }
.actions { display: flex; gap: 8px; flex-shrink: 0; }
.btn { padding: 7px 14px; border-radius: 6px; border: 1px solid var(--border); background: #fff; font-size: 12px; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all .15s; color: var(--text); white-space: nowrap; }
.btn:hover { border-color: var(--primary); color: var(--primary); }
.btn-primary { background: var(--primary); color: #fff; border-color: var(--primary); }
.btn-primary:hover { background: var(--primary-dark); }
.btn .ico { font-size: 14px; }

.content { flex: 1; overflow-y: auto; padding: 24px; }

/* Cards grid - landing page */
.cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 16px; }
.card { background: var(--card-bg); border: 1px solid var(--border); border-radius: var(--radius); padding: 20px; cursor: pointer; transition: all .2s; }
.card:hover { border-color: var(--primary); box-shadow: var(--shadow-hover); transform: translateY(-2px); }
.card-icon { font-size: 28px; margin-bottom: 12px; }
.card-name { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.card-meta { font-size: 12px; color: var(--text-secondary); }

/* File list */
.file-list { background: var(--card-bg); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
.file-row { display: flex; align-items: center; padding: 10px 16px; border-bottom: 1px solid var(--border); cursor: pointer; transition: background .15s; gap: 12px; }
.file-row:last-child { border-bottom: none; }
.file-row:hover { background: var(--hover); }
.file-row .file-icon { font-size: 18px; flex-shrink: 0; }
.file-row .file-name { flex: 1; font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.file-row .file-meta { font-size: 12px; color: var(--text-secondary); flex-shrink: 0; display: flex; gap: 16px; }
.file-row .badge { padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 500; }
.badge-md { background: #e8f5e9; color: #2e7d32; }
.badge-txt { background: #e3f2fd; color: #1565c0; }
.badge-code { background: #fff3e0; color: #e65100; }

/* MD Preview */
.preview-container { background: var(--card-bg); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
.preview-header { padding: 12px 20px; background: #fafbfc; border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; }
.preview-header .file-title { font-size: 14px; font-weight: 600; display: flex; align-items: center; gap: 8px; }
.preview-tabs { display: flex; gap: 0; }
.preview-tab { padding: 6px 16px; font-size: 12px; cursor: pointer; border: 1px solid var(--border); background: #fff; color: var(--text-secondary); transition: all .15s; }
.preview-tab:first-child { border-radius: 6px 0 0 6px; }
.preview-tab:last-child { border-radius: 0 6px 6px 0; }
.preview-tab.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.preview-body { padding: 24px 32px; max-width: 900px; overflow-y: auto; max-height: calc(100vh - 220px); }
.preview-raw { padding: 16px; max-height: calc(100vh - 220px); overflow: auto; }
.preview-raw pre { white-space: pre-wrap; word-break: break-word; font-size: 13px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; line-height: 1.6; color: var(--text); }

/* MD rendered styles */
.md-body { line-height: 1.7; font-size: 15px; }
.md-body h1 { font-size: 28px; font-weight: 700; margin: 24px 0 16px; padding-bottom: 8px; border-bottom: 2px solid var(--border); }
.md-body h2 { font-size: 22px; font-weight: 600; margin: 20px 0 12px; padding-bottom: 6px; border-bottom: 1px solid var(--border); }
.md-body h3 { font-size: 18px; font-weight: 600; margin: 16px 0 8px; }
.md-body h4 { font-size: 15px; font-weight: 600; margin: 12px 0 6px; }
.md-body p { margin: 0 0 12px; }
.md-body a { color: var(--primary); text-decoration: none; }
.md-body a:hover { text-decoration: underline; }
.md-body code { background: #f0f1f3; padding: 2px 6px; border-radius: 4px; font-size: 13px; font-family: "SFMono-Regular", Consolas, monospace; }
.md-body pre { background: #f6f8fa; border: 1px solid var(--border); border-radius: var(--radius); padding: 16px; overflow-x: auto; margin: 0 0 16px; }
.md-body pre code { background: none; padding: 0; font-size: 13px; line-height: 1.5; }
.md-body blockquote { border-left: 4px solid var(--primary); padding: 8px 16px; margin: 0 0 12px; background: var(--primary-light); color: var(--text-secondary); }
.md-body ul, .md-body ol { padding-left: 24px; margin: 0 0 12px; }
.md-body li { margin: 4px 0; }
.md-body table { border-collapse: collapse; width: 100%; margin: 0 0 16px; }
.md-body th, .md-body td { border: 1px solid var(--border); padding: 8px 12px; text-align: left; font-size: 14px; }
.md-body th { background: #f6f8fa; font-weight: 600; }
.md-body tr:nth-child(even) { background: #fafbfc; }
.md-body img { max-width: 100%; border-radius: var(--radius); }
.md-body hr { border: none; border-top: 1px solid var(--border); margin: 24px 0; }

/* Toast */
.toast { position: fixed; bottom: 24px; right: 24px; padding: 10px 20px; background: #333; color: #fff; border-radius: var(--radius); font-size: 13px; opacity: 0; transition: opacity .3s; pointer-events: none; z-index: 999; }
.toast.show { opacity: 1; }

/* Loading */
.loading { text-align: center; padding: 60px; color: var(--text-secondary); }
.loading::after { content: ""; display: block; width: 28px; height: 28px; margin: 12px auto; border: 3px solid var(--border); border-top-color: var(--primary); border-radius: 50%; animation: spin .6s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

/* Empty state */
.empty { text-align: center; padding: 60px 20px; color: var(--text-secondary); }
.empty-icon { font-size: 48px; margin-bottom: 12px; }

/* Responsive */
@media (max-width: 768px) {
  .sidebar { width: 240px; min-width: 240px; }
  .content { padding: 16px; }
  .preview-body { padding: 16px; }
}
</style>
</head>
<body>

<div class="sidebar">
  <div class="sidebar-header"><h1>Doc Portal</h1></div>
  <div class="sidebar-search"><input id="searchInput" type="text" placeholder="搜索文件..."></div>
  <div class="sidebar-tree" id="sidebarTree"></div>
</div>

<div class="main">
  <div class="topbar">
    <div class="breadcrumb" id="breadcrumb"></div>
    <div class="actions" id="actions"></div>
  </div>
  <div class="content" id="content"><div class="loading">加载中</div></div>
</div>

<div class="toast" id="toast"></div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/marked/12.0.0/marked.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
<script>
const API = '/docs';
let currentPath = '/';
let currentView = 'list'; // list | preview
let rawContent = '';
let previewMode = 'rendered'; // rendered | raw
let treeCache = {};

// Marked config
marked.setOptions({
  highlight: (code, lang) => {
    if (lang && hljs.getLanguage(lang)) return hljs.highlight(code, { language: lang }).value;
    return hljs.highlightAuto(code).value;
  },
  breaks: true
});

function toast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2000);
}

function fileIcon(entry) {
  if (entry.isDir) return '📁';
  const ext = entry.ext;
  if (['.md','.markdown'].includes(ext)) return '📝';
  if (['.js','.ts','.go','.py','.java','.rs','.c','.cpp','.h'].includes(ext)) return '💻';
  if (['.json','.yaml','.yml','.toml','.xml'].includes(ext)) return '⚙️';
  if (['.sh','.bash','.zsh'].includes(ext)) return '🔧';
  if (['.png','.jpg','.jpeg','.gif','.svg','.webp'].includes(ext)) return '🖼️';
  if (['.pdf'].includes(ext)) return '📕';
  if (['.txt','.log','.csv'].includes(ext)) return '📄';
  return '📎';
}

function fileBadge(ext) {
  if (['.md','.markdown'].includes(ext)) return '<span class="badge badge-md">MD</span>';
  if (['.txt','.log','.csv'].includes(ext)) return '<span class="badge badge-txt">TXT</span>';
  if (['.js','.ts','.go','.py','.java','.rs','.sh','.c','.cpp'].includes(ext)) return '<span class="badge badge-code">CODE</span>';
  return '';
}

function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / 1048576).toFixed(1) + ' MB';
}

function formatTime(iso) {
  const d = new Date(iso);
  return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', {hour:'2-digit',minute:'2-digit'});
}

function isPreviewable(ext) {
  return ['.md','.markdown','.txt','.log','.json','.yaml','.yml','.toml','.xml',
          '.js','.ts','.go','.py','.java','.rs','.sh','.bash','.c','.cpp','.h',
          '.css','.html','.sql','.csv','.ini','.conf','.cfg','.env.example','.dockerfile'].includes(ext);
}

function isMarkdown(ext) {
  return ['.md','.markdown'].includes(ext);
}

// Breadcrumb
function renderBreadcrumb(filePath, isFile) {
  const bc = document.getElementById('breadcrumb');
  const parts = filePath.split('/').filter(Boolean);
  let html = '<a href="#" onclick="navigate(\'/\');return false">🏠 首页</a>';
  let accumulated = '';
  parts.forEach((p, i) => {
    accumulated += '/' + p;
    const isLast = i === parts.length - 1;
    html += '<span class="sep">/</span>';
    if (isLast && isFile) {
      html += '<span class="current">' + p + '</span>';
    } else if (isLast) {
      html += '<span class="current">' + p + '</span>';
    } else {
      const path = accumulated;
      html += '<a href="#" onclick="navigate(\'' + path + '\');return false">' + p + '</a>';
    }
  });
  bc.innerHTML = html;
}

// Actions
function renderActions(entry) {
  const acts = document.getElementById('actions');
  if (!entry) { acts.innerHTML = ''; return; }
  let html = '';
  html += '<button class="btn" onclick="downloadFile(\'' + entry.path + '\')"><span class="ico">⬇️</span>下载</button>';
  if (isPreviewable(entry.ext)) {
    html += '<button class="btn" onclick="copyRaw()"><span class="ico">📋</span>复制原文</button>';
  }
  const dlUrl = location.origin + API + '/download?path=' + encodeURIComponent(entry.path);
  html += '<button class="btn" onclick="copyLink(\'' + dlUrl + '\')"><span class="ico">🔗</span>复制链接</button>';
  acts.innerHTML = html;
}

// Sidebar tree
async function loadTree(dirPath, parentEl, depth) {
  if (depth > 3) return;
  try {
    const resp = await fetch(API + '/list?path=' + encodeURIComponent(dirPath));
    const data = await resp.json();
    treeCache[dirPath] = data.entries;

    data.entries.forEach(entry => {
      const item = document.createElement('div');

      if (entry.isDir) {
        const row = document.createElement('div');
        row.className = 'tree-item' + (entry.path === currentPath ? ' active' : '');
        row.innerHTML = '<span class="tree-toggle">▶</span><span class="icon">📁</span><span class="name">' + entry.name + '</span>';
        item.appendChild(row);

        const children = document.createElement('div');
        children.className = 'tree-children';
        children.style.display = 'none';
        item.appendChild(children);

        let loaded = false;
        row.onclick = (e) => {
          e.stopPropagation();
          const toggle = row.querySelector('.tree-toggle');
          if (children.style.display === 'none') {
            children.style.display = 'block';
            toggle.classList.add('open');
            if (!loaded) { loadTree(entry.path, children, depth + 1); loaded = true; }
          } else {
            children.style.display = 'none';
            toggle.classList.remove('open');
          }
          navigate(entry.path);
        };
      } else {
        const row = document.createElement('div');
        row.className = 'tree-item';
        row.innerHTML = '<span class="icon">' + fileIcon(entry) + '</span><span class="name">' + entry.name + '</span>';
        row.onclick = (e) => {
          e.stopPropagation();
          if (isPreviewable(entry.ext)) openPreview(entry);
          else downloadFile(entry.path);
        };
        item.appendChild(row);
      }

      parentEl.appendChild(item);
    });
  } catch (err) {
    console.error('loadTree error:', err);
  }
}

// Navigate to directory
async function navigate(dirPath) {
  currentPath = dirPath;
  currentView = 'list';
  renderBreadcrumb(dirPath, false);
  document.getElementById('actions').innerHTML = '';

  const content = document.getElementById('content');
  content.innerHTML = '<div class="loading">加载中</div>';

  try {
    const resp = await fetch(API + '/list?path=' + encodeURIComponent(dirPath));
    const data = await resp.json();

    if (dirPath === '/' && data.entries.some(e => e.isDir)) {
      renderLanding(data.entries);
    } else {
      renderFileList(data.entries);
    }
  } catch (err) {
    content.innerHTML = '<div class="empty"><div class="empty-icon">❌</div>加载失败</div>';
  }
}

// Landing page with cards
function renderLanding(entries) {
  const content = document.getElementById('content');
  const dirs = entries.filter(e => e.isDir);
  const files = entries.filter(e => !e.isDir);

  let html = '';
  if (dirs.length) {
    html += '<h3 style="margin-bottom:16px;font-size:16px;color:var(--text-secondary)">📂 目录</h3>';
    html += '<div class="cards">';
    dirs.forEach(d => {
      html += '<div class="card" onclick="navigate(\'' + d.path + '\')">';
      html += '<div class="card-icon">📁</div>';
      html += '<div class="card-name">' + d.name + '</div>';
      html += '<div class="card-meta">' + formatTime(d.modTime) + '</div>';
      html += '</div>';
    });
    html += '</div>';
  }
  if (files.length) {
    html += '<h3 style="margin:24px 0 16px;font-size:16px;color:var(--text-secondary)">📄 文件</h3>';
    html += '<div class="file-list">';
    files.forEach(f => {
      html += '<div class="file-row" onclick="' + (isPreviewable(f.ext) ? 'openPreview(' + JSON.stringify(f).replace(/"/g,'&quot;') + ')' : 'downloadFile(\'' + f.path + '\')') + '">';
      html += '<span class="file-icon">' + fileIcon(f) + '</span>';
      html += '<span class="file-name">' + f.name + '</span>';
      html += '<span class="file-meta">' + fileBadge(f.ext) + '<span>' + formatSize(f.size) + '</span><span>' + formatTime(f.modTime) + '</span></span>';
      html += '</div>';
    });
    html += '</div>';
  }
  if (!dirs.length && !files.length) {
    html = '<div class="empty"><div class="empty-icon">📭</div>空目录</div>';
  }
  content.innerHTML = html;
}

function renderFileList(entries) {
  const content = document.getElementById('content');
  if (!entries.length) {
    content.innerHTML = '<div class="empty"><div class="empty-icon">📭</div>空目录</div>';
    return;
  }
  let html = '<div class="file-list">';

  // Parent dir
  if (currentPath !== '/') {
    const parent = currentPath.split('/').slice(0, -1).join('/') || '/';
    html += '<div class="file-row" onclick="navigate(\'' + parent + '\')"><span class="file-icon">⬆️</span><span class="file-name">..</span><span class="file-meta"></span></div>';
  }

  entries.forEach(f => {
    const action = f.isDir
      ? 'navigate(\'' + f.path + '\')'
      : (isPreviewable(f.ext) ? 'openPreview(' + JSON.stringify(f).replace(/"/g,'&quot;') + ')' : 'downloadFile(\'' + f.path + '\')');
    html += '<div class="file-row" onclick="' + action + '">';
    html += '<span class="file-icon">' + fileIcon(f) + '</span>';
    html += '<span class="file-name">' + f.name + '</span>';
    html += '<span class="file-meta">' + fileBadge(f.ext) + '<span>' + formatSize(f.size) + '</span><span>' + formatTime(f.modTime) + '</span></span>';
    html += '</div>';
  });
  html += '</div>';
  content.innerHTML = html;
}

// Preview
async function openPreview(entry) {
  currentView = 'preview';
  previewMode = 'rendered';
  renderBreadcrumb(entry.path, true);
  renderActions(entry);

  const content = document.getElementById('content');
  content.innerHTML = '<div class="loading">加载中</div>';

  try {
    const resp = await fetch(API + '/content?path=' + encodeURIComponent(entry.path));
    rawContent = await resp.text();

    renderPreview(entry, rawContent);
  } catch (err) {
    content.innerHTML = '<div class="empty"><div class="empty-icon">❌</div>加载失败</div>';
  }
}

function renderPreview(entry, text) {
  const content = document.getElementById('content');
  const isMd = isMarkdown(entry.ext);
  const renderedHTML = isMd ? marked.parse(text) : '<pre><code>' + hljs.highlightAuto(text).value + '</code></pre>';

  let html = '<div class="preview-container">';
  html += '<div class="preview-header">';
  html += '<div class="file-title">' + fileIcon(entry) + ' ' + entry.name + ' <span style="color:var(--text-secondary);font-weight:400;font-size:12px">' + formatSize(entry.size) + '</span></div>';

  if (isMd) {
    html += '<div class="preview-tabs">';
    html += '<div class="preview-tab' + (previewMode === 'rendered' ? ' active' : '') + '" onclick="switchPreviewMode(\'rendered\')">渲染预览</div>';
    html += '<div class="preview-tab' + (previewMode === 'raw' ? ' active' : '') + '" onclick="switchPreviewMode(\'raw\')">原始内容</div>';
    html += '</div>';
  }
  html += '</div>';

  if (previewMode === 'rendered' && isMd) {
    html += '<div class="preview-body"><div class="md-body">' + renderedHTML + '</div></div>';
  } else {
    html += '<div class="preview-raw"><pre>' + escapeHtml(text) + '</pre></div>';
  }

  html += '</div>';
  content.innerHTML = html;

  // Store entry for tab switching
  content.dataset.entry = JSON.stringify(entry);
}

function switchPreviewMode(mode) {
  previewMode = mode;
  const content = document.getElementById('content');
  const entry = JSON.parse(content.dataset.entry);
  renderPreview(entry, rawContent);
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Copy raw markdown
function copyRaw() {
  navigator.clipboard.writeText(rawContent).then(() => toast('已复制原始内容'));
}

function copyLink(url) {
  navigator.clipboard.writeText(url).then(() => toast('已复制下载链接'));
}

function downloadFile(filePath) {
  window.open(API + '/download?path=' + encodeURIComponent(filePath), '_blank');
}

// Search
document.getElementById('searchInput').addEventListener('input', function(e) {
  const q = e.target.value.toLowerCase().trim();
  document.querySelectorAll('.sidebar-tree .tree-item').forEach(item => {
    const name = item.querySelector('.name');
    if (!name) return;
    const match = !q || name.textContent.toLowerCase().includes(q);
    item.style.display = match ? '' : 'none';
  });
});

// Hash routing: #preview:/path/to/file.md or #dir:/path/to/dir
function handleHash() {
  const hash = decodeURIComponent(location.hash.slice(1));
  if (hash.startsWith('preview:')) {
    const filePath = hash.slice(8);
    const name = filePath.split('/').pop();
    const ext = '.' + name.split('.').pop().toLowerCase();
    openPreview({ name, path: filePath, ext, size: 0, isDir: false, modTime: '' });
    return true;
  } else if (hash.startsWith('dir:')) {
    navigate(hash.slice(4));
    return true;
  }
  return false;
}

window.addEventListener('hashchange', handleHash);

// Init
(async function init() {
  const tree = document.getElementById('sidebarTree');
  await loadTree('/', tree, 0);
  if (!handleHash()) navigate('/');
})();
</script>
</body>
</html>`
