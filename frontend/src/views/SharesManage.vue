<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <errors v-if="error" :errorCode="error.status" />

    <main v-else class="shares-main">
      <div class="head">
        <div>
          <h1>{{ t("settings.shareManagement") }}</h1>
          <div class="sub">
            {{ t("settings.shareManagement") }} · 共 {{ links.length }} 条
            <template v-if="filtered.length !== links.length">
              · 筛选出 {{ filtered.length }} 条
            </template>
          </div>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filters">
        <label class="fl">
          <span>类型</span>
          <select v-model="filterType">
            <option value="all">全部</option>
            <option value="preview">预览</option>
            <option value="download">下载</option>
          </select>
        </label>
        <label class="fl">
          <span>文件类型</span>
          <select v-model="filterFileType">
            <option value="all">全部</option>
            <option value="image">图片</option>
            <option value="document">文档</option>
            <option value="pdf">PDF</option>
            <option value="office">Office</option>
            <option value="archive">压缩包</option>
            <option value="other">其他</option>
          </select>
        </label>
        <label class="fl">
          <span>创建时间</span>
          <select v-model="filterCreated">
            <option value="all">全部</option>
            <option value="7">近 7 天</option>
            <option value="30">近 30 天</option>
          </select>
        </label>
        <label class="fl">
          <span>过期状态</span>
          <select v-model="filterExpire">
            <option value="all">全部</option>
            <option value="permanent">永久</option>
            <option value="expired">已过期</option>
          </select>
        </label>
      </div>

      <!-- 批量操作栏 -->
      <div v-if="selected.size > 0" class="bulk">
        <span class="cnt">已选中 {{ selected.size }} 项</span>
        <span class="sp"></span>
        <!-- 批量设过期已移除：后端暂无更新分享(share)的 API -->
        <button class="b-del" @click="bulkDelete">
          <i class="material-icons">delete</i>批量删除
        </button>
      </div>

      <div class="card">
        <table v-if="filtered.length > 0">
          <thead>
            <tr>
              <th class="ck-col">
                <span
                  class="ck"
                  :class="{ on: allSelected }"
                  @click="toggleAll"
                ></span>
              </th>
              <th>文件</th>
              <th class="w-type">类型</th>
              <th class="w-time">创建时间</th>
              <th class="w-time">过期</th>
              <th class="w-pw">密码</th>
              <th class="w-act">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="link in filtered"
              :key="link.hash"
              :class="{ sel: selected.has(link.hash) }"
            >
              <td>
                <span
                  class="ck"
                  :class="{ on: selected.has(link.hash) }"
                  @click="toggleOne(link.hash)"
                ></span>
              </td>
              <td>
                <div class="fname">
                  <div class="fi" :style="{ color: badge(link.path).color }">
                    {{ badge(link.path).label }}
                  </div>
                  <div>
                    <div class="n">{{ baseName(link.path) }}</div>
                    <div class="p">
                      {{ link.path
                      }}<template v-if="isAdmin && link.username">
                        · {{ link.username }}</template
                      >
                    </div>
                  </div>
                </div>
              </td>
              <td>
                <span
                  class="tag"
                  :class="link.type === 'download' ? 'dl' : 'preview'"
                >
                  {{ link.type === "download" ? "下载" : "预览" }}
                </span>
              </td>
              <td class="muted">{{ createdLabel(link) }}</td>
              <td :class="{ muted: link.expire === 0, expired: isExpired(link) }">
                {{ expireLabel(link) }}
              </td>
              <td>
                <i
                  v-if="link.password_hash"
                  class="material-icons lock on"
                  title="有密码"
                  >lock</i
                >
                <span v-else class="lock muted">—</span>
              </td>
              <td>
                <div class="row-act">
                  <button title="复制链接" @click="copyLink(link)">
                    <i class="material-icons">content_copy</i>
                  </button>
                  <button class="del" title="删除" @click="deleteOne(link)">
                    <i class="material-icons">delete</i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty">
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{
            links.length === 0 ? t("files.lonely") : "没有符合条件的分享"
          }}</span>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { share as api, users } from "@/api";
import dayjs from "dayjs";
import Errors from "@/views/Errors.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const { t } = useI18n();

const layoutStore = useLayoutStore();
const authStore = useAuthStore();

const error = ref<StatusError | null>(null);
const links = ref<Share[]>([]);
const selected = ref<Set<string>>(new Set());

const filterType = ref("all");
const filterFileType = ref("all");
const filterCreated = ref("all");
const filterExpire = ref("all");

const isAdmin = computed(() => !!authStore.user?.perm.admin);

onMounted(async () => {
  layoutStore.loading = true;
  try {
    const newLinks = await api.list();
    if (authStore.user?.perm.admin) {
      const userMap = new Map<number, string>();
      for (const user of await users.getAll())
        userMap.set(user.id, user.username);
      for (const link of newLinks) {
        if (link.userID && userMap.has(link.userID))
          link.username = userMap.get(link.userID);
      }
    }
    links.value = newLinks;
  } catch (err) {
    if (err instanceof Error) error.value = err as StatusError;
  } finally {
    layoutStore.loading = false;
  }
});

// ---- 文件类型判定 ----
const EXT_MAP: Record<string, string> = {
  png: "image",
  jpg: "image",
  jpeg: "image",
  gif: "image",
  webp: "image",
  bmp: "image",
  svg: "image",
  ico: "image",
  tiff: "image",
  heic: "image",
  pdf: "pdf",
  doc: "office",
  docx: "office",
  xls: "office",
  xlsx: "office",
  ppt: "office",
  pptx: "office",
  txt: "document",
  md: "document",
  rtf: "document",
  odt: "document",
  csv: "document",
  zip: "archive",
  rar: "archive",
  "7z": "archive",
  tar: "archive",
  gz: "archive",
  bz2: "archive",
  xz: "archive",
};

const extOf = (path: string): string => {
  const name = baseName(path);
  const idx = name.lastIndexOf(".");
  if (idx < 0 || idx === name.length - 1) return "";
  return name.slice(idx + 1).toLowerCase();
};

const categoryOf = (path: string): string => {
  const ext = extOf(path);
  return EXT_MAP[ext] || "other";
};

const baseName = (path: string): string => {
  const clean = path.replace(/\/+$/, "");
  const parts = clean.split("/");
  return parts[parts.length - 1] || path;
};

const badge = (path: string): { label: string; color: string } => {
  const ext = extOf(path);
  const cat = categoryOf(path);
  if (cat === "image") return { label: "IMG", color: "#5f5f63" };
  if (cat === "pdf") return { label: "PDF", color: "#e5484d" };
  if (cat === "archive") return { label: "ZIP", color: "#f59e0b" };
  if (cat === "office") {
    if (ext.startsWith("xls")) return { label: "XLS", color: "#1e9e5a" };
    if (ext.startsWith("ppt")) return { label: "PPT", color: "#f59e0b" };
    return { label: "DOC", color: "#2b7cd3" };
  }
  if (cat === "document") return { label: "TXT", color: "#5f5f63" };
  return {
    label: (ext || "?").slice(0, 3).toUpperCase(),
    color: "#5f5f63",
  };
};

// ---- 过期/时间 ----
const isExpired = (link: Share): boolean =>
  !!link.expire && link.expire !== 0 && link.expire * 1000 < Date.now();

const expireLabel = (link: Share): string => {
  if (!link.expire || link.expire === 0) return "永久";
  if (isExpired(link)) return "已过期";
  return dayjs(link.expire * 1000).fromNow();
};

const createdLabel = (link: Share): string => {
  if (!link.createdAt) return "未知";
  return dayjs(link.createdAt * 1000).format("YYYY-MM-DD HH:mm");
};

// ---- 筛选 ----
const filtered = computed(() => {
  const now = Date.now();
  return links.value.filter((link) => {
    if (filterType.value !== "all") {
      const ty = link.type === "download" ? "download" : "preview";
      if (ty !== filterType.value) return false;
    }
    if (filterFileType.value !== "all") {
      if (categoryOf(link.path) !== filterFileType.value) return false;
    }
    if (filterCreated.value !== "all") {
      if (!link.createdAt) return false;
      const days = parseInt(filterCreated.value, 10);
      if (link.createdAt * 1000 < now - days * 86400000) return false;
    }
    if (filterExpire.value === "permanent") {
      if (link.expire && link.expire !== 0) return false;
    } else if (filterExpire.value === "expired") {
      if (!isExpired(link)) return false;
    }
    return true;
  });
});

// ---- 多选 ----
const allSelected = computed(
  () =>
    filtered.value.length > 0 &&
    filtered.value.every((l) => selected.value.has(l.hash))
);

const toggleOne = (hash: string) => {
  const next = new Set(selected.value);
  if (next.has(hash)) next.delete(hash);
  else next.add(hash);
  selected.value = next;
};

const toggleAll = () => {
  if (allSelected.value) {
    selected.value = new Set();
  } else {
    selected.value = new Set(filtered.value.map((l) => l.hash));
  }
};

// ---- 操作 ----
const buildLink = (share: Share) => api.getShareURL(share);

const copyLink = (link: Share) => {
  const text = buildLink(link);
  copy({ text }).then(
    () => $showSuccess(t("success.linkCopied")),
    () =>
      copy({ text }, { permission: true }).then(
        () => $showSuccess(t("success.linkCopied")),
        (e) => $showError(e)
      )
  );
};

const removeLink = async (hash: string) => {
  await api.remove(hash);
  links.value = links.value.filter((l) => l.hash !== hash);
  const next = new Set(selected.value);
  next.delete(hash);
  selected.value = next;
};

const deleteOne = (link: Share) => {
  layoutStore.showHover({
    prompt: "share-delete",
    confirm: async () => {
      layoutStore.closeHovers();
      try {
        await removeLink(link.hash);
        $showSuccess(t("settings.shareDeleted"));
      } catch (err) {
        if (err instanceof Error) $showError(err);
      }
    },
  });
};

const bulkDelete = () => {
  if (selected.value.size === 0) return;
  layoutStore.showHover({
    prompt: "share-delete",
    confirm: async () => {
      layoutStore.closeHovers();
      const hashes = [...selected.value];
      try {
        for (const hash of hashes) await removeLink(hash);
        $showSuccess(t("settings.shareDeleted"));
      } catch (err) {
        if (err instanceof Error) $showError(err);
      }
    },
  });
};

</script>

<style scoped>
.shares-main {
  --ink: var(--lumen-accent);
  --t2: #5f5f63;
  --t3: #9a9a9e;
  --line: #ececee;
  --line2: #e2e2e5;
  --soft: #f3f3f4;
  --blue: #2f7bf6;
  --green: #1e9e5a;
  --red: #e5484d;
  max-width: 1120px;
  margin: 0 auto;
  padding: 28px 34px;
  color: var(--ink);
}

.head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 6px;
}
.head h1 {
  font-size: 25px;
  font-weight: 780;
  letter-spacing: -0.02em;
  margin: 0;
}
.sub {
  font-size: 13.5px;
  color: var(--t3);
  margin-bottom: 22px;
}

/* 筛选栏 */
.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-bottom: 18px;
}
.fl {
  display: flex;
  flex-direction: column;
  gap: 5px;
  font-size: 12px;
  font-weight: 600;
  color: var(--t2);
}
.fl select {
  appearance: none;
  height: 36px;
  min-width: 130px;
  padding: 0 12px;
  border: 1px solid var(--line2);
  border-radius: 9px;
  background: #fff;
  color: var(--ink);
  font-size: 13.5px;
  font-weight: 500;
  cursor: pointer;
  outline: none;
}
.fl select:focus {
  border-color: var(--ink);
}

/* 批量操作栏 */
.bulk {
  display: flex;
  align-items: center;
  gap: 12px;
  background: var(--ink);
  color: #fff;
  border-radius: 12px;
  padding: 11px 16px;
  margin-bottom: 16px;
}
.bulk .cnt {
  font-size: 14px;
  font-weight: 600;
}
.bulk .sp {
  flex: 1;
}
.bulk button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 13px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 0;
}
.bulk button .material-icons {
  font-size: 16px;
}
.bulk .b-ghost {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}
.bulk .b-ghost:hover {
  background: rgba(255, 255, 255, 0.22);
}
.bulk .b-del {
  background: var(--red);
  color: #fff;
}

/* 表格 */
.card {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 14px;
  overflow: hidden;
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
thead {
  background: #fafafa;
}
th {
  text-align: left;
  font-size: 12px;
  font-weight: 700;
  color: var(--t3);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 13px 16px;
}
td {
  padding: 14px 16px;
  border-top: 1px solid var(--line);
  color: var(--ink);
  vertical-align: middle;
}
tr.sel td {
  background: #f6f9ff;
}
.ck-col {
  width: 40px;
}
.w-type {
  width: 100px;
}
.w-time {
  width: 150px;
}
.w-pw {
  width: 70px;
}
.w-act {
  width: 110px;
}

.ck {
  width: 17px;
  height: 17px;
  border: 1.5px solid var(--line2);
  border-radius: 5px;
  cursor: pointer;
  display: inline-block;
  vertical-align: middle;
}
.ck.on {
  background: var(--ink);
  border-color: var(--ink);
  position: relative;
}
.ck.on::after {
  content: "";
  position: absolute;
  left: 5px;
  top: 2px;
  width: 4px;
  height: 8px;
  border: solid #fff;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.fname {
  display: flex;
  align-items: center;
  gap: 10px;
}
.fi {
  width: 30px;
  height: 30px;
  border-radius: 7px;
  background: var(--soft);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: var(--t2);
  flex-shrink: 0;
}
.fname .n {
  font-weight: 600;
}
.fname .p {
  font-size: 12px;
  color: var(--t3);
  word-break: break-all;
}

.tag {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 99px;
}
.tag.preview {
  background: rgba(47, 123, 246, 0.12);
  color: #1f5fd0;
}
.tag.dl {
  background: rgba(30, 158, 90, 0.14);
  color: #15834a;
}

.muted {
  color: var(--t3);
}
.expired {
  color: var(--red);
  font-weight: 600;
}
.lock {
  color: var(--t3);
  font-size: 17px;
}
.lock.on {
  color: var(--ink);
}

.row-act {
  display: flex;
  gap: 4px;
}
.row-act button {
  width: 30px;
  height: 30px;
  border: 0;
  background: transparent;
  border-radius: 7px;
  color: var(--t3);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.row-act button .material-icons {
  font-size: 17px;
}
.row-act button:hover {
  background: var(--soft);
  color: var(--ink);
}
.row-act .del:hover {
  background: rgba(229, 72, 77, 0.1);
  color: var(--red);
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 56px 24px;
  color: var(--t3);
}
.empty .material-icons {
  font-size: 40px;
}
</style>
