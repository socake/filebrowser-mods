<template>
  <div>
    <template v-if="(!isPreviewable || !req?.content) && !isDownloadShare">
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
      <div v-if="error.status === 401">
        <div class="card floating" id="password" style="z-index: 9999999">
          <div v-if="attemptedPasswordLogin" class="share__wrong__password">
            {{ t("login.wrongCredentials") }}
          </div>
          <div class="card-title">
            <h2>{{ t("login.password") }}</h2>
          </div>

          <div class="card-content">
            <input
              v-focus
              class="input input--block"
              type="password"
              :placeholder="t('login.password')"
              v-model="password"
              @keyup.enter="fetchData"
            />
          </div>
          <div class="card-action">
            <button
              class="button button--flat"
              @click="fetchData"
              :aria-label="t('buttons.submit')"
              :data-title="t('buttons.submit')"
            >
              {{ t("buttons.submit") }}
            </button>
          </div>
        </div>
        <div class="overlay" />
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

      <!-- Normal share view -->
      <div v-else class="share">
        <div
          class="share__box share__box__info"
          style="
            position: -webkit-sticky;
            position: sticky;
            top: -20.6em;
            z-index: 999;
          "
        >
          <div class="share__box__header" style="height: 3em">
            {{
              req.isDir
                ? t("download.downloadFolder")
                : t("download.downloadFile")
            }}
          </div>
          <div
            v-if="!req.isDir"
            class="share__box__element share__box__center share__box__icon"
          >
            <i class="material-icons">{{ icon }}</i>
          </div>
          <div class="share__box__element" style="height: 3em">
            <strong>{{ $t("prompts.displayName") }}</strong> {{ req.name }}
          </div>
          <div v-if="!req.isDir" class="share__box__element" :title="modTime">
            <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
          </div>
          <div class="share__box__element" style="height: 3em">
            <strong>{{ $t("prompts.size") }}:</strong> {{ humanSize }}
          </div>
          <div class="share__box__element share__box__center">
            <a
              target="_blank"
              :href="link"
              class="button button--flat"
              style="height: 4em"
            >
              <div>
                <i class="material-icons">file_download</i
                >{{ t("buttons.download") }}
              </div>
            </a>
            <a
              target="_blank"
              :href="inlineLink"
              class="button button--flat"
              v-if="!req.isDir"
            >
              <div>
                <i class="material-icons">open_in_new</i
                >{{ t("buttons.openFile") }}
              </div>
            </a>
            <qrcode-vue
              v-if="req.isDir"
              :value="link"
              :size="100"
              level="M"
            ></qrcode-vue>
          </div>
          <div v-if="!req.isDir" class="share__box__element share__box__center">
            <qrcode-vue :value="link" :size="200" level="M"></qrcode-vue>
          </div>
          <div
            v-if="req.isDir"
            class="share__box__element share__box__header"
            style="height: 3em"
          >
            {{ $t("sidebar.preview") }}
          </div>
          <div
            v-if="req.isDir"
            class="share__box__element share__box__center share__box__icon"
            style="padding: 0em !important; height: 12em !important"
          >
            <a
              target="_blank"
              :href="raw"
              class="button button--flat"
              v-if="
                !fileStore.multiple &&
                fileStore.selectedCount === 1 &&
                req.items[fileStore.selected[0]].type === 'image'
              "
              style="height: 12em; padding: 0; margin: 0"
            >
              <img style="height: 12em" :src="raw" />
            </a>
            <div
              v-else-if="
                fileStore.multiple &&
                fileStore.selectedCount === 1 &&
                req.items[fileStore.selected[0]].type === 'audio'
              "
              style="height: 12em; padding-top: 1em; margin: 0"
            >
              <button
                @click="play"
                v-if="!tag"
                style="
                  font-size: 6em !important;
                  border: 0px;
                  outline: none;
                  background: white;
                "
                class="material-icons"
              >
                play_circle_filled
              </button>
              <button
                @click="play"
                v-if="tag"
                style="
                  font-size: 6em !important;
                  border: 0px;
                  outline: none;
                  background: white;
                "
                class="material-icons"
              >
                pause_circle_filled
              </button>
              <audio
                id="myaudio"
                ref="audio"
                :src="raw"
                controls
                :autoplay="tag"
              ></audio>
            </div>
            <video
              v-else-if="
                !fileStore.multiple &&
                fileStore.selectedCount === 1 &&
                req.items[fileStore.selected[0]].type === 'video'
              "
              style="height: 12em; padding: 0; margin: 0"
              :src="raw"
              controls
            >
              Sorry, your browser doesn't support embedded videos, but don't
              worry, you can <a :href="raw">download it</a>
              and watch it with your favorite video player!
            </video>
            <i
              v-else-if="
                !fileStore.multiple &&
                fileStore.selectedCount === 1 &&
                req.items[fileStore.selected[0]].isDir
              "
              class="material-icons"
              >folder
            </i>
            <i v-else class="material-icons">call_to_action</i>
          </div>
        </div>
        <div
          id="shareList"
          v-if="req.isDir && req.items.length > 0"
          class="share__box share__box__items"
        >
          <div class="share__box__header" v-if="req.isDir">
            {{ t("files.files") }}
          </div>
          <div id="listing" class="list file-icons">
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
        </div>
        <div
          v-else-if="req.isDir && req.items.length === 0"
          class="share__box share__box__items"
        >
          <h2 class="message">
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ t("files.lonely") }}</span>
          </h2>
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
import { copy } from "@/utils/clipboard";
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
  copy({ text }).then(
    () => {
      // clipboard successfully set
      $showSuccess(t("success.linkCopied"));
    },
    () => {
      // clipboard write failed
      copy({ text }, { permission: true }).then(
        () => {
          // clipboard successfully set
          $showSuccess(t("success.linkCopied"));
        },
        (e) => {
          // clipboard write failed
          $showError(e);
        }
      );
    }
  );
};

let injectedStyle: HTMLStyleElement | null = null;

const hideSidebarForMdPreview = () => {
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

onMounted(async () => {
  hash.value = route.params.path[0];
  window.addEventListener("keydown", keyEvent);
  await fetchData();
  if (isPreviewable.value && req.value?.content) {
    hideSidebarForMdPreview();
  }
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
  color: #141414;
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
</style>
