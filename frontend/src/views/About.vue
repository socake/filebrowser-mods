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
          <a href="#features" class="ab-btn-ghost">了解功能</a>
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

    <!-- 功能 -->
    <section class="ab-features" id="features">
      <div class="ab-sec-head">
        <h2>一个入口，搞定文件的一切</h2>
        <p>浏览、预览、分享、权限、AI 操作，都在同一处优雅完成。</p>
      </div>
      <div class="ab-grid">
        <div class="ab-feat" v-for="f in feats" :key="f.t">
          <span class="ab-fi" v-html="f.icon"></span>
          <div class="ab-ft">{{ f.t }}</div>
          <div class="ab-fd">{{ f.d }}</div>
        </div>
      </div>
    </section>

    <!-- 底部 CTA -->
    <section class="ab-bottom">
      <h2>现在就把文件带到光下</h2>
      <p>自托管部署，数据归你掌控。</p>
      <router-link to="/login" class="ab-btn ab-btn-lg">进入 LumenBrowser →</router-link>
    </section>

    <footer class="ab-foot">
      <span>LumenBrowser · Bring files to light</span>
      <span>基于 filebrowser · Apache-2.0</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
const heroFiles = [
  { t: "DOC", n: "季度报告.docx", s: "预览", c: "#2b7cd3" },
  { t: "XLS", n: "财务表.xlsx", s: "下载分享", c: "#1e9e5a" },
  { t: "IMG", n: "风景.png", s: "在线预览", c: "#8a8a8e" },
  { t: "MD", n: "笔记.md", s: "即点即看", c: "#5b5b60" },
];
const feats = [
  {
    t: "AI 助手接入（MCP）",
    d: "接入 Claude / qwen，一句话让 AI 浏览、搜索、整理、分享你的文件。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="4" y="6" width="16" height="13" rx="3"/><path d="M9 3v3M15 3v3M9 13h.01M15 13h.01M9 16h6"/></svg>',
  },
  {
    t: "全类型即点即看",
    d: "图片 / 视频 / Office / PDF / Markdown 在线预览，无需下载到本地。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/></svg>',
  },
  {
    t: "两种分享方式",
    d: "预览分享在线看、下载分享直接下；私有云、公有云皆可对外分享。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4"/></svg>',
  },
  {
    t: "RBAC 角色权限",
    d: "用角色统一管理权限，精细控制每个人能看什么、能改什么。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 2l8 4v6c0 5-3.5 8-8 10-4.5-2-8-5-8-10V6z"/><path d="M9 12l2 2 4-4"/></svg>',
  },
  {
    t: "命令行与自动化",
    d: "lumen CLI 脚本化操作文件，配合 MCP 让自动化与 AI 无缝衔接。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M7 8l-4 4 4 4M17 8l4 4-4 4M14 4l-4 16"/></svg>',
  },
  {
    t: "自托管 · 数据归你",
    d: "装在自己的服务器或云主机，不依赖第三方网盘，数据始终掌控在手。",
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/></svg>',
  },
];
</script>

<style scoped>
.lb-about {
  position: fixed;
  inset: 0;
  overflow: auto;
  background: #fcfcfd;
  color: #141414;
  font-family: "Inter", -apple-system, "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
  -webkit-font-smoothing: antialiased;
}
.lb-about::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(1100px circle at 80% 8%, rgba(20, 20, 20, 0.05), transparent 55%),
    radial-gradient(800px circle at 10% 70%, rgba(20, 20, 20, 0.035), transparent 55%);
}
.lb-about > * { position: relative; z-index: 1; }
.ab-nav { max-width: 1180px; margin: 0 auto; padding: 26px 40px; display: flex; align-items: center; justify-content: space-between; }
.ab-brand { display: flex; align-items: center; gap: 11px; }
.ab-logo { width: 38px; height: 38px; border-radius: 10px; background: #141414; display: flex; align-items: center; justify-content: center; }
.ab-logo i { width: 14px; height: 14px; border: 2px solid #fff; border-radius: 50%; position: relative; }
.ab-logo i::after { content: ""; position: absolute; inset: 3px; border: 1.6px solid #fff; border-radius: 50%; }
.ab-brand b { font-size: 19px; font-weight: 800; }
.ab-nav-login { font-size: 14px; font-weight: 700; color: #141414; text-decoration: none; padding: 9px 18px; border: 1px solid #e6e6e8; border-radius: 10px; background: #fff; transition: 0.15s; }
.ab-nav-login:hover { border-color: #141414; }

.ab-hero { max-width: 1180px; margin: 0 auto; padding: 40px 40px 60px; display: grid; grid-template-columns: 1.05fr 0.95fr; gap: 48px; align-items: center; }
.ab-chip { display: inline-flex; align-items: center; gap: 9px; background: #fff; border: 1px solid #ededee; border-radius: 99px; padding: 8px 16px; font-size: 13px; font-weight: 700; margin-bottom: 26px; box-shadow: 0 6px 18px -12px rgba(20, 20, 20, 0.2); }
.ab-dot { width: 8px; height: 8px; border-radius: 50%; background: #141414; }
.ab-hero h1 { font-size: 76px; line-height: 0.99; font-weight: 900; letter-spacing: -0.025em; margin: 0 0 24px; }
.ab-lede { font-size: 17px; line-height: 1.75; color: #5b5b60; max-width: 460px; margin-bottom: 32px; }
.ab-lede b { color: #141414; font-weight: 700; }
.ab-cta { display: flex; gap: 13px; align-items: center; }
.ab-btn { display: inline-block; background: #141414; color: #fff; font-size: 15px; font-weight: 700; padding: 13px 24px; border-radius: 12px; text-decoration: none; transition: 0.15s; }
.ab-btn:hover { transform: translateY(-1px); box-shadow: 0 14px 30px -12px rgba(20, 20, 20, 0.4); }
.ab-btn-ghost { font-size: 15px; font-weight: 700; color: #5b5b60; text-decoration: none; padding: 13px 18px; }
.ab-btn-ghost:hover { color: #141414; }

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

.ab-features { max-width: 1180px; margin: 0 auto; padding: 50px 40px; }
.ab-sec-head { text-align: center; margin-bottom: 40px; }
.ab-sec-head h2 { font-size: 36px; font-weight: 850; letter-spacing: -0.02em; margin: 0 0 12px; }
.ab-sec-head p { font-size: 16px; color: #9a9a9e; }
.ab-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 18px; }
.ab-feat { background: #fff; border: 1px solid #ededee; border-radius: 16px; padding: 26px 24px; box-shadow: 0 14px 32px -26px rgba(20, 20, 20, 0.3); }
.ab-fi { display: inline-flex; width: 44px; height: 44px; border-radius: 11px; background: #141414; color: #fff; align-items: center; justify-content: center; margin-bottom: 16px; }
.ab-fi :deep(svg) { width: 22px; height: 22px; }
.ab-ft { font-size: 17px; font-weight: 750; margin-bottom: 8px; }
.ab-fd { font-size: 14px; color: #6b6b70; line-height: 1.65; }

.ab-bottom { max-width: 1180px; margin: 0 auto; padding: 50px 40px 60px; text-align: center; }
.ab-bottom h2 { font-size: 40px; font-weight: 850; letter-spacing: -0.02em; margin: 0 0 12px; }
.ab-bottom p { font-size: 16px; color: #9a9a9e; margin-bottom: 28px; }
.ab-btn-lg { padding: 16px 32px; font-size: 16px; }

.ab-foot { max-width: 1180px; margin: 0 auto; padding: 26px 40px 40px; display: flex; justify-content: space-between; color: #b6b6ba; font-size: 13px; border-top: 1px solid #f0f0f1; }

@media (max-width: 900px) {
  .ab-hero { grid-template-columns: 1fr; }
  .ab-hero-card { display: none; }
  .ab-hero h1 { font-size: 54px; }
  .ab-grid { grid-template-columns: 1fr; }
}
</style>
