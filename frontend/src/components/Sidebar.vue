<template>
  <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <template v-if="isLoggedIn">
      <button @click="toAccountSettings" class="action sb-user">
        <i class="material-icons">account_circle</i>
        <span>{{ user.username }}</span>
      </button>
      <button
        class="action"
        :class="{ 'sb-active': $route.path.startsWith('/files') }"
        @click="toRoot"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </button>

      <div v-if="user.perm.create">
        <button
          @click="newItem"
          class="action"
          :aria-label="$t('sidebar.new')"
          :title="$t('sidebar.new')"
        >
          <i class="material-icons">add</i>
          <span>{{ $t("sidebar.new") }}</span>
        </button>
      </div>

      <button
        v-if="user.perm.share"
        class="action"
        :class="{ 'sb-active': $route.path.startsWith('/shares') }"
        @click="toShares"
        :aria-label="$t('sidebar.shareManagement')"
        :title="$t('sidebar.shareManagement')"
      >
        <i class="material-icons">share</i>
        <span>{{ $t("sidebar.shareManagement") }}</span>
      </button>

      <button
        v-if="user.perm.admin"
        class="action"
        :class="{ 'sb-active': $route.path.startsWith('/roles') }"
        @click="toRoles"
        :aria-label="$t('sidebar.roleManagement')"
        :title="$t('sidebar.roleManagement')"
      >
        <i class="material-icons">verified_user</i>
        <span>{{ $t("sidebar.roleManagement") }}</span>
      </button>

      <button
        v-if="user.perm.admin"
        class="action"
        :class="{ 'sb-active': $route.path.startsWith('/users') }"
        @click="toUsers"
        :aria-label="$t('sidebar.userManagement')"
        :title="$t('sidebar.userManagement')"
      >
        <i class="material-icons">group</i>
        <span>{{ $t("sidebar.userManagement") }}</span>
      </button>

      <div v-if="user.perm.admin">
        <button
          class="action"
          :class="{ 'sb-active': $route.path.startsWith('/settings') }"
          @click="toGlobalSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>
      </div>
      <button
        v-if="canLogout"
        @click="logout"
        class="action"
        id="logout"
        :aria-label="$t('sidebar.logout')"
        :title="$t('sidebar.logout')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.logout") }}</span>
      </button>
    </template>
    <template v-else>
      <router-link
        v-if="!hideLoginButton"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.login") }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t("sidebar.signup") }}</span>
      </router-link>
    </template>

    <div
      class="credits"
      v-if="isFiles && !disableUsedPercentage"
      style="width: 90%; margin: 2em 2.5em 3em 2.5em"
    >
      <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
      <br />
      {{ usage.used }} of {{ usage.total }} used
    </div>

    <p class="credits">
      <span>
        <span v-if="disableExternal">{{ name }}</span>
        <a
          v-else
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/socake/LumenBrowser"
          >{{ name }}</a
        >
        <span> {{ " " }} {{ version }}</span>
        <span v-if="!disableExternal">
          · Powered by
          <a rel="noopener" target="_blank" href="https://github.com/socake" style="color: var(--lumen-accent)">socake</a></span
        >
      </span>
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import {
  name,
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    name: () => name,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      this.$router.push({ path: "/files" });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    toShares() {
      this.$router.push({ path: "/shares" });
      this.closeHovers();
    },
    toRoles() {
      this.$router.push({ path: "/roles" });
      this.closeHovers();
    },
    toUsers() {
      this.$router.push({ path: "/users" });
      this.closeHovers();
    },
    newItem() {
      if (this.$route.path.startsWith("/files")) {
        this.showHover("newItem");
      } else {
        this.$showError(this.$t("prompts.newItemWrongPage"), false);
      }
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>
