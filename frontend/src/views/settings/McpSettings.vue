<template>
  <div class="mcp-wrap">
    <div class="mcp-card">
      <div class="mcp-head">
        <div class="mcp-badge">
          <i class="material-icons">smart_toy</i>
        </div>
        <div>
          <h2 class="mcp-title">AI 助手接入（MCP）</h2>
          <p class="mcp-sub">
            LumenBrowser 内置 MCP server，让 Claude、qwen 等 AI 助手通过 Model
            Context Protocol 直接操作你的文件——浏览、搜索、读取、上传、创建分享。
          </p>
        </div>
      </div>

      <div class="mcp-server">
        <div class="mcp-server-label">当前服务地址</div>
        <div class="mcp-server-val">
          <code>{{ origin }}</code>
          <button class="mcp-copy" @click="copy(origin, 'srv')">
            {{ copied === "srv" ? "已复制" : "复制" }}
          </button>
        </div>
      </div>

      <div class="mcp-token">
        <div class="mcp-token-label">访问令牌</div>
        <p class="mcp-token-desc">
          直接生成 MCP 访问令牌，填入下方配置即可，无需运行 lumen login。
        </p>
        <div class="mcp-token-row">
          <div class="mcp-expiry">
            <button
              v-for="opt in expiryOptions"
              :key="opt.value"
              class="mcp-expiry-btn"
              :class="{ active: expiry === opt.value }"
              type="button"
              @click="expiry = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
          <button
            class="mcp-gen-btn"
            type="button"
            :disabled="generating"
            @click="generateToken"
          >
            {{ generating ? "生成中…" : "生成令牌" }}
          </button>
        </div>
        <p v-if="tokenError" class="mcp-token-err">{{ tokenError }}</p>
        <div v-if="generatedToken" class="mcp-token-result">
          <div class="mcp-code">
            <pre>{{ generatedToken }}</pre>
            <button class="mcp-copy" @click="copy(generatedToken, 'tok')">
              {{ copied === "tok" ? "已复制" : "复制" }}
            </button>
          </div>
          <div class="mcp-token-expiry-text">{{ expiresText }}</div>
        </div>
      </div>

      <h3 class="mcp-h3">可用工具</h3>
      <div class="mcp-tools">
        <div class="mcp-tool" v-for="tool in tools" :key="tool.name">
          <code>{{ tool.name }}</code>
          <span>{{ tool.desc }}</span>
        </div>
      </div>

      <h3 class="mcp-h3">配置步骤（以 Claude Desktop 为例）</h3>
      <ol class="mcp-steps">
        <li>
          <div class="mcp-step-title">下载 lumen-mcp 可执行文件</div>
          <div class="mcp-dl-row">
            <select v-model="dlTarget" class="mcp-dl-sel">
              <option v-for="t in dlTargets" :key="t.key" :value="t.key">{{ t.label }}</option>
            </select>
            <button class="mcp-gen-btn" type="button" :disabled="dling" @click="downloadMcp">
              {{ dling ? "下载中…" : "下载" }}
            </button>
          </div>
          <p v-if="dlError" class="mcp-token-err">{{ dlError }}</p>
          <div class="mcp-hint">
            已自动识别你的系统（{{ detectedLabel }}）。{{ dlHint }}
          </div>
        </li>
        <li>
          <div class="mcp-step-title">生成访问令牌</div>
          <div class="mcp-hint">
            在上方「访问令牌」选有效期、点「生成令牌」，会自动填入下方配置；命令行用户也可
            <code>{{ loginCmd }}</code>。
          </div>
        </li>
        <li>
          <div class="mcp-step-title">
            编辑 claude_desktop_config.json，加入 lumen server
          </div>
          <div class="mcp-code">
            <pre>{{ configJson }}</pre>
            <button class="mcp-copy" @click="copy(configJson, 's3')">
              {{ copied === "s3" ? "已复制" : "复制" }}
            </button>
          </div>
          <div class="mcp-hint">
            省略 env 则自动读取 ~/.lumen/config.json 里的 token；macOS 配置文件在
            <code>~/Library/Application Support/Claude/</code>，Windows 在
            <code>%APPDATA%\Claude\</code>。
          </div>
        </li>
        <li>
          <div class="mcp-step-title">重启 Claude Desktop</div>
          <div class="mcp-hint">
            之后即可对 AI 说「列出根目录」「搜索包含 report 的文件」「把本地
            a.pdf 上传到 /uploads」「给 /uploads/a.pdf 创建一个下载分享链接」等。
          </div>
        </li>
      </ol>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { fetchURL } from "@/api/utils";

const origin = window.location.origin;

type Expiry = "permanent" | "7d" | "30d" | "90d" | "1y";

const expiryOptions: { value: Expiry; label: string }[] = [
  { value: "permanent", label: "永久" },
  { value: "7d", label: "7 天" },
  { value: "30d", label: "30 天" },
  { value: "90d", label: "90 天" },
  { value: "1y", label: "1 年" },
];

const expiry = ref<Expiry>("permanent");
const generatedToken = ref("");
const generating = ref(false);
const tokenError = ref("");
const expiresText = ref("");

const generateToken = async () => {
  generating.value = true;
  tokenError.value = "";
  try {
    const res = await fetchURL("/api/token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expiry: expiry.value }),
    });
    const data = (await res.json()) as { token: string; expiresAt: number };
    generatedToken.value = data.token;
    if (!data.expiresAt) {
      expiresText.value = "永久有效";
    } else {
      const d = new Date(data.expiresAt * 1000);
      expiresText.value = `有效期至 ${d.toLocaleString()}`;
    }
  } catch (e) {
    tokenError.value =
      e instanceof Error ? e.message : "生成失败，请稍后重试";
  } finally {
    generating.value = false;
  }
};

const tools = [
  { name: "list_files", desc: "列目录，返回名称/类型/大小/修改时间" },
  { name: "search_files", desc: "按关键词搜索文件" },
  { name: "read_file", desc: "读取文件文本内容" },
  { name: "upload_file", desc: "上传本地文件到远程路径" },
  { name: "create_share", desc: "创建公开分享链接（preview/download）" },
];

const loginCmd = `lumen login ${origin} -u admin`;

// MCP 二进制下载（各平台已交叉编译嵌入服务端，用户直接下载，无需自行构建）
const dlTargets = [
  { key: "windows-amd64", label: "Windows (x64)" },
  { key: "darwin-arm64", label: "macOS (Apple 芯片 M1/M2/M3)" },
  { key: "darwin-amd64", label: "macOS (Intel)" },
  { key: "linux-amd64", label: "Linux (x64)" },
];
const detectDefault = () => {
  const ua = navigator.userAgent;
  if (/Win/i.test(ua)) return "windows-amd64";
  if (/Mac/i.test(ua)) return "darwin-arm64";
  return "linux-amd64";
};
const dlTarget = ref(detectDefault());
const detectedLabel = computed(
  () => dlTargets.find((t) => t.key === dlTarget.value)?.label ?? ""
);
const dling = ref(false);
const dlError = ref("");
const downloadMcp = async () => {
  dling.value = true;
  dlError.value = "";
  try {
    const [os, arch] = dlTarget.value.split("-");
    const res = await fetchURL(`/api/mcp/download?os=${os}&arch=${arch}`);
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = os === "windows" ? "lumen-mcp.exe" : "lumen-mcp";
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  } catch (e) {
    dlError.value = e instanceof Error ? e.message : "下载失败，请稍后重试";
  } finally {
    dling.value = false;
  }
};

// MCP 可执行文件的安装路径，按所选平台动态生成。
// Windows 用 %USERPROFILE%\lumen-mcp\lumen-mcp.exe（源码里 \\ = 单个反斜杠，
// 经 JSON.stringify 后在 pre 代码块里渲染为合法的 \\ 双反斜杠）。
const mcpCommand = computed(() =>
  dlTarget.value.startsWith("windows")
    ? "%USERPROFILE%\\lumen-mcp\\lumen-mcp.exe"
    : "/usr/local/bin/lumen-mcp"
);

// 下载步骤的「放哪 + 怎么用」提示，按平台切换。
const dlHint = computed(() =>
  dlTarget.value.startsWith("windows")
    ? "下载后放到 %USERPROFILE%\\lumen-mcp\\ 目录（无需 chmod），下方配置的 command 已指向该 lumen-mcp.exe。"
    : "下载后放到固定路径（如 /usr/local/bin/lumen-mcp），并执行 chmod +x lumen-mcp 赋予可执行权限。"
);

const configJson = computed(() =>
  JSON.stringify(
    {
      mcpServers: {
        lumen: {
          command: mcpCommand.value,
          env: {
            LUMEN_SERVER: origin,
            LUMEN_TOKEN: generatedToken.value || "<你的 token>",
          },
        },
      },
    },
    null,
    2
  )
);

const copied = ref("");
const copy = (text: string, key: string) => {
  navigator.clipboard?.writeText(text);
  copied.value = key;
  setTimeout(() => {
    if (copied.value === key) copied.value = "";
  }, 1500);
};
</script>

<style scoped>
.mcp-dl-row { display: flex; gap: 10px; margin-bottom: 10px; }
.mcp-dl-sel { flex: 1; height: 40px; border: 1.5px solid var(--lumen-line, #ececee); border-radius: 9px; padding: 0 12px; font-size: 14px; background: #fff; color: #232326; outline: none; cursor: pointer; }
.mcp-dl-sel:focus { border-color: var(--lumen-accent, #141414); }
.mcp-wrap {
  max-width: 820px;
  margin: 0 auto;
  padding: 1.5em 1em 3em;
}
.mcp-card {
  background: #fff;
  border: 1px solid var(--lumen-line, #ececee);
  border-radius: 14px;
  padding: 28px 30px;
}
.mcp-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 22px;
}
.mcp-badge {
  flex-shrink: 0;
  width: 46px;
  height: 46px;
  border-radius: 12px;
  background: var(--lumen-accent, #141414);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}
.mcp-badge i {
  font-size: 26px;
}
.mcp-title {
  font-size: 21px;
  font-weight: 700;
  margin: 2px 0 6px;
}
.mcp-sub {
  font-size: 14px;
  color: #6b6b70;
  line-height: 1.6;
  margin: 0;
}
.mcp-server {
  background: #f7f7f8;
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 24px;
}
.mcp-server-label {
  font-size: 12px;
  font-weight: 700;
  color: #9a9a9e;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 8px;
}
.mcp-server-val {
  display: flex;
  align-items: center;
  gap: 10px;
}
.mcp-server-val code {
  flex: 1;
  font-size: 14px;
  color: var(--lumen-accent, #141414);
  font-weight: 600;
}
.mcp-token {
  background: #f7f7f8;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 24px;
}
.mcp-token-label {
  font-size: 12px;
  font-weight: 700;
  color: #9a9a9e;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 6px;
}
.mcp-token-desc {
  font-size: 13px;
  color: #6b6b70;
  line-height: 1.6;
  margin: 0 0 12px;
}
.mcp-token-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}
.mcp-expiry {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 6px;
  flex: 1;
}
.mcp-expiry-btn {
  height: 30px;
  padding: 0 14px;
  border: 1px solid var(--lumen-line, #ececee);
  border-radius: 7px;
  background: #fff;
  color: #6b6b70;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.mcp-expiry-btn.active {
  background: var(--lumen-accent, #141414);
  border-color: var(--lumen-accent, #141414);
  color: #fff;
}
.mcp-gen-btn {
  height: 32px;
  padding: 0 18px;
  border: 0;
  border-radius: 7px;
  background: var(--lumen-accent, #141414);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.mcp-gen-btn:disabled {
  opacity: 0.6;
  cursor: default;
}
.mcp-token-err {
  font-size: 12.5px;
  color: #d23b3b;
  margin: 10px 0 0;
}
.mcp-token-result {
  margin-top: 12px;
}
.mcp-token-expiry-text {
  font-size: 12.5px;
  color: #6b6b70;
  margin-top: 8px;
}
.mcp-h3 {
  font-size: 15px;
  font-weight: 700;
  margin: 26px 0 12px;
}
.mcp-tools {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
.mcp-tool {
  display: flex;
  flex-direction: column;
  gap: 3px;
  background: #f7f7f8;
  border-radius: 8px;
  padding: 10px 13px;
}
.mcp-tool code {
  font-size: 13px;
  font-weight: 700;
  color: var(--lumen-accent, #141414);
}
.mcp-tool span {
  font-size: 12.5px;
  color: #6b6b70;
}
.mcp-steps {
  margin: 0;
  padding-left: 20px;
}
.mcp-steps li {
  margin-bottom: 18px;
}
.mcp-step-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}
.mcp-code {
  position: relative;
  background: #f6f6f7;
  border: 1px solid var(--lumen-line, #ececee);
  border-radius: 8px;
  padding: 12px 14px;
}
.mcp-code pre {
  margin: 0;
  font-family: ui-monospace, Menlo, Consolas, monospace;
  font-size: 12.5px;
  line-height: 1.55;
  color: #232326;
  white-space: pre-wrap;
  word-break: break-all;
}
.mcp-copy {
  position: absolute;
  top: 8px;
  right: 8px;
  height: 26px;
  padding: 0 11px;
  border: 0;
  border-radius: 6px;
  background: var(--lumen-accent, #141414);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
.mcp-server-val .mcp-copy {
  position: static;
}
.mcp-hint {
  font-size: 12.5px;
  color: #9a9a9e;
  line-height: 1.6;
  margin-top: 8px;
}
.mcp-hint code {
  font-size: 12px;
  color: #6b6b70;
}
</style>
