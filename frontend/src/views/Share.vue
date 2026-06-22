<template>
  <div>
    <template v-if="req && req.isDir && !isDownloadShare">
      <header-bar showMenu showLogo>
        <title />

        <action
          v-if="fileStore.selectedCount"
          icon="file_download"
          :label="t('buttons.download')"
          @action="download"
          :counter="fileStore.selectedCount"
        />
        <button
          v-if="isSingleFile()"
          class="action copy-clipboard"
          :aria-label="t('buttons.copyDownloadLinkToClipboard')"
          :data-title="t('buttons.copyDownloadLinkToClipboard')"
          @click="copyToClipboard(linkSelected())"
        >
          <i class="material-icons">content_paste</i>
        </button>
        <action
          icon="check_circle"
          :label="t('buttons.selectMultiple')"
          @action="toggleMultipleSelection"
        />
      </header-bar>

      <breadcrumbs :base="'/share/' + hash" />
    </template>

    <div v-if="layoutStore.loading">
      <h2 class="message delayed" style="padding-top: 3em !important">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>
    <div v-else-if="error">
      <div v-if="error.status === 401" class="ls-pw-page">
        <div class="ls-pw-card">
          <div class="ls-pw-ic"><i class="material-icons">lock</i></div>
          <h2 class="ls-pw-title">{{ t("login.password") }}</h2>
          <p class="ls-pw-hint">{{ t("prompts.sharePasswordRequired") }}</p>
          <input
            v-focus
            class="ls-pw-field"
            type="password"
            :placeholder="t('login.password')"
            v-model="password"
            @keyup.enter="fetchData"
          />
          <div v-if="attemptedPasswordLogin" class="ls-pw-wrong">
            {{ t("login.wrongCredentials") }}
          </div>
          <button
            class="ls-pw-submit"
            @click="fetchData"
            :aria-label="t('buttons.submit')"
            :data-title="t('buttons.submit')"
          >
            {{ t("buttons.submit") }}
          </button>
        </div>
      </div>
      <errors v-else :errorCode="error.status" />
    </div>
    <div v-else-if="req !== null">
      <!-- Download share mode: auto-triggers download, no preview -->
      <div v-if="isDownloadShare" class="share-download">
        <div class="share-download-card">
          <i class="material-icons share-download-ic">file_download</i>
          <h2>{{ t("prompts.shareDownloadStarting") }}</h2>
          <p>{{ req.name }}</p>
          <a :href="link" class="button button--flat">
            <i class="material-icons">file_download</i>
            {{ t("prompts.shareDownloadManual") }}
          </a>
        </div>
      </div>

      <!-- Markdown preview mode -->
      <div v-else-if="isMarkdownFile && req.content" class="share-md-preview">
        <div class="share-md-header">
          <h3>{{ req.name }}</h3>
          <div class="share-md-actions">
            <button class="share-md-btn" @click="decreasePreviewFont">
              <i class="material-icons">remove</i>
            </button>
            <span class="share-md-fontsize">{{ previewFontSize }}px</span>
            <button class="share-md-btn" @click="increasePreviewFont">
              <i class="material-icons">add</i>
            </button>
            <button class="share-md-btn" @click="copyAllContent">
              <i class="material-icons">{{ copyAllIcon }}</i>
              {{ copyAllLabel }}
            </button>
            <a :href="link" class="share-md-btn" target="_blank">
              <i class="material-icons">file_download</i>
              {{ t("buttons.download") }}
            </a>
          </div>
        </div>
        <div class="share-md-body">
          <div
            id="share-preview-container"
            class="md_preview"
            v-html="renderedMarkdown"
          ></div>
        </div>
      </div>

      <!-- HTML preview mode -->
      <div v-else-if="isHtmlFile && req.content" class="share-md-preview">
        <div class="share-md-header">
          <h3>{{ req.name }}</h3>
          <div class="share-md-actions">
            <button class="share-md-btn" @click="copyAllContent">
              <i class="material-icons">{{ copyAllIcon }}</i>
              {{ copyAllLabel }}
            </button>
            <a :href="link" class="share-md-btn" target="_blank">
              <i class="material-icons">file_download</i>
              {{ t("buttons.download") }}
            </a>
            <button v-if="snapshotHash" class="share-md-btn" @click="downloadSnapshot">
              <i class="material-icons">image</i>
              下载图片
            </button>
          </div>
        </div>
        <div class="share-html-body">
          <iframe
            ref="htmlPreviewFrame"
            :srcdoc="htmlPreviewContent"
            sandbox="allow-scripts allow-same-origin allow-popups"
            class="share-html-iframe"
            @load="resizeHtmlFrame"
          ></iframe>
        </div>
      </div>

      <!-- Folder listing -->
      <div v-else-if="req.isDir" class="ls-page">
        <div id="shareList" class="ls-folder">
          <div
            id="listing"
            class="list file-icons"
            v-if="req.items.length > 0"
          >
            <item
              v-for="item in req.items.slice(0, showLimit)"
              :key="base64(item.name)"
              v-bind:index="item.index"
              v-bind:name="item.name"
              v-bind:isDir="item.isDir"
              v-bind:url="item.url"
              v-bind:modified="item.modified"
              v-bind:type="item.type"
              v-bind:size="item.size"
              readOnly
            >
            </item>
            <div
              v-if="req.items.length > showLimit"
              class="item"
              @click="showLimit += 100"
            >
              <div>
                <p class="name">+ {{ req.items.length - showLimit }}</p>
              </div>
            </div>

            <div
              :class="{ active: fileStore.multiple }"
              id="multiple-selection"
            >
              <p>{{ t("files.multipleSelectionEnabled") }}</p>
              <div
                @click="() => (fileStore.multiple = false)"
                tabindex="0"
                role="button"
                :data-title="t('buttons.clear')"
                :aria-label="t('buttons.clear')"
                class="action"
              >
                <i class="material-icons">clear</i>
              </div>
            </div>
          </div>
          <h2 v-else class="ls-empty">
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ t("files.lonely") }}</span>
          </h2>
        </div>
      </div>

      <!-- Generic single-file preview -->
      <div v-else class="ls-page">
        <div class="ls-ph">
          <div class="ls-ph-name">
            <i class="material-icons">{{ icon }}</i>
            <span :title="req.name">{{ req.name }}</span>
          </div>
          <div class="ls-ph-actions">
            <a :href="link" class="ls-ph-btn ls-ph-primary">
              <i class="material-icons">file_download</i>
              <span>{{ t("buttons.download") }}</span>
            </a>
            <button class="ls-ph-btn" @click="copyToClipboard(shareLink)">
              <i class="material-icons">link</i>
              <span>{{ t("buttons.copyToClipboard") }}</span>
            </button>
          </div>
        </div>

        <div class="ls-pv">
          <img
            v-if="previewKind === 'image'"
            :src="inlineLink"
            :alt="req.name"
            class="ls-pv-img"
          />
          <video
            v-else-if="previewKind === 'video'"
            :src="inlineLink"
            controls
            class="ls-pv-video"
          >
            {{ t("prompts.shareDownloadManual") }}
          </video>
          <div v-else-if="previewKind === 'audio'" class="ls-pv-audio-wrap">
            <i class="material-icons ls-pv-audio-ic">volume_up</i>
            <audio :src="inlineLink" controls class="ls-pv-audio"></audio>
          </div>
          <iframe
            v-else-if="previewKind === 'pdf'"
            :src="inlineLink"
            class="ls-pv-pdf"
          ></iframe>
          <pre v-else-if="previewKind === 'text'" class="ls-pv-text">{{
            textContent
          }}</pre>
          <office-preview
            v-else-if="previewKind === 'office'"
            :src="inlineLink"
            :ext="officeExt"
            class="ls-pv-office"
          />

          <!-- Unpreviewable: elegant file card -->
          <div v-else class="ls-fc">
            <div class="ls-fc-ic"><i class="material-icons">{{ icon }}</i></div>
            <div class="ls-fc-badge">{{ fileExtBadge }}</div>
            <div class="ls-fc-name" :title="req.name">{{ req.name }}</div>
            <div class="ls-fc-meta">
              {{ humanSize }} · {{ humanTime }}
            </div>
            <div class="ls-fc-actions">
              <a :href="link" class="ls-fc-dl">
                <i class="material-icons">file_download</i>
                {{ t("buttons.download") }}
              </a>
              <button class="ls-fc-copy" @click="copyToClipboard(shareLink)">
                <i class="material-icons">link</i>
                {{ t("buttons.copyToClipboard") }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { pub as api } from "@/api";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { Base64 } from "js-base64";
import { createURL } from "@/api/utils";
import HeaderBar from "@/components/header/HeaderBar.vue";
import OfficePreview from "@/components/OfficePreview.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem.vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { computed, inject, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy, copyToClipboardWithFallback } from "@/utils/clipboard";
import { getExtBadge } from "@/utils/fileType";
import { marked } from "marked";
import markedKatex from "marked-katex-extension";
import DOMPurify from "dompurify";

const error = ref<StatusError | null>(null);
const showLimit = ref<number>(100);
const password = ref<string>("");
const attemptedPasswordLogin = ref<boolean>(false);
const hash = ref<string>("");
const token = ref<string>("");
const audio = ref<HTMLAudioElement>();
const tag = ref<boolean>(false);

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n({});

const route = useRoute();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

watch(route, () => {
  showLimit.value = 100;
  fetchData();
});

const req = computed(() => fileStore.req);

const isDownloadShare = computed(
  () => (req.value as any)?.shareType === "download"
);
let downloadStarted = false;

// Define computes

const icon = computed(() => {
  if (req.value === null) return "insert_drive_file";
  if (req.value.isDir) return "folder";
  if (req.value.type === "image") return "insert_photo";
  if (req.value.type === "audio") return "volume_up";
  if (req.value.type === "video") return "movie";
  return "insert_drive_file";
});

const link = computed(() => (req.value ? api.getDownloadURL(req.value) : ""));
const raw = computed(() => {
  if (!req.value || !req.value.items[fileStore.selected[0]]) return "";
  return createURL(
    `api/public/dl/${hash.value}${req.value.items[fileStore.selected[0]].path}`,
    { token: token.value }
  );
});
const inlineLink = computed(() =>
  req.value ? api.getDownloadURL(req.value, true) : ""
);
const humanSize = computed(() => {
  if (req.value) {
    return req.value.isDir
      ? req.value.items.length
      : filesize(req.value.size ?? 0);
  } else {
    return "";
  }
});
const humanTime = computed(() => dayjs(req.value?.modified).fromNow());
const modTime = computed(() =>
  req.value
    ? new Date(Date.parse(req.value.modified)).toLocaleString()
    : new Date().toLocaleString()
);

// Link to the current share page (single, correct URL — used by copy button)
const shareLink = computed(() =>
  typeof window !== "undefined" ? window.location.href : ""
);

const fileExtBadge = computed(() => getExtBadge(req.value?.extension));

const isPdfFile = computed(
  () =>
    req.value &&
    !req.value.isDir &&
    (req.value.type === "pdf" || req.value.name.toLowerCase().endsWith(".pdf"))
);

// Office documents (docx/xlsx/pptx) — previewed via vue-office.
const officeExt = computed(() => {
  const n = req.value?.name.toLowerCase() || "";
  const m = n.match(/\.(docx?|xlsx?|xlsm|pptx?)$/);
  return m ? m[1] : "";
});
const isOfficeFile = computed(
  () => !!req.value && !req.value.isDir && officeExt.value !== ""
);

// Decide how to preview a single shared file.
const previewKind = computed(() => {
  const r = req.value;
  if (!r || r.isDir) return "other";
  // markdown / html have dedicated branches in the template
  if (isMarkdownFile.value || isHtmlFile.value) return "other";
  if (r.type === "image") return "image";
  if (r.type === "video") return "video";
  if (r.type === "audio") return "audio";
  if (isPdfFile.value) return "pdf";
  if (isOfficeFile.value) return "office";
  if (r.type === "text" || r.type === "textImmutable") return "text";
  return "other";
});

// Text content: prefer inline content from the API, otherwise fetch the raw file.
const fetchedText = ref<string>("");
const textContent = computed(() => req.value?.content || fetchedText.value);

watch(
  () => req.value,
  async () => {
    fetchedText.value = "";
    if (previewKind.value === "text" && req.value && !req.value.content) {
      try {
        const resp = await fetch(inlineLink.value);
        fetchedText.value = await resp.text();
      } catch {
        /* ignore — user can still download */
      }
    }
  }
);

// Functions
const base64 = (name: any) => Base64.encodeURI(name);
const play = () => {
  if (tag.value) {
    audio.value?.pause();
    tag.value = false;
  } else {
    audio.value?.play();
    tag.value = true;
  }
};
const fetchData = async () => {
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  // Set loading to true and reset the error.
  layoutStore.loading = true;
  error.value = null;
  if (password.value !== "") {
    attemptedPasswordLogin.value = true;
  }

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;

  try {
    const file = await api.fetch(url, password.value);
    file.hash = hash.value;

    token.value = file.token || "";

    fileStore.updateRequest(file);
    document.title = `${file.name} - ${document.title}`;

    // Chrome handling: folders keep the standard header bar + breadcrumbs for
    // navigation; any single-file / download view gets a clean, chrome-less page.
    if (file.isDir) {
      // 文件夹：隐藏侧栏 nav，但保留分享页自己的 header-bar + breadcrumbs 以便导航
      restoreSidebar();
      hideNavOnly();
    } else {
      hideSidebarForMdPreview();
    }

    // Download share: trigger the download immediately instead of previewing.
    if ((file as any).shareType === "download" && !downloadStarted) {
      downloadStarted = true;
      hideSidebarForMdPreview();
      // defer so the DOM (manual-download fallback) is in place first
      setTimeout(() => {
        window.location.href = link.value;
      }, 100);
    }
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
      // Password prompt / error pages also render as a clean centered card.
      hideSidebarForMdPreview();
    }
  } finally {
    layoutStore.loading = false;
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    // If we're on a listing, unselect all
    // files and folders.
    if (fileStore.selectedCount > 0) {
      fileStore.selected = [];
    }
  }
};

const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
};

const isSingleFile = () =>
  fileStore.selectedCount === 1 &&
  !req.value?.items[fileStore.selected[0]].isDir;

const download = () => {
  if (!req.value) return false;

  if (isSingleFile()) {
    api.download(
      null,
      hash.value,
      token.value,
      req.value.items[fileStore.selected[0]].path
    );
    return true;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: DownloadFormat) => {
      if (req.value === null) return false;
      layoutStore.closeHovers();

      const files: string[] = [];

      for (const i of fileStore.selected) {
        files.push(req.value.items[i].path);
      }

      api.download(format, hash.value, token.value, ...files);
      return true;
    },
  });

  return true;
};

const linkSelected = () => {
  return isSingleFile() && req.value
    ? api.getDownloadURL({
        ...req.value,
        hash: hash.value,
        path: req.value.items[fileStore.selected[0]].path,
      })
    : "";
};

const copyToClipboard = (text: string) => {
  copyToClipboardWithFallback(text).then(
    () => $showSuccess(t("success.linkCopied")),
    (e) => $showError(e)
  );
};

let injectedStyle: HTMLStyleElement | null = null;

const hideSidebarForMdPreview = () => {
  if (injectedStyle) return; // idempotent
  const style = document.createElement("style");
  style.id = "share-preview-override";
  style.textContent = `
    html, body { padding: 0 !important; margin: 0 !important; height: auto !important; min-height: auto !important; }
    #app { padding: 0 !important; height: auto !important; min-height: auto !important; }
    header, nav, nav + .overlay, .breadcrumbs, .progress { display: none !important; }
    main { width: 100% !important; margin: 0 !important; padding: 0 !important; height: auto !important; min-height: auto !important; }
  `;
  document.head.appendChild(style);
  injectedStyle = style;
};

const restoreSidebar = () => {
  if (injectedStyle) {
    injectedStyle.remove();
    injectedStyle = null;
  }
};

// 只隐藏侧栏 nav（保留分享页自己的 header-bar 和 breadcrumbs），文件夹分享用
const hideNavOnly = () => {
  if (injectedStyle) return;
  const style = document.createElement("style");
  style.id = "share-nav-hide";
  style.textContent = `
    nav, nav + .overlay { display: none !important; }
    main { width: 100% !important; margin: 0 auto !important; }
  `;
  document.head.appendChild(style);
  injectedStyle = style;
};

onMounted(async () => {
  hash.value = route.params.path[0];
  window.addEventListener("keydown", keyEvent);
  // Chrome (header/breadcrumbs) is managed inside fetchData based on result.
  await fetchData();
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
  restoreSidebar();
});

// Markdown preview
marked.use(markedKatex({ output: "mathml" as const, throwOnError: false }));
marked.use({
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      const langLabel = lang ? `<span class="code-lang">${lang}</span>` : "";
      const escaped = text
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;");
      return `<div class="code-block-wrapper">
        <div class="code-block-header">${langLabel}<button class="copy-btn" data-code="${escaped}"><i class="material-icons">content_copy</i><span>Copy</span></button></div>
        <pre><code class="language-${lang || ""}">${escaped}</code></pre>
      </div>`;
    },
  },
});

const isMarkdownFile = computed(
  () =>
    req.value &&
    !req.value.isDir &&
    (req.value.name.endsWith(".md") || req.value.name.endsWith(".markdown"))
);

const isDrawioFile = computed(
  () =>
    req.value &&
    !req.value.isDir &&
    req.value.name.endsWith(".drawio")
);

const isHtmlFile = computed(
  () =>
    req.value &&
    !req.value.isDir &&
    (req.value.name.endsWith(".html") || req.value.name.endsWith(".htm") || isDrawioFile.value)
);

const isPreviewable = computed(() => isMarkdownFile.value || isHtmlFile.value);

const htmlPreviewContent = computed(() => {
  if (!req.value?.content) return "";
  if (isDrawioFile.value) {
    const xml = req.value.content;
    const config = JSON.stringify({highlight:"#0000ff",nav:true,resize:true,toolbar:"zoom layers tags lightbox",xml:xml});
    const escaped = config.replace(/&/g,"&amp;").replace(/'/g,"&#39;").replace(/</g,"&lt;");
    return `<!DOCTYPE html>
<html><head><meta charset="UTF-8">
<meta http-equiv="Content-Security-Policy" content="default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#f8f9fa;min-height:100vh;display:flex;align-items:center;justify-content:center}
.mxgraph{max-width:100%}
.geToolbar{position:fixed!important;top:8px;right:8px;z-index:999}
</style>
</head><body>
<div class="mxgraph" data-mxgraph='${escaped}'></div>
<script>
(function(){
  var s=document.createElement("script");
  s.src="https://viewer.diagrams.net/js/viewer-static.min.js";
  s.onerror=function(){
    document.body.innerHTML='<div style="padding:40px;text-align:center;color:#666"><h2>Draw.io Viewer 加载失败</h2><p>请点击下载按钮获取 .drawio 文件，用 <a href=&quot;https://app.diagrams.net&quot; target=&quot;_blank&quot;>app.diagrams.net</a> 打开</p></div>';
  };
  document.body.appendChild(s);
})();
<\/script>
</body></html>`;
  }
  return req.value.content;
});

const renderedMarkdown = computed(() => {
  if (!isMarkdownFile.value || !req.value?.content) return "";
  try {
    return DOMPurify.sanitize(marked(req.value.content) as string, {
      ADD_ATTR: ["data-code"],
    });
  } catch {
    return "";
  }
});

const htmlPreviewFrame = ref<HTMLIFrameElement>();

const resizeHtmlFrame = () => {
  const frame = htmlPreviewFrame.value;
  if (frame?.contentDocument?.body) {
    frame.style.height = frame.contentDocument.body.scrollHeight + "px";
  }
};

const snapshotHash = computed(() => {
  if (!req.value?.content || !isHtmlFile.value) return "";
  const m = req.value.content.match(/<meta\s+name="snapshot-hash"\s+content="([^"]+)"/i);
  return m ? m[1] : "";
});

const snapshotName = computed(() => {
  if (!req.value?.content || !isHtmlFile.value) return "snapshot.png";
  const m = req.value.content.match(/<meta\s+name="snapshot-name"\s+content="([^"]+)"/i);
  return m ? m[1] : "snapshot.png";
});

const downloadSnapshot = async () => {
  if (!snapshotHash.value) return;
  try {
    const resp = await fetch(`/api/public/dl/${snapshotHash.value}`);
    const blob = await resp.blob();
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = snapshotName.value;
    a.click();
    URL.revokeObjectURL(a.href);
  } catch {
    window.open(`/api/public/dl/${snapshotHash.value}`, "_blank");
  }
};

const previewFontSize = ref(16);

const increasePreviewFont = () => {
  previewFontSize.value += 1;
  const el = document.getElementById("share-preview-container");
  if (el) el.style.fontSize = previewFontSize.value + "px";
};

const decreasePreviewFont = () => {
  if (previewFontSize.value > 10) {
    previewFontSize.value -= 1;
    const el = document.getElementById("share-preview-container");
    if (el) el.style.fontSize = previewFontSize.value + "px";
  }
};

const copyAllIcon = ref("content_copy");
const copyAllLabel = ref("Copy All");

const copyAllContent = () => {
  const content = req.value?.content || "";
  copy({ text: content }).then(() => {
    copyAllIcon.value = "check";
    copyAllLabel.value = "Copied!";
    setTimeout(() => {
      copyAllIcon.value = "content_copy";
      copyAllLabel.value = "Copy All";
    }, 2000);
  });
};

onMounted(() => {
  document.addEventListener("click", (e: Event) => {
    const btn = (e.target as HTMLElement).closest(".copy-btn");
    if (!btn) return;
    const code =
      (btn as HTMLElement).dataset.code
        ?.replace(/&amp;/g, "&")
        .replace(/&lt;/g, "<")
        .replace(/&gt;/g, ">")
        .replace(/&quot;/g, '"') || "";
    copy({ text: code }).then(() => {
      const span = btn.querySelector("span");
      if (span) {
        span.textContent = "Copied!";
        setTimeout(() => (span.textContent = "Copy"), 2000);
      }
    });
  });
});
</script>

<style scoped>
#listing.list {
  height: auto;
}

#shareList {
  overflow-y: scroll;
}

@media (min-width: 930px) {
  #shareList {
    height: calc(100vh - 9.8em);
    overflow-y: auto;
  }
}

.share-md-preview {
  max-width: 92%;
  margin: 0 auto;
  padding: 0;
}

@media (max-width: 768px) {
  .share-md-preview {
    max-width: 100%;
    padding: 0 0.5em;
  }
}

.share-md-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #f6f8fa;
  border: 1px solid #d1d9e0;
  border-radius: 8px 8px 0 0;
  margin-top: 0;
  flex-wrap: wrap;
  gap: 0.5em;
}

.share-md-header h3 {
  margin: 0;
  font-size: 1em;
  font-weight: 600;
  color: #1f2328;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
}

.share-md-actions {
  display: flex;
  gap: 6px;
  align-items: center;
}

.share-md-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border: 1px solid #d1d9e0;
  border-radius: 6px;
  background: #fff;
  color: #1f2328;
  cursor: pointer;
  font-size: 13px;
  text-decoration: none;
  transition: all 0.12s ease;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.share-md-btn:hover {
  background: #f3f4f6;
  border-color: #9a9fa5;
}

.share-md-btn i {
  font-size: 16px;
}

.share-md-fontsize {
  font-size: 13px;
  color: #656d76;
  min-width: 3em;
  text-align: center;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
}

.share-md-body {
  border: 1px solid #d1d9e0;
  border-top: none;
  border-radius: 0 0 8px 8px;
  background: #fff;
  margin-bottom: 0;
}

.share-html-body {
  border: 1px solid #d1d9e0;
  border-top: none;
  border-radius: 0 0 8px 8px;
  background: #fff;
  margin-bottom: 0;
}

.share-html-iframe {
  width: 100%;
  height: 0;
  border: none;
  border-radius: 0 0 8px 8px;
}

.share-download {
  min-height: 70vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
.share-download-card {
  text-align: center;
  padding: 2.5em 2em;
}
.share-download-ic {
  font-size: 56px;
  color: #1e9e5a;
}
.share-download-card h2 {
  margin: 0.6em 0 0.2em;
  font-size: 1.3em;
  font-weight: 700;
  color: var(--lumen-accent);
}
.share-download-card p {
  color: #9a9a9e;
  font-family: ui-monospace, Menlo, monospace;
  margin-bottom: 1.4em;
  word-break: break-all;
}
.share-download-card .button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

/* ===== LumenBrowser clean white share page ===== */
.ls-page {
  --ls-ink: var(--lumen-accent);
  --ls-t2: #5f5f63;
  --ls-t3: #9a9a9e;
  --ls-line: #ececee;
  --ls-soft: #f6f6f7;
  max-width: 1040px;
  margin: 0 auto;
  padding: 22px 18px 48px;
  box-sizing: border-box;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}

/* preview header */
.ls-ph {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 18px;
  background: #fff;
  border: 1px solid var(--ls-line);
  border-radius: 14px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}
.ls-ph-name {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  font-size: 15px;
  font-weight: 650;
  color: var(--ls-ink);
}
.ls-ph-name .material-icons {
  font-size: 22px;
  color: var(--ls-t2);
  flex-shrink: 0;
}
.ls-ph-name span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ls-ph-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.ls-ph-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  border: 1px solid var(--ls-line);
  border-radius: 10px;
  background: #fff;
  color: var(--ls-ink);
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
  text-decoration: none;
  transition: 0.14s;
}
.ls-ph-btn:hover {
  border-color: #c8c8cc;
  background: var(--ls-soft);
}
.ls-ph-btn .material-icons {
  font-size: 18px;
}
.ls-ph-primary {
  background: var(--ls-ink);
  border-color: var(--ls-ink);
  color: #fff;
}
.ls-ph-primary:hover {
  background: #000;
  border-color: #000;
}

/* preview body */
.ls-pv {
  display: flex;
  justify-content: center;
}
.ls-pv-img {
  max-width: 100%;
  max-height: 80vh;
  border-radius: 14px;
  border: 1px solid var(--ls-line);
  background: #fff;
  object-fit: contain;
}
.ls-pv-video {
  max-width: 100%;
  max-height: 80vh;
  border-radius: 14px;
  background: #000;
}
.ls-pv-audio-wrap {
  width: 100%;
  max-width: 560px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 18px;
  padding: 40px 24px;
  background: #fff;
  border: 1px solid var(--ls-line);
  border-radius: 16px;
}
.ls-pv-audio-ic {
  font-size: 64px;
  color: var(--ls-t3);
}
.ls-pv-audio {
  width: 100%;
}
.ls-pv-pdf {
  width: 100%;
  height: 82vh;
  border: 1px solid var(--ls-line);
  border-radius: 14px;
  background: #fff;
}
.ls-pv-text {
  width: 100%;
  margin: 0;
  padding: 18px 20px;
  background: #fff;
  border: 1px solid var(--ls-line);
  border-radius: 14px;
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #1f2328;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-x: auto;
  box-sizing: border-box;
}

/* unpreviewable file card */
.ls-fc {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  width: 100%;
  max-width: 440px;
  padding: 44px 28px 32px;
  background: #fff;
  border: 1px solid var(--ls-line);
  border-radius: 18px;
  margin-top: 10px;
}
.ls-fc-ic {
  width: 88px;
  height: 88px;
  border-radius: 22px;
  background: var(--ls-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 14px;
}
.ls-fc-ic .material-icons {
  font-size: 46px;
  color: var(--ls-t2);
}
.ls-fc-badge {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--ls-t3);
  background: var(--ls-soft);
  border-radius: 99px;
  padding: 3px 11px;
  margin-bottom: 12px;
}
.ls-fc-name {
  font-size: 16px;
  font-weight: 650;
  color: var(--ls-ink);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ls-fc-meta {
  font-size: 13px;
  color: var(--ls-t3);
  margin-top: 6px;
  margin-bottom: 22px;
}
.ls-fc-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: center;
}
.ls-fc-dl,
.ls-fc-copy {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 42px;
  padding: 0 20px;
  border-radius: 11px;
  font-size: 14px;
  font-weight: 650;
  cursor: pointer;
  text-decoration: none;
  border: 1px solid transparent;
}
.ls-fc-dl {
  background: var(--ls-ink);
  color: #fff;
}
.ls-fc-dl:hover {
  background: #000;
}
.ls-fc-copy {
  background: #fff;
  border-color: var(--ls-line);
  color: var(--ls-ink);
}
.ls-fc-copy:hover {
  border-color: #c8c8cc;
  background: var(--ls-soft);
}
.ls-fc-dl .material-icons,
.ls-fc-copy .material-icons {
  font-size: 18px;
}

/* folder listing */
.ls-folder {
  background: #fff;
  border: 1px solid var(--ls-line);
  border-radius: 14px;
  padding: 8px;
}
.ls-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: var(--ls-t3);
  padding: 48px 0;
}

/* password page */
.ls-pw-page {
  min-height: 80vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
.ls-pw-card {
  width: 360px;
  max-width: 92vw;
  background: #fff;
  border: 1px solid #ececee;
  border-radius: 18px;
  padding: 30px 26px 26px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.06);
  text-align: center;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}
.ls-pw-ic {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: #f3f3f4;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 14px;
}
.ls-pw-ic .material-icons {
  font-size: 26px;
  color: var(--lumen-accent);
}
.ls-pw-title {
  font-size: 18px;
  font-weight: 750;
  color: var(--lumen-accent);
  margin: 0 0 6px;
}
.ls-pw-hint {
  font-size: 13px;
  color: #9a9a9e;
  margin: 0 0 18px;
}
.ls-pw-field {
  width: 100%;
  height: 44px;
  border: 1px solid #e2e2e5;
  border-radius: 11px;
  padding: 0 14px;
  font-size: 14px;
  box-sizing: border-box;
  outline: none;
  transition: 0.14s;
}
.ls-pw-field:focus {
  border-color: var(--lumen-accent);
}
.ls-pw-wrong {
  color: #e5484d;
  font-size: 12.5px;
  margin-top: 10px;
}
.ls-pw-submit {
  width: 100%;
  height: 44px;
  margin-top: 16px;
  border: 0;
  border-radius: 11px;
  background: var(--lumen-accent);
  color: #fff;
  font-size: 14px;
  font-weight: 650;
  cursor: pointer;
  transition: 0.14s;
}
.ls-pw-submit:hover {
  background: #000;
}

@media (max-width: 600px) {
  .ls-ph {
    flex-direction: column;
    align-items: stretch;
  }
  .ls-ph-actions {
    justify-content: flex-end;
  }
}
</style>
