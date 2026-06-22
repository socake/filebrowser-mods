<template>
  <div class="lb-about">
    <!-- 顶部导航 -->
    <header class="ab-nav">
      <div class="ab-brand">
        <span class="ab-logo"><i></i></span>
        <b>LumenBrowser</b>
      </div>
      <router-link to="/login" class="ab-nav-login">登录 →</router-link>
    </header>

    <!-- Hero -->
    <section class="ab-hero">
      <div class="ab-hero-text">
        <div class="ab-chip"><span class="ab-dot"></span>文件管理 · AI 接入 · 私有部署</div>
        <h1>把你的文件<br />带到光下。</h1>
        <p class="ab-lede">
          不只是文件管理——更接入 <b>AI 助手</b>，让它用自然语言帮你浏览、整理、分享。一个安静、克制的<b>私有或云端</b>文件工作台。
        </p>
        <div class="ab-cta">
          <router-link to="/login" class="ab-btn">开始使用 →</router-link>
          <a href="#features" class="ab-btn-ghost">了解功能 ↓</a>
        </div>
      </div>
      <div class="ab-hero-card">
        <div class="ab-window">
          <div class="ab-win-bar"><span></span><span></span><span></span><em>lumen · /files</em></div>
          <div class="ab-win-body">
            <div class="ab-row" v-for="f in heroFiles" :key="f.n">
              <span class="ab-ico" :style="{ background: f.c }">{{ f.t }}</span>
              <span class="ab-fn">{{ f.n }}</span>
              <span class="ab-fm">{{ f.s }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 数据条 -->
    <section class="ab-stats">
      <div class="ab-stat" v-for="s in stats" :key="s.l">
        <div class="ab-stat-n">{{ s.n }}</div>
        <div class="ab-stat-l">{{ s.l }}</div>
      </div>
    </section>

    <!-- 功能详述（交替图文） -->
    <section class="ab-features" id="features">
      <div class="ab-sec-head">
        <h2>一个入口，搞定文件的一切</h2>
        <p>浏览、预览、AI 操作、权限、分享，全部在同一处优雅完成。</p>
      </div>

      <div
        class="ab-block"
        v-for="(blk, i) in blocks"
        :key="blk.t"
        :class="{ reverse: i % 2 === 1 }"
      >
        <div class="ab-block-text">
          <span class="ab-block-tag" v-if="blk.tag">{{ blk.tag }}</span>
          <h3 v-html="blk.t"></h3>
          <p>{{ blk.d }}</p>
          <ul class="ab-points">
            <li v-for="pt in blk.points" :key="pt">{{ pt }}</li>
          </ul>
        </div>
        <div class="ab-block-visual">
          <div class="ab-vcard">
            <span class="ab-vico" v-html="blk.icon"></span>
            <div class="ab-vrows">
              <div class="ab-vrow" v-for="(vr, k) in blk.visual" :key="k" :style="{ width: vr + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 适用场景 -->
    <section class="ab-scenes">
      <div class="ab-sec-head">
        <h2>适合每一种使用方式</h2>
        <p>个人网盘、团队协作、开发自动化，都能胜任。</p>
      </div>
      <div class="ab-scene-grid">
        <div class="ab-scene" v-for="sc in scenes" :key="sc.t">
          <div class="ab-scene-t">{{ sc.t }}</div>
          <div class="ab-scene-d">{{ sc.d }}</div>
        </div>
      </div>
    </section>

    <!-- 底部 CTA -->
    <section class="ab-bottom">
      <h2>现在就把文件带到光下</h2>
      <p>自托管部署，数据归你掌控，几分钟即可上手。</p>
      <router-link to="/login" class="ab-btn ab-btn-lg">进入 LumenBrowser →</router-link>
    </section>

    <footer class="ab-foot">
      <span>LumenBrowser · Bring files to light</span>
      <span>Powered by <a href="https://github.com/socake" target="_blank" rel="noopener" style="color:var(--lumen-accent);text-decoration:none">socake</a> · 基于 filebrowser · Apache-2.0</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
const heroFiles = [
  { t: "DOC", n: "季度报告.docx", s: "在线预览", c: "#2b7cd3" },
  { t: "XLS", n: "财务表.xlsx", s: "下载分享", c: "#1e9e5a" },
  { t: "IMG", n: "风景.png", s: "缩略图", c: "#8a8a8e" },
  { t: "MD", n: "笔记.md", s: "即点即看", c: "#5b5b60" },
];

const stats = [
  { n: "16+", l: "文件类型图标" },
  { n: "全平台", l: "在线预览" },
  { n: "MCP", l: "AI 助手接入" },
  { n: "RBAC", l: "角色权限" },
  { n: "自托管", l: "数据归你" },
];

const blocks = [
  {
    tag: "PREVIEW",
    t: "多文件类型，<br>即点即看",
    d: "图片、视频、音频、PDF、Office、Markdown、代码……常见类型全部支持在线预览，无需下载到本地再打开。",
    points: [
      "Office 文档（Word / Excel / PPT）网页内直接渲染",
      "图片缩略图、视频/音频内置播放器",
      "Markdown 实时渲染、代码语法高亮",
      "16+ 精致文件类型图标，一眼识别",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg>',
    visual: [90, 70, 80, 55],
  },
  {
    tag: "AI · MCP",
    t: "接入 AI 助手，<br>一句话管理文件",
    d: "内置 MCP server，接入 Claude、qwen 等 AI 助手，用自然语言让 AI 直接浏览、搜索、读取、上传、分享你的文件。",
    points: [
      "5 个工具：列目录 / 搜索 / 读取 / 上传 / 创建分享",
      "「把这些报告打包分享给我」——一句话搞定",
      "网页一键生成访问令牌，免命令行配置",
      "各平台二进制可直接下载，免自行构建",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><rect x="4" y="6" width="16" height="13" rx="3"/><path d="M9 3v3M15 3v3M9 13h.01M15 13h.01M9 16h6"/></svg>',
    visual: [80, 95, 60, 75],
  },
  {
    tag: "RBAC",
    t: "基于角色的<br>权限管理",
    d: "用角色统一管理权限，而不是给每个人单独配置。给用户分配角色即套用权限，精细控制谁能看、谁能改。",
    points: [
      "4 个预设角色：管理员 / 编辑者 / 查看者 / 访客",
      "权限矩阵可视化配置，自定义角色",
      "新用户默认角色，注册即套用",
      "权限直接作用到界面：没权限的操作自动隐藏",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path d="M12 2l8 4v6c0 5-3.5 8-8 10-4.5-2-8-5-8-10V6z"/><path d="M9 12l2 2 4-4"/></svg>',
    visual: [70, 85, 90, 50],
  },
  {
    tag: "BROWSE",
    t: "目录树预览，<br>结构一目了然",
    d: "右侧目录树面板，整个文件结构尽收眼底，懒加载子目录、当前路径自动展开，大目录也能快速跳转。",
    points: [
      "懒加载子目录，海量文件不卡顿",
      "当前路径自动展开、高亮定位",
      "文件夹项目数 badge，折叠也能看",
      "一键开关，默认不打扰",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path d="M3 5h7l2 2h9v12H3z"/><path d="M8 12h8M8 15h5"/></svg>',
    visual: [60, 80, 70, 85],
  },
  {
    tag: "SHARE",
    t: "两种分享，<br>私有云公有云皆可",
    d: "预览分享在线看、下载分享直接下；可设密码、有效期，配合分享管理页统一掌控所有对外链接。",
    points: [
      "预览分享 / 下载分享，按需选择",
      "密码保护、二维码、有效期控制",
      "分享管理：多选、按类型/时间筛选、批量删除",
      "私有云、公有云部署都能对外分享",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4"/></svg>',
    visual: [85, 65, 75, 90],
  },
  {
    tag: "CLI · SELF-HOSTED",
    t: "命令行与自托管，<br>数据始终归你",
    d: "lumen 命令行客户端脚本化操作文件，配合 MCP 让自动化与 AI 无缝衔接；装在自己的服务器或云主机，不依赖第三方网盘。",
    points: [
      "lumen CLI：登录 / 上传 / 下载 / 搜索 / 分享等",
      "脚本化批处理，融入你的工作流",
      "私有服务器或云主机，一份二进制即可部署",
      "数据完全掌控，不经过第三方",
    ],
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7"><path d="M7 8l-4 4 4 4M17 8l4 4-4 4M14 4l-4 16"/></svg>',
    visual: [75, 90, 55, 70],
  },
];

const scenes = [
  { t: "个人网盘", d: "把家里或 VPS 当私有网盘，随时随地在线看图、读文档，不再受公有网盘容量与限速摆布。" },
  { t: "团队协作", d: "按角色分权，成员各司其职；对外用分享链接交付文件，密码与有效期可控。" },
  { t: "开发自动化", d: "lumen CLI + MCP，让脚本和 AI 直接读写文件，把文件操作接进你的自动化流水线。" },
];
</script>

<style scoped>
.lb-about {
  position: fixed;
  inset: 0;
  overflow: auto;
  background: #fcfcfd;
  color: var(--lumen-accent);
  font-family: "Inter", -apple-system, "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
  -webkit-font-smoothing: antialiased;
}
.lb-about::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(1100px circle at 80% 6%, rgba(20, 20, 20, 0.05), transparent 52%),
    radial-gradient(800px circle at 8% 60%, rgba(20, 20, 20, 0.03), transparent 50%);
}
.lb-about > * { position: relative; z-index: 1; }

.ab-nav { max-width: 1180px; margin: 0 auto; padding: 26px 40px; display: flex; align-items: center; justify-content: space-between; }
.ab-brand { display: flex; align-items: center; gap: 11px; }
.ab-logo { width: 38px; height: 38px; border-radius: 10px; background: var(--lumen-accent); display: flex; align-items: center; justify-content: center; }
.ab-logo i { width: 14px; height: 14px; border: 2px solid #fff; border-radius: 50%; position: relative; }
.ab-logo i::after { content: ""; position: absolute; inset: 3px; border: 1.6px solid #fff; border-radius: 50%; }
.ab-brand b { font-size: 19px; font-weight: 800; }
.ab-nav-login { font-size: 14px; font-weight: 700; color: #fff; text-decoration: none; padding: 9px 20px; background: var(--lumen-accent); border-radius: 10px; transition: 0.15s; }
.ab-nav-login:hover { transform: translateY(-1px); box-shadow: 0 10px 22px -10px rgba(20, 20, 20, 0.45); }

.ab-hero { max-width: 1180px; margin: 0 auto; padding: 40px 40px 56px; display: grid; grid-template-columns: 1.05fr 0.95fr; gap: 48px; align-items: center; }
.ab-chip { display: inline-flex; align-items: center; gap: 9px; background: #fff; border: 1px solid #ededee; border-radius: 99px; padding: 8px 16px; font-size: 13px; font-weight: 700; margin-bottom: 26px; box-shadow: 0 6px 18px -12px rgba(20, 20, 20, 0.2); }
.ab-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--lumen-accent); }
.ab-hero h1 { font-size: 74px; line-height: 0.99; font-weight: 900; letter-spacing: -0.025em; margin: 0 0 24px; }
.ab-lede { font-size: 17px; line-height: 1.75; color: #5b5b60; max-width: 460px; margin-bottom: 32px; }
.ab-lede b { color: var(--lumen-accent); font-weight: 700; }
.ab-cta { display: flex; gap: 13px; align-items: center; }
.ab-btn { display: inline-block; background: var(--lumen-accent); color: #fff; font-size: 15px; font-weight: 700; padding: 13px 24px; border-radius: 12px; text-decoration: none; transition: 0.15s; }
.ab-btn:hover { transform: translateY(-1px); box-shadow: 0 14px 30px -12px rgba(20, 20, 20, 0.4); }
.ab-btn-ghost { font-size: 15px; font-weight: 700; color: #5b5b60; text-decoration: none; padding: 13px 18px; }
.ab-btn-ghost:hover { color: var(--lumen-accent); }

.ab-hero-card { display: flex; justify-content: center; perspective: 1400px; }
.ab-window { width: 100%; max-width: 440px; background: #fff; border: 1px solid #ededee; border-radius: 16px; overflow: hidden; box-shadow: 0 50px 90px -40px rgba(20, 20, 20, 0.4), 0 16px 34px -22px rgba(20, 20, 20, 0.2); transform: rotateY(-4deg) rotateX(1.5deg); }
.ab-win-bar { display: flex; align-items: center; gap: 6px; padding: 12px 15px; border-bottom: 1px solid #f0f0f1; }
.ab-win-bar span { width: 10px; height: 10px; border-radius: 50%; background: #e2e2e5; }
.ab-win-bar em { margin-left: 12px; font-style: normal; font-size: 12px; color: #b6b6ba; font-family: ui-monospace, monospace; }
.ab-win-body { padding: 10px 8px; }
.ab-row { display: flex; align-items: center; gap: 12px; padding: 12px 14px; border-radius: 10px; }
.ab-row:hover { background: #f7f7f8; }
.ab-ico { width: 34px; height: 34px; border-radius: 8px; color: #fff; font-size: 10px; font-weight: 800; display: flex; align-items: center; justify-content: center; }
.ab-fn { font-size: 14px; font-weight: 600; flex: 1; }
.ab-fm { font-size: 12px; color: #9a9a9e; }

/* 数据条 */
.ab-stats { max-width: 1180px; margin: 0 auto; padding: 0 40px 20px; display: flex; flex-wrap: wrap; gap: 14px; }
.ab-stat { flex: 1; min-width: 150px; background: #fff; border: 1px solid #ededee; border-radius: 14px; padding: 20px 22px; box-shadow: 0 10px 26px -22px rgba(20, 20, 20, 0.25); }
.ab-stat-n { font-size: 26px; font-weight: 850; letter-spacing: -0.02em; }
.ab-stat-l { font-size: 13px; color: #9a9a9e; margin-top: 4px; font-weight: 600; }

/* 功能详述 */
.ab-features { max-width: 1100px; margin: 0 auto; padding: 50px 40px 30px; }
.ab-sec-head { text-align: center; margin-bottom: 44px; }
.ab-sec-head h2 { font-size: 38px; font-weight: 850; letter-spacing: -0.02em; margin: 0 0 12px; }
.ab-sec-head p { font-size: 16px; color: #9a9a9e; }
.ab-block { display: grid; grid-template-columns: 1fr 1fr; gap: 50px; align-items: center; margin-bottom: 64px; }
.ab-block.reverse .ab-block-text { order: 2; }
.ab-block-tag { display: inline-block; font-size: 11px; font-weight: 800; letter-spacing: 0.1em; color: #9a9a9e; margin-bottom: 12px; }
.ab-block-text h3 { font-size: 30px; font-weight: 850; line-height: 1.15; letter-spacing: -0.02em; margin: 0 0 14px; }
.ab-block-text p { font-size: 15.5px; color: #5b5b60; line-height: 1.7; margin: 0 0 18px; }
.ab-points { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 10px; }
.ab-points li { position: relative; padding-left: 26px; font-size: 14.5px; color: #232326; line-height: 1.5; }
.ab-points li::before { content: ""; position: absolute; left: 0; top: 6px; width: 16px; height: 16px; border-radius: 50%; background: var(--lumen-accent); }
.ab-points li::after { content: ""; position: absolute; left: 5px; top: 10px; width: 5px; height: 8px; border: solid #fff; border-width: 0 2px 2px 0; transform: rotate(45deg); }
.ab-block-visual { display: flex; justify-content: center; }
.ab-vcard { width: 100%; max-width: 360px; background: #fff; border: 1px solid #ededee; border-radius: 18px; padding: 30px; box-shadow: 0 30px 60px -34px rgba(20, 20, 20, 0.32); }
.ab-vico { display: inline-flex; width: 54px; height: 54px; border-radius: 14px; background: var(--lumen-accent); color: #fff; align-items: center; justify-content: center; margin-bottom: 22px; }
.ab-vico :deep(svg) { width: 28px; height: 28px; }
.ab-vrows { display: flex; flex-direction: column; gap: 12px; }
.ab-vrow { height: 12px; border-radius: 6px; background: linear-gradient(90deg, #ececee, #f6f6f7); }

/* 场景 */
.ab-scenes { max-width: 1100px; margin: 0 auto; padding: 30px 40px 20px; }
.ab-scene-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 18px; }
.ab-scene { background: #fff; border: 1px solid #ededee; border-radius: 16px; padding: 26px 24px; box-shadow: 0 14px 32px -26px rgba(20, 20, 20, 0.3); }
.ab-scene-t { font-size: 18px; font-weight: 800; margin-bottom: 10px; }
.ab-scene-d { font-size: 14px; color: #6b6b70; line-height: 1.65; }

/* 底部 CTA */
.ab-bottom { max-width: 1100px; margin: 0 auto; padding: 56px 40px 60px; text-align: center; }
.ab-bottom h2 { font-size: 40px; font-weight: 850; letter-spacing: -0.02em; margin: 0 0 12px; }
.ab-bottom p { font-size: 16px; color: #9a9a9e; margin-bottom: 28px; }
.ab-btn-lg { padding: 16px 32px; font-size: 16px; }

.ab-foot { max-width: 1180px; margin: 0 auto; padding: 26px 40px 40px; display: flex; justify-content: space-between; color: #b6b6ba; font-size: 13px; border-top: 1px solid #f0f0f1; }

@media (max-width: 900px) {
  .ab-hero { grid-template-columns: 1fr; }
  .ab-hero-card { display: none; }
  .ab-hero h1 { font-size: 52px; }
  .ab-block { grid-template-columns: 1fr; gap: 24px; }
  .ab-block.reverse .ab-block-text { order: 0; }
  .ab-block-visual { display: none; }
  .ab-scene-grid { grid-template-columns: 1fr; }
  .ab-stats { flex-direction: column; }
}
</style>
