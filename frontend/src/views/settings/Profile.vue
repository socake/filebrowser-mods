<template>
  <div class="toc-wrap">
    <form class="toc-card" @submit="updateSettings">
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.profileSettings") }}</h2>
        <p class="toc-card-sub">{{ t("settings.language") }} · {{ t("settings.themes.title") }}</p>
      </div>

      <div class="toc-section">
        <label class="toc-check">
          <input type="checkbox" name="hideDotfiles" v-model="hideDotfiles" />
          <span>{{ t("settings.hideDotfiles") }}</span>
        </label>
        <label class="toc-check">
          <input
            type="checkbox"
            name="loadThumbnails"
            v-model="loadThumbnails"
          />
          <span>{{ t("settings.loadThumbnails") }}</span>
        </label>
        <label class="toc-check">
          <input type="checkbox" name="singleClick" v-model="singleClick" />
          <span>{{ t("settings.singleClick") }}</span>
        </label>
        <label class="toc-check">
          <input
            type="checkbox"
            name="redirectAfterCopyMove"
            v-model="redirectAfterCopyMove"
          />
          <span>{{ t("settings.redirectAfterCopyMove") }}</span>
        </label>
        <label class="toc-check">
          <input type="checkbox" name="dateFormat" v-model="dateFormat" />
          <span>{{ t("settings.setDateFormat") }}</span>
        </label>
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.language") }}</label>
        <languages class="toc-input" v-model:locale="locale"></languages>
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.aceEditorTheme") }}</label>
        <AceEditorTheme
          class="toc-input"
          v-model:aceEditorTheme="aceEditorTheme"
          id="aceTheme"
        ></AceEditorTheme>
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit" name="submitProfile">
          {{ t("buttons.update") }}
        </button>
      </div>
    </form>

    <form
      v-if="!noAuth && !authStore.user?.lockPassword"
      class="toc-card"
      @submit="updatePassword"
    >
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.changePassword") }}</h2>
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.newPassword") }}</label>
        <input
          :class="['toc-input', passwordStateClass]"
          type="password"
          :placeholder="t('settings.newPassword')"
          v-model="password"
          name="password"
        />
      </div>
      <div class="toc-field">
        <label class="toc-label">{{ t("settings.newPasswordConfirm") }}</label>
        <input
          :class="['toc-input', passwordStateClass]"
          type="password"
          :placeholder="t('settings.newPasswordConfirm')"
          v-model="passwordConf"
          name="passwordConf"
        />
      </div>
      <div v-if="isCurrentPasswordRequired" class="toc-field">
        <label class="toc-label">{{ t("settings.currentPassword") }}</label>
        <input
          :class="['toc-input', passwordStateClass]"
          type="password"
          :placeholder="t('settings.currentPassword')"
          v-model="currentPassword"
          name="current_password"
          autocomplete="current-password"
        />
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit" name="submitPassword">
          {{ t("buttons.update") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { users as api } from "@/api";
import AceEditorTheme from "@/components/settings/AceEditorTheme.vue";
import Languages from "@/components/settings/Languages.vue";
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { authMethod, noAuth } from "@/utils/constants";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

const password = ref<string>("");
const passwordConf = ref<string>("");
const currentPassword = ref<string>("");
const isCurrentPasswordRequired = ref<boolean>(false);
const hideDotfiles = ref<boolean>(false);
const loadThumbnails = ref<boolean>(true);
const singleClick = ref<boolean>(false);
const redirectAfterCopyMove = ref<boolean>(false);
const dateFormat = ref<boolean>(false);
const locale = ref<string>("");
const aceEditorTheme = ref<string>("");

const passwordStateClass = computed(() => {
  if (password.value === "" && passwordConf.value === "") {
    return "";
  }

  if (password.value === passwordConf.value) {
    return "toc-input--green";
  }

  return "toc-input--red";
});

onMounted(async () => {
  layoutStore.loading = true;
  if (authStore.user === null) return false;
  locale.value = authStore.user.locale;
  hideDotfiles.value = authStore.user.hideDotfiles;
  loadThumbnails.value = !authStore.user.disableThumbnails;
  singleClick.value = authStore.user.singleClick;
  redirectAfterCopyMove.value = authStore.user.redirectAfterCopyMove;
  dateFormat.value = authStore.user.dateFormat;
  aceEditorTheme.value = authStore.user.aceEditorTheme;
  layoutStore.loading = false;
  isCurrentPasswordRequired.value = authMethod == "json";

  return true;
});

const updatePassword = async (event: Event) => {
  event.preventDefault();

  if (
    password.value !== passwordConf.value ||
    password.value === "" ||
    currentPassword.value === "" ||
    authStore.user === null
  ) {
    return;
  }

  try {
    const data = {
      ...authStore.user,
      id: authStore.user.id,
      password: password.value,
    };
    await api.update(data, ["password"], currentPassword.value);
    authStore.updateUser(data);
    $showSuccess(t("settings.passwordUpdated"));
  } catch (e: any) {
    $showError(e);
  } finally {
    password.value = passwordConf.value = "";
  }
};
const updateSettings = async (event: Event) => {
  event.preventDefault();

  try {
    if (authStore.user === null) throw new Error("User is not set!");

    const data = {
      ...authStore.user,
      id: authStore.user.id,
      locale: locale.value,
      hideDotfiles: hideDotfiles.value,
      disableThumbnails: !loadThumbnails.value,
      singleClick: singleClick.value,
      redirectAfterCopyMove: redirectAfterCopyMove.value,
      dateFormat: dateFormat.value,
      aceEditorTheme: aceEditorTheme.value,
    };

    await api.update(data, [
      "locale",
      "hideDotfiles",
      "disableThumbnails",
      "singleClick",
      "redirectAfterCopyMove",
      "dateFormat",
      "aceEditorTheme",
    ]);
    authStore.updateUser(data);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (err) {
    if (err instanceof Error) {
      $showError(err);
    }
  }
};
</script>

<style scoped>
.toc-wrap {
  max-width: 720px;
  margin: 0 auto;
  padding: 1.5em 1em 3em;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.toc-card {
  background: #fff;
  border: 1px solid var(--lumen-line, #ececee);
  border-radius: 14px;
  padding: 26px 28px;
}
.toc-card-head {
  margin-bottom: 20px;
}
.toc-card-title {
  font-size: 19px;
  font-weight: 700;
  margin: 0;
}
.toc-card-sub {
  font-size: 13px;
  color: #9a9a9e;
  margin: 6px 0 0;
}
.toc-section {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 22px;
}
.toc-check {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #232326;
  cursor: pointer;
}
.toc-check input[type="checkbox"] {
  width: 17px;
  height: 17px;
  accent-color: var(--lumen-accent, #141414);
  cursor: pointer;
  margin: 0;
}
.toc-field {
  margin-bottom: 18px;
}
.toc-field:last-of-type {
  margin-bottom: 0;
}
.toc-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #6b6b70;
  margin-bottom: 7px;
}
.toc-input,
.toc-wrap :deep(select.toc-input),
.toc-wrap :deep(.toc-input) {
  display: block;
  width: 100%;
  max-width: 360px;
  height: 40px;
  padding: 0 12px;
  font-size: 14px;
  color: #232326;
  background: #fff;
  border: 1px solid var(--lumen-line, #ececee);
  border-radius: 9px;
  box-shadow: none;
  outline: none;
}
.toc-input:focus,
.toc-wrap :deep(.toc-input):focus {
  border-color: var(--lumen-accent, #141414);
}
.toc-input--green {
  border-color: #2e9e5b;
}
.toc-input--red {
  border-color: #d64242;
}
.toc-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
.toc-btn {
  height: 40px;
  padding: 0 22px;
  border: 0;
  border-radius: 9px;
  background: var(--lumen-accent, #141414);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.toc-btn:hover {
  opacity: 0.9;
}
</style>
