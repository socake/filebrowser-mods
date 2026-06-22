<template>
  <div id="login" class="lumen-login" :class="{ recaptcha: recaptcha }">
    <!-- 左：宣发 -->
    <div class="lp-promo">
      <div class="lp-glow lp-glow-a"></div>
      <div class="lp-glow lp-glow-b"></div>

      <div class="lp-brand">
        <svg viewBox="0 0 96 96" fill="none">
          <rect x="22" y="22" width="52" height="52" rx="14" stroke="#fff" stroke-width="5" />
          <circle cx="48" cy="48" r="7" fill="#fff" />
          <g stroke="#fff" stroke-width="5" stroke-linecap="round">
            <line x1="48" y1="29" x2="48" y2="36" /><line x1="48" y1="60" x2="48" y2="67" />
            <line x1="29" y1="48" x2="36" y2="48" /><line x1="60" y1="48" x2="67" y2="48" />
            <line x1="34.5" y1="34.5" x2="39.5" y2="39.5" /><line x1="56.5" y1="56.5" x2="61.5" y2="61.5" />
            <line x1="61.5" y1="34.5" x2="56.5" y2="39.5" /><line x1="39.5" y1="56.5" x2="34.5" y2="61.5" />
          </g>
        </svg>
        <b>{{ name }}</b>
      </div>

      <div class="lp-hero">
        <h1>你的文件，<br />之光。</h1>
        <p>
          自托管的私有文件库 —— 在线浏览、对外文档门户、免登录投递、命令行客户端，数据始终归你掌控。
        </p>
      </div>

      <div class="lp-feats">
        <div class="lp-feat">
          <div class="lp-fi">
            <svg viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="1.7"><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
          </div>
          <div><div class="lp-ft">自托管 · 数据完全归你</div><div class="lp-fd">装在自己的服务器，不依赖第三方网盘</div></div>
        </div>
        <div class="lp-feat">
          <div class="lp-fi">
            <svg viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="1.7"><path d="M4 4h16v14H4z" /><path d="M8 8h8M8 12h6" /></svg>
          </div>
          <div><div class="lp-ft">公开文档门户 · 免登录投递</div><div class="lp-fd">把目录当在线文档库，任何人可投递文件</div></div>
        </div>
        <div class="lp-feat">
          <div class="lp-fi">
            <svg viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="1.7"><path d="M7 8l-4 4 4 4M17 8l4 4-4 4M14 4l-4 16" /></svg>
          </div>
          <div><div class="lp-ft">lumen 命令行 · 即将支持 MCP</div><div class="lp-fd">脚本化操作，AI 也能直接管理你的文件</div></div>
        </div>
      </div>

      <div class="lp-foot">© 2026 {{ name }} · 基于 filebrowser · Apache-2.0</div>
    </div>

    <!-- 右：登录 -->
    <div class="lp-auth">
      <form @submit="submit" class="lp-card">
        <img :src="logoURL" :alt="name" class="lp-logo" />
        <h2>{{ createMode ? t("login.signup") : "欢迎回来" }}</h2>
        <div class="lp-sub">登录到你的 {{ name }}</div>

        <p v-if="reason != null" class="logout-message">
          {{ t(`login.logout_reasons.${reason}`) }}
        </p>
        <div v-if="error !== ''" class="wrong">{{ error }}</div>

        <input
          autofocus
          class="lp-input"
          type="text"
          autocapitalize="off"
          v-model="username"
          :placeholder="t('login.username')"
        />
        <input
          class="lp-input"
          type="password"
          v-model="password"
          :placeholder="t('login.password')"
        />
        <input
          class="lp-input"
          v-if="createMode"
          type="password"
          v-model="passwordConfirm"
          :placeholder="t('login.passwordConfirm')"
        />

        <div v-if="recaptcha" id="recaptcha"></div>
        <input
          class="lp-btn"
          type="submit"
          :value="createMode ? t('login.signup') : t('login.submit')"
        />

        <p class="lp-toggle" @click="toggleMode" v-if="signup">
          {{ createMode ? t("login.loginInstead") : t("login.createAnAccount") }}
        </p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { StatusError } from "@/api/utils";
import * as auth from "@/utils/auth";
import {
  name,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
} from "@/utils/constants";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

// Define refs
const createMode = ref<boolean>(false);
const error = ref<string>("");
const username = ref<string>("");
const password = ref<string>("");
const passwordConfirm = ref<string>("");

const route = useRoute();
const router = useRouter();
const { t } = useI18n({});
// Define functions
const toggleMode = () => (createMode.value = !createMode.value);

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

  if (createMode.value) {
    if (password.value !== passwordConfirm.value) {
      error.value = t("login.passwordsDontMatch");
      return;
    }
  }

  try {
    if (createMode.value) {
      await auth.signup(username.value, password.value);
    }

    await auth.login(username.value, password.value, captcha);
    router.push({ path: redirect });
  } catch (e: any) {
    // console.error(e);
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

// Run hooks
onMounted(() => {
  if (!recaptcha) return;

  window.grecaptcha.ready(function () {
    window.grecaptcha.render("recaptcha", {
      sitekey: recaptchaKey,
    });
  });
});
</script>

<style>
/* LumenBrowser 登录页：宣发 + 登录联合页（#login.lumen-login 前缀覆盖原 login.css 居中布局） */
#login.lumen-login {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  background: #fff;
}

/* 左：宣发 */
#login.lumen-login .lp-promo {
  background: var(--lumen-accent);
  color: #fff;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 56px 60px;
}
#login.lumen-login .lp-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  background: #fff;
}
#login.lumen-login .lp-glow-a { width: 380px; height: 380px; top: -90px; right: -80px; opacity: 0.16; }
#login.lumen-login .lp-glow-b { width: 300px; height: 300px; bottom: -60px; left: -40px; opacity: 0.1; }
#login.lumen-login .lp-promo > *:not(.lp-glow) { position: relative; z-index: 1; }
#login.lumen-login .lp-brand { display: flex; align-items: center; gap: 13px; }
#login.lumen-login .lp-brand svg { width: 34px; height: 34px; }
#login.lumen-login .lp-brand b { font-size: 21px; font-weight: 750; letter-spacing: -0.02em; }
#login.lumen-login .lp-hero h1 { font-size: 52px; line-height: 1.08; letter-spacing: -0.03em; font-weight: 800; margin-bottom: 20px; color: #fff; }
#login.lumen-login .lp-hero p { font-size: 17px; color: rgba(255, 255, 255, 0.62); max-width: 440px; line-height: 1.65; }
#login.lumen-login .lp-feats { display: flex; flex-direction: column; gap: 18px; }
#login.lumen-login .lp-feat { display: flex; align-items: center; gap: 14px; }
#login.lumen-login .lp-fi { width: 40px; height: 40px; border-radius: 11px; border: 1px solid rgba(255, 255, 255, 0.16); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
#login.lumen-login .lp-fi svg { width: 19px; height: 19px; }
#login.lumen-login .lp-ft { font-size: 14.5px; font-weight: 600; }
#login.lumen-login .lp-fd { font-size: 12.5px; color: rgba(255, 255, 255, 0.5); }
#login.lumen-login .lp-foot { font-size: 12.5px; color: rgba(255, 255, 255, 0.4); }

/* 右：登录 */
#login.lumen-login .lp-auth { display: flex; align-items: center; justify-content: center; padding: 40px; background: #fff; }
#login.lumen-login .lp-card { position: static; transform: none; width: 100%; max-width: 360px; display: block; }
#login.lumen-login .lp-logo { width: 52px; height: 52px; margin-bottom: 26px; }
#login.lumen-login .lp-auth h2 { font-size: 27px; font-weight: 780; letter-spacing: -0.02em; margin: 0 0 6px; text-align: left; }
#login.lumen-login .lp-sub { font-size: 14px; color: #9a9a9e; margin-bottom: 28px; }
#login.lumen-login .lp-input { width: 100%; height: 46px; border: 1px solid #ececee; border-radius: 11px; padding: 0 14px; font-size: 14.5px; color: var(--lumen-accent); outline: none; transition: 0.15s; margin-bottom: 14px; background: #fff; box-sizing: border-box; }
#login.lumen-login .lp-input:focus { border-color: var(--lumen-accent); box-shadow: 0 0 0 4px rgba(20, 20, 20, 0.05); }
#login.lumen-login .lp-input::placeholder { color: #c2c2c6; }
#login.lumen-login .lp-btn { width: 100%; height: 48px; background: var(--lumen-accent); color: #fff; border: 0; border-radius: 12px; font-size: 15px; font-weight: 650; cursor: pointer; margin-top: 6px; transition: 0.15s; }
#login.lumen-login .lp-btn:hover { transform: translateY(-1px); box-shadow: 0 12px 28px -12px rgba(20, 20, 20, 0.4); }
#login.lumen-login .lp-toggle { margin-top: 18px; font-size: 13px; color: #5f5f63; cursor: pointer; text-align: center; }
#login.lumen-login .lp-toggle:hover { color: var(--lumen-accent); }
#login.lumen-login .wrong { background: #fde8e8; color: #c0392b; padding: 10px 12px; border-radius: 9px; font-size: 13px; margin-bottom: 14px; text-align: center; }
#login.lumen-login .logout-message { background: #fff4e5; color: #b76e00; padding: 10px 12px; border-radius: 9px; font-size: 13px; margin-bottom: 14px; text-align: center; text-transform: none; }
#login.lumen-login #recaptcha { margin-bottom: 14px; }

@media (max-width: 820px) {
  #login.lumen-login { grid-template-columns: 1fr; }
  #login.lumen-login .lp-promo { display: none; }
}
</style>
