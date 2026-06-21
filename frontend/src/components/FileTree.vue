<template>
  <aside class="lumen-tree" v-show="layoutStore.showFileTree">
    <div class="lumen-tree-head">
      <span class="tt">{{ t("files.directoryTree") }}</span>
      <button
        class="x"
        :aria-label="t('buttons.close')"
        :title="t('buttons.close')"
        @click="layoutStore.toggleFileTree()"
      >
        <i class="material-icons">close</i>
      </button>
    </div>

    <div class="lumen-tree-body">
      <div
        v-for="row in visibleRows"
        :key="row.node.url"
        class="ti"
        :class="{
          open: row.node.isDir && expanded.has(row.node.url),
          cur: row.node.url === currentUrl,
        }"
        :style="{ paddingLeft: 8 + row.depth * 18 + 'px' }"
        @click="onRowClick(row.node)"
      >
        <i
          v-if="row.node.isDir"
          class="material-icons chev"
          @click.stop="toggle(row.node)"
          >chevron_right</i
        >
        <span v-else class="leaf"></span>

        <i v-if="row.node.isDir" class="material-icons ic fold">folder</i>
        <i v-else class="material-icons ic file">insert_drive_file</i>

        <span class="nm">{{ row.node.name }}</span>

        <span
          v-if="row.node.isDir && childCount(row.node.url) !== null"
          class="cnt"
          >{{ childCount(row.node.url) }}</span
        >
        <span
          v-else-if="row.node.isDir && loading.has(row.node.url)"
          class="cnt spin"
          >…</span
        >
      </div>
    </div>

    <div class="lumen-tree-foot">{{ t("files.lazyHint") }}</div>
  </aside>
</template>

<script setup lang="ts">
import { computed, reactive, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { files as api } from "@/api";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";

const ROOT = "/files/";

interface TreeNode {
  name: string;
  url: string;
  isDir: boolean;
}

const { t } = useI18n();
const router = useRouter();
const layoutStore = useLayoutStore();
const fileStore = useFileStore();

// folder url -> loaded children (lazy)
const childrenMap = reactive<Record<string, TreeNode[]>>({});
// folder url -> item count, prefetched so collapsed folders still show a badge
const countMap = reactive<Record<string, number>>({});
// currently expanded folder urls
const expanded = reactive(new Set<string>());
// folder urls with an in-flight fetch
const loading = reactive(new Set<string>());
// nodes the user has manually toggled; auto logic never overrides these
const userTouched = new Set<string>();

const rootNode: TreeNode = {
  name: t("sidebar.myFiles"),
  url: ROOT,
  isDir: true,
};

// ---- url helpers (segments are encodeURIComponent'd, so '/' split is safe) ----
function ancestorUrls(dirUrl: string): string[] {
  const res = [ROOT];
  if (!dirUrl || dirUrl === ROOT) return res;
  const rest = dirUrl.slice(ROOT.length).replace(/\/$/, "");
  if (!rest) return res;
  let acc = ROOT;
  for (const p of rest.split("/")) {
    acc += p + "/";
    res.push(acc);
  }
  return res;
}

function parentDir(url: string): string {
  // /files/a/b/file.txt -> /files/a/b/   ;  /files/a/b/ -> /files/a/
  const trimmed = url.replace(/\/$/, "");
  const idx = trimmed.lastIndexOf("/");
  if (idx <= ROOT.length - 1) return ROOT;
  return trimmed.slice(0, idx + 1);
}

const currentUrl = computed(() => fileStore.req?.url ?? "");

function childCount(url: string): number | null {
  const kids = childrenMap[url];
  if (kids) return kids.length;
  return countMap[url] ?? null;
}

// 只取项目数、不展开：让折叠的文件夹也能直接显示数字（深度 1 预取）
async function loadCountOnly(url: string) {
  if (countMap[url] !== undefined || childrenMap[url]) return;
  try {
    const res = await api.fetch(url);
    countMap[url] = (res.items ?? []).length;
  } catch {
    countMap[url] = 0;
  }
}

async function loadChildren(url: string) {
  if (childrenMap[url] || loading.has(url)) return;
  loading.add(url);
  try {
    const res = await api.fetch(url);
    const items = (res.items ?? []).map((it) => ({
      name: it.name,
      url: it.url,
      isDir: it.isDir,
    }));
    // dirs first, then files; alphabetical within group
    items.sort((a, b) => {
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
    childrenMap[url] = items;
    countMap[url] = items.length;
    // 预取这一层每个子文件夹的项目数，使其折叠时也显示 badge（不再深挖）
    for (const it of items) {
      if (it.isDir) loadCountOnly(it.url);
    }
  } catch {
    childrenMap[url] = [];
  } finally {
    loading.delete(url);
  }
}

// ---- flatten visible tree ----
interface Row {
  node: TreeNode;
  depth: number;
}

function flatten(url: string, depth: number, out: Row[]) {
  const kids = childrenMap[url];
  if (!kids) return;
  for (const node of kids) {
    out.push({ node, depth });
    if (node.isDir && expanded.has(node.url)) {
      flatten(node.url, depth + 1, out);
    }
  }
}

const visibleRows = computed<Row[]>(() => {
  const rows: Row[] = [{ node: rootNode, depth: 0 }];
  if (expanded.has(ROOT)) flatten(ROOT, 1, rows);
  return rows;
});

// ---- interactions ----
function toggle(node: TreeNode) {
  if (!node.isDir) return;
  userTouched.add(node.url);
  if (expanded.has(node.url)) {
    expanded.delete(node.url);
  } else {
    expanded.add(node.url);
    loadChildren(node.url);
  }
}

function onRowClick(node: TreeNode) {
  // folders navigate into the dir; files reuse the open/preview behaviour
  router.push({ path: node.url });
}

// ---- auto-expand on navigation (only for nodes not in userTouched) ----
function autoExpand() {
  const req = fileStore.req;
  let baseDir: string;
  if (req?.isDir && req.url) {
    baseDir = req.url; // expand into current dir -> shows its direct children
  } else if (currentUrl.value) {
    baseDir = parentDir(currentUrl.value); // a file: expand its parent
  } else {
    baseDir = ROOT;
  }

  for (const u of ancestorUrls(baseDir)) {
    if (userTouched.has(u)) continue; // honour manual expand/collapse
    expanded.add(u);
    loadChildren(u);
  }
}

onMounted(() => {
  expanded.add(ROOT);
  loadChildren(ROOT);
  autoExpand();
});

watch(
  () => fileStore.req?.url,
  () => autoExpand()
);
</script>

<style scoped>
.lumen-tree {
  position: fixed;
  top: 4em;
  right: 0;
  bottom: 0;
  width: 312px;
  background: #fff;
  border-left: 1px solid #ececee;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 900;
}

.lumen-tree-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 16px 12px;
  border-bottom: 1px solid #ececee;
}
.lumen-tree-head .tt {
  font-size: 13px;
  font-weight: 700;
  color: #5f5f63;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}
.lumen-tree-head .x {
  width: 26px;
  height: 26px;
  border: 0;
  background: transparent;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9a9a9e;
  cursor: pointer;
  padding: 0;
}
.lumen-tree-head .x:hover {
  background: #f3f3f4;
  color: #141414;
}
.lumen-tree-head .x .material-icons {
  font-size: 17px;
}

.lumen-tree-body {
  flex: 1;
  overflow: auto;
  padding: 8px;
}

.ti {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 8px;
  border-radius: 8px;
  font-size: 13.5px;
  color: #5f5f63;
  cursor: pointer;
  white-space: nowrap;
}
.ti:hover {
  background: rgba(20, 20, 20, 0.05);
}
.ti.cur {
  background: rgba(47, 123, 246, 0.1);
  color: #1f5fd0;
  font-weight: 600;
}
.ti .chev {
  font-size: 18px;
  flex-shrink: 0;
  color: #9a9a9e;
  transition: transform 0.15s ease;
  border-radius: 4px;
}
.ti .chev:hover {
  background: rgba(20, 20, 20, 0.08);
  color: #141414;
}
.ti.open .chev {
  transform: rotate(90deg);
}
.ti .leaf {
  width: 18px;
  flex-shrink: 0;
}
.ti .ic {
  font-size: 18px;
  flex-shrink: 0;
}
.ti .ic.fold {
  color: #2f7bf6;
}
.ti .ic.file {
  color: #9a9a9e;
}
.ti .nm {
  overflow: hidden;
  text-overflow: ellipsis;
}
.ti .cnt {
  margin-left: auto;
  font-size: 11px;
  color: #9a9a9e;
  background: #f3f3f4;
  border-radius: 99px;
  padding: 1px 7px;
  font-weight: 600;
}
.ti.cur .cnt {
  background: rgba(47, 123, 246, 0.16);
  color: #1f5fd0;
}

.lumen-tree-foot {
  padding: 10px 14px;
  border-top: 1px solid #ececee;
  font-size: 11px;
  color: #9a9a9e;
}

@media (max-width: 900px) {
  .lumen-tree {
    width: 84vw;
    max-width: 312px;
    box-shadow: -10px 0 28px -16px rgba(20, 20, 20, 0.3);
  }
}
</style>
