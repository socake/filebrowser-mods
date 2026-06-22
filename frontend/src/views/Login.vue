<template>
  <div id="login" class="lumen-login" :class="{ recaptcha: recaptcha }">
    <div class="ll-wrap">
      <!-- 顶部 -->
      <div class="ll-top">
        <div class="ll-brand">
          <span class="ll-logo"><i></i></span>
          <b>{{ name }}</b>
        </div>
        <div class="ll-top-right">
          <router-link to="/about" class="ll-about">产品介绍 ↗</router-link>
          <div class="ll-tag">SELF-HOSTED FILE WORKSPACE</div>
        </div>
      </div>

      <div class="ll-main">
        <!-- 左：宣传 -->
        <div class="ll-promo">
          <div class="ll-chip"><span class="ll-dot"></span>文件管理 · AI 接入 · 私有部署</div>
          <h1>把你的文件<br />带到光下。</h1>
          <p class="ll-lede">
            不只是文件管理——更接入 <b>AI 助手</b>，让它用自然语言帮你浏览、整理、分享。一个安静、克制的<b>私有或云端</b>文件工作台。
          </p>
          <div class="ll-feats">
            <div class="ll-feat lead">
              <span class="ll-fi"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><rect x="4" y="6" width="16" height="13" rx="3" /><path d="M9 3v3M15 3v3M9 13h.01M15 13h.01M9 16h6" /></svg></span>
              <div><div class="ll-ft">AI 助手接入</div><div class="ll-fd">接入 Claude / qwen，一句话让 AI 浏览、搜索、整理、分享你的文件</div></div>
              <span class="ll-badge">MCP</span>
            </div>
            <div class="ll-feat">
              <span class="ll-fi"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></svg></span>
              <div><div class="ll-ft">全类型即点即看</div><div class="ll-fd">图片 / 视频 / Office / PDF / Markdown 在线预览，无需下载</div></div>
            </div>
            <div class="ll-feat">
              <span class="ll-fi"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M12 2l8 4v6c0 5-3.5 8-8 10-4.5-2-8-5-8-10V6z" /><path d="M9 12l2 2 4-4" /></svg></span>
              <div><div class="ll-ft">角色权限与分享</div><div class="ll-fd">RBAC 精细控制谁能看、谁能改；预览 / 下载分享，私有云公有云皆可</div></div>
            </div>
          </div>
        </div>

        <!-- 右：登录 / 注册 -->
        <div class="ll-auth">
          <div class="ll-card">
            <div class="ll-tabs">
              <button type="button" :class="{ on: tab === 'login' }" @click="tab = 'login'">登录</button>
              <button type="button" :class="{ on: tab === 'reg' }" @click="tab = 'reg'">注册</button>
            </div>

            <!-- 登录 -->
            <form v-show="tab === 'login'" @submit="submit" class="ll-pane">
              <div class="ll-ttl">欢迎回来</div>
              <div class="ll-sub">登录以进入你的文件中枢</div>

              <p v-if="reason != null" class="ll-logout">
                {{ t(`login.logout_reasons.${reason}`) }}
              </p>
              <div v-if="error !== ''" class="ll-wrong">{{ error }}</div>

              <div class="ll-fld">
                <label>{{ t("login.username") }}</label>
                <input autofocus type="text" autocapitalize="off" v-model="username" placeholder="you@example.com" />
              </div>
              <div class="ll-fld ll-fld-pwd">
                <label>{{ t("login.password") }}</label>
                <input :type="showLoginPwd ? 'text' : 'password'" v-model="password" placeholder="••••••••" />
                <button
                  type="button"
                  class="ll-eye"
                  :aria-label="showLoginPwd ? '隐藏密码' : '显示密码'"
                  @click="showLoginPwd = !showLoginPwd"
                >
                  <svg v-if="showLoginPwd" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M3 3l18 18M10.6 10.7a3 3 0 0 0 4.2 4.2M9.9 5.2A9.6 9.6 0 0 1 12 5c6 0 10 7 10 7a17 17 0 0 1-3.3 3.9M6.6 6.7A17 17 0 0 0 2 12s4 7 10 7a9.5 9.5 0 0 0 3.1-.5" /></svg>
                </button>
              </div>

              <div v-if="recaptcha" id="recaptcha"></div>
              <input class="ll-btn" type="submit" :value="t('login.submit')" />
            </form>

            <!-- 注册：完整邮箱表单(占坑),按钮禁用 + 常显演示环境提示 -->
            <div v-show="tab === 'reg'" class="ll-pane">
              <div class="ll-ttl">创建账号</div>
              <div class="ll-sub">用邮箱注册，开启你的文件之光</div>
              <div class="ll-fld">
                <label>邮箱</label>
                <input type="email" placeholder="you@example.com" />
              </div>
              <div class="ll-fld ll-fld-pwd">
                <label>密码</label>
                <input :type="showRegPwd ? 'text' : 'password'" placeholder="设置登录密码" />
                <button
                  type="button"
                  class="ll-eye"
                  :aria-label="showRegPwd ? '隐藏密码' : '显示密码'"
                  @click="showRegPwd = !showRegPwd"
                >
                  <svg v-if="showRegPwd" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M3 3l18 18M10.6 10.7a3 3 0 0 0 4.2 4.2M9.9 5.2A9.6 9.6 0 0 1 12 5c6 0 10 7 10 7a17 17 0 0 1-3.3 3.9M6.6 6.7A17 17 0 0 0 2 12s4 7 10 7a9.5 9.5 0 0 0 3.1-.5" /></svg>
                </button>
              </div>
              <div class="ll-fld ll-fld-pwd">
                <label>确认密码</label>
                <input :type="showRegPwd2 ? 'text' : 'password'" placeholder="再次输入密码" />
                <button
                  type="button"
                  class="ll-eye"
                  :aria-label="showRegPwd2 ? '隐藏密码' : '显示密码'"
                  @click="showRegPwd2 = !showRegPwd2"
                >
                  <svg v-if="showRegPwd2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></svg>
                  <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><path d="M3 3l18 18M10.6 10.7a3 3 0 0 0 4.2 4.2M9.9 5.2A9.6 9.6 0 0 1 12 5c6 0 10 7 10 7a17 17 0 0 1-3.3 3.9M6.6 6.7A17 17 0 0 0 2 12s4 7 10 7a9.5 9.5 0 0 0 3.1-.5" /></svg>
                </button>
              </div>
              <div class="ll-reg-note">当前为演示环境，暂不开放注册</div>
              <button type="button" class="ll-btn ll-btn-off" disabled>注 册</button>
            </div>
          </div>
        </div>
      </div>

      <div class="ll-foot"><span class="ll-foot-ln"></span>{{ name }} · Bring files to light · Powered by <a href="https://github.com/socake" target="_blank" rel="noopener" style="color:var(--lumen-accent);text-decoration:none;margin-left:4px">socake</a></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { StatusError } from "@/api/utils";
import * as auth from "@/utils/auth";
import { name, recaptcha, recaptchaKey } from "@/utils/constants";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

const tab = ref<"login" | "reg">("login");
const showLoginPwd = ref(false);
const showRegPwd = ref(false);
const showRegPwd2 = ref(false);
const error = ref<string>("");
const username = ref<string>("");
const password = ref<string>("");

const route = useRoute();
const router = useRouter();
const { t } = useI18n({});

const $showError = inject<IToastError>("$showError")!;

const reason = route.query["logout-reason"] ?? null;

const submit = async (event: Event) => {
  event.preventDefault();
  event.stopPropagation();

  const redirect = (route.query.redirect || "/files/") as string;

  let captcha = "";
  if (recaptcha) {
    captcha = window.grecaptcha.getResponse();
    if (captcha === "") {
      error.value = t("login.wrongCredentials");
      return;
    }
  }

  try {
    await auth.login(username.value, password.value, captcha);
    router.push({ path: redirect });
  } catch (e: any) {
    if (e instanceof StatusError) {
      if (e.status === 409) {
        error.value = t("login.usernameTaken");
      } else if (e.status === 403) {
        error.value = t("login.wrongCredentials");
      } else if (e.status === 400) {
        const match = e.message.match(/minimum length is (\d+)/);
        if (match) {
          error.value = t("login.passwordTooShort", { min: match[1] });
        } else {
          error.value = e.message;
        }
      } else {
        $showError(e);
      }
    }
  }
};

onMounted(() => {
  if (!recaptcha) return;
  window.grecaptcha.ready(function () {
    window.grecaptcha.render("recaptcha", { sitekey: recaptchaKey });
  });
});
</script>

<style>
/* LumenBrowser 登录页 v2：宣传 + 登录联合页。#login.lumen-login 前缀覆盖原 login.css 居中布局 */
#login.lumen-login {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  background: #fcfcfd;
  overflow: auto;
}
/* 呼应「光」的柔和光晕，统一黑白配色 */
#login.lumen-login::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(1100px circle at 78% 32%, rgba(20, 20, 20, 0.05), transparent 58%),
    radial-gradient(800px circle at 12% 88%, rgba(20, 20, 20, 0.035), transparent 55%);
}
#login.lumen-login .ll-wrap {
  position: relative;
  z-index: 1;
  max-width: 1360px;
  margin: 0 auto;
  padding: 40px 60px;
  min-height: 100%;
  display: flex;
  flex-direction: column;
}
#login.lumen-login .ll-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; }
#login.lumen-login .ll-brand { display: flex; align-items: center; gap: 12px; }
#login.lumen-login .ll-logo { width: 42px; height: 42px; border-radius: 11px; background: var(--lumen-accent); display: flex; align-items: center; justify-content: center; }
#login.lumen-login .ll-logo i { display: block; width: 16px; height: 16px; border: 2.2px solid #fff; border-radius: 50%; position: relative; }
#login.lumen-login .ll-logo i::after { content: ""; position: absolute; inset: 3px; border: 1.8px solid #fff; border-radius: 50%; }
#login.lumen-login .ll-brand b { font-size: 20px; font-weight: 800; letter-spacing: -0.01em; color: var(--lumen-accent); }
#login.lumen-login .ll-tag { font-size: 12px; font-weight: 700; letter-spacing: 0.22em; color: #9a9a9e; }
#login.lumen-login .ll-top-right { display: flex; align-items: center; gap: 18px; }
#login.lumen-login .ll-about { font-size: 13px; font-weight: 700; color: #fff; background: var(--lumen-accent); text-decoration: none; letter-spacing: 0.01em; padding: 9px 17px; border-radius: 10px; transition: 0.15s; }
#login.lumen-login .ll-about:hover { transform: translateY(-1px); box-shadow: 0 10px 22px -10px rgba(20, 20, 20, 0.45); }

#login.lumen-login .ll-main { flex: 1; display: grid; grid-template-columns: 1.04fr 0.96fr; gap: 48px; align-items: center; }

/* 左宣传 */
#login.lumen-login .ll-chip { display: inline-flex; align-items: center; gap: 9px; background: #fff; border: 1px solid #ededee; border-radius: 99px; padding: 8px 16px; font-size: 13px; font-weight: 700; margin-bottom: 28px; box-shadow: 0 6px 18px -12px rgba(20, 20, 20, 0.2); color: var(--lumen-accent); }
#login.lumen-login .ll-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--lumen-accent); }
#login.lumen-login .ll-promo h1 { font-size: 80px; line-height: 0.99; font-weight: 900; letter-spacing: -0.025em; margin: 0 0 26px; color: var(--lumen-accent); }
#login.lumen-login .ll-lede { font-size: 16.5px; line-height: 1.75; color: #5b5b60; max-width: 470px; margin-bottom: 34px; }
#login.lumen-login .ll-lede b { color: var(--lumen-accent); font-weight: 700; }
#login.lumen-login .ll-feats { display: flex; flex-direction: column; gap: 12px; max-width: 470px; }
#login.lumen-login .ll-feat { display: flex; gap: 14px; align-items: center; background: #fff; border: 1px solid #ededee; border-radius: 14px; padding: 15px 17px; box-shadow: 0 10px 26px -20px rgba(20, 20, 20, 0.25); }
#login.lumen-login .ll-fi { flex-shrink: 0; width: 38px; height: 38px; border-radius: 10px; background: var(--lumen-accent); color: #fff; display: flex; align-items: center; justify-content: center; }
#login.lumen-login .ll-fi svg { width: 20px; height: 20px; }
#login.lumen-login .ll-ft { font-size: 14.5px; font-weight: 750; color: var(--lumen-accent); }
#login.lumen-login .ll-fd { font-size: 12.5px; color: #9a9a9e; margin-top: 2px; line-height: 1.5; }
#login.lumen-login .ll-badge { margin-left: auto; font-size: 10px; font-weight: 800; letter-spacing: 0.05em; color: #fff; background: var(--lumen-accent); padding: 3px 9px; border-radius: 99px; }

/* 右登录卡（悬浮立体） */
#login.lumen-login .ll-auth { display: flex; justify-content: center; perspective: 1400px; }
#login.lumen-login .ll-card { background: #fff; border: 1px solid #ededee; border-radius: 22px; padding: 34px 36px 30px; width: 100%; max-width: 412px; box-shadow: 0 50px 90px -40px rgba(20, 20, 20, 0.4), 0 16px 34px -22px rgba(20, 20, 20, 0.22); transform: rotateY(-3deg) rotateX(1.2deg); transition: 0.3s; }
#login.lumen-login .ll-card:hover { transform: rotateY(0) rotateX(0); }
#login.lumen-login .ll-tabs { display: flex; gap: 4px; background: #f4f4f5; border-radius: 12px; padding: 4px; margin-bottom: 24px; }
#login.lumen-login .ll-tabs button { flex: 1; height: 38px; border: 0; background: transparent; border-radius: 9px; font-size: 14px; font-weight: 700; color: #9a9a9e; cursor: pointer; transition: 0.15s; }
#login.lumen-login .ll-tabs button.on { background: #fff; color: var(--lumen-accent); box-shadow: 0 2px 8px -4px rgba(20, 20, 20, 0.25); }
/* 重置原 login.css 对 form 的 fixed 居中 + max-width 压窄(否则脱离卡片重叠) */
#login.lumen-login .ll-pane { display: block; position: static; transform: none; top: auto; left: auto; right: auto; max-width: none; min-width: 0; width: auto; margin: 0; }
#login.lumen-login .ll-ttl { font-size: 22px; font-weight: 800; margin-bottom: 5px; color: var(--lumen-accent); }
#login.lumen-login .ll-sub { font-size: 13.5px; color: #9a9a9e; margin-bottom: 22px; }
#login.lumen-login .ll-fld { margin-bottom: 14px; position: relative; }
#login.lumen-login .ll-fld label { display: block; font-size: 12.5px; font-weight: 700; color: #5b5b60; margin-bottom: 7px; }
#login.lumen-login .ll-fld input { width: 100%; height: 46px; border: 1.5px solid #ededee; border-radius: 11px; padding: 0 14px; font-size: 14.5px; background: #fafafa; outline: none; transition: 0.15s; color: var(--lumen-accent); box-sizing: border-box; }
#login.lumen-login .ll-fld input:focus { border-color: var(--lumen-accent); background: #fff; box-shadow: 0 0 0 4px rgba(20, 20, 20, 0.05); }
#login.lumen-login .ll-fld input::placeholder { color: #c2c2c6; }
/* 密码框留出右侧眼睛按钮空间 */
#login.lumen-login .ll-fld-pwd input { padding-right: 46px; }
#login.lumen-login .ll-eye { position: absolute; right: 6px; bottom: 0; width: 40px; height: 46px; display: flex; align-items: center; justify-content: center; background: transparent; border: 0; padding: 0; cursor: pointer; color: #b4b4b8; transition: color 0.15s; }
#login.lumen-login .ll-eye:hover { color: var(--lumen-accent); }
#login.lumen-login .ll-eye svg { width: 19px; height: 19px; display: block; }
#login.lumen-login .ll-btn { width: 100%; height: 48px; border: 0; border-radius: 11px; background: var(--lumen-accent); color: #fff; font-size: 15px; font-weight: 700; cursor: pointer; margin-top: 6px; transition: 0.15s; }
#login.lumen-login .ll-btn:hover { transform: translateY(-1px); box-shadow: 0 12px 28px -12px rgba(20, 20, 20, 0.4); }
/* 注册按钮禁用态：演示环境暂不开放 */
#login.lumen-login .ll-btn-off, #login.lumen-login .ll-btn-off:hover { background: #d8d8dc; color: #fafafa; cursor: not-allowed; transform: none; box-shadow: none; }
/* 常显的演示环境提示，低调灰字 */
#login.lumen-login .ll-reg-note { text-align: center; font-size: 12.5px; color: #9a9a9e; margin: 4px 0 10px; letter-spacing: 0.01em; }
#login.lumen-login .ll-wrong { background: #fde8e8; color: #c0392b; padding: 10px 12px; border-radius: 9px; font-size: 13px; margin-bottom: 14px; text-align: center; }
#login.lumen-login .ll-logout { background: #fff4e5; color: #b76e00; padding: 10px 12px; border-radius: 9px; font-size: 13px; margin-bottom: 14px; text-align: center; text-transform: none; }
#login.lumen-login #recaptcha { margin-bottom: 14px; }
#login.lumen-login .ll-foot { display: flex; justify-content: flex-end; align-items: center; gap: 10px; margin-top: 18px; color: #9a9a9e; font-size: 12.5px; font-weight: 600; }
#login.lumen-login .ll-foot-ln { width: 38px; height: 1px; background: #9a9a9e; opacity: 0.5; }

@media (max-width: 880px) {
  #login.lumen-login .ll-main { grid-template-columns: 1fr; gap: 28px; }
  #login.lumen-login .ll-promo h1 { font-size: 56px; }
  #login.lumen-login .ll-card { transform: none; }
}
</style>
