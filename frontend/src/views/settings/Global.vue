<template>
  <errors v-if="error" :errorCode="error.status" />
  <div
    class="toc-wrap"
    v-else-if="!layoutStore.loading && settings !== null"
  >
    <!-- 全局设置 -->
    <form class="toc-card" @submit.prevent="save">
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.globalSettings") }}</h2>
      </div>

      <div class="toc-section">
        <label class="toc-check">
          <input type="checkbox" v-model="settings.signup" />
          <span>{{ t("settings.allowSignup") }}</span>
        </label>
        <label class="toc-check">
          <input type="checkbox" v-model="settings.createUserDir" />
          <span>{{ t("settings.createUserDir") }}</span>
        </label>
        <label class="toc-check">
          <input type="checkbox" v-model="settings.hideLoginButton" />
          <span>{{ t("settings.hideLoginButton") }}</span>
        </label>
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.userHomeBasePath") }}</label>
        <input
          class="toc-input"
          type="text"
          v-model="settings.userHomeBasePath"
        />
      </div>

      <div class="toc-field">
        <label class="toc-label" for="minimumPasswordLength">{{
          t("settings.minimumPasswordLength")
        }}</label>
        <vue-number-input
          controls
          v-model.number="settings.minimumPasswordLength"
          id="minimumPasswordLength"
          :min="1"
        />
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.rules") }}</label>
        <p class="toc-hint">{{ t("settings.globalRules") }}</p>
        <rules v-model:rules="settings.rules" />
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit">{{ t("buttons.update") }}</button>
      </div>
    </form>

    <!-- 新用户默认 -->
    <form class="toc-card" @submit.prevent="save">
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.userDefaults") }}</h2>
        <p class="toc-card-sub">{{ t("settings.defaultUserDescription") }}</p>
      </div>

      <div class="toc-field">
        <label class="toc-label" for="defaultRole">新用户默认角色</label>
        <select
          id="defaultRole"
          class="toc-input"
          v-model.number="defaultRoleID"
        >
          <option :value="0">{{ t("roles.noRole") }}</option>
          <option v-for="r in roles" :key="r.id" :value="r.id">
            {{ r.name }}
          </option>
        </select>
        <p class="toc-hint">新注册用户将自动套用所选角色的权限。</p>
      </div>

      <div class="toc-field">
        <label class="toc-label" for="default-scope">{{
          t("settings.scope")
        }}</label>
        <input
          id="default-scope"
          class="toc-input"
          type="text"
          v-model="settings.defaults.scope"
        />
      </div>

      <div class="toc-field">
        <label class="toc-label">{{ t("settings.language") }}</label>
        <languages
          class="toc-input"
          v-model:locale="settings.defaults.locale"
        ></languages>
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit">{{ t("buttons.update") }}</button>
      </div>
    </form>

    <!-- 品牌与外观 -->
    <form class="toc-card" @submit.prevent="save">
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.branding") }}</h2>
      </div>

      <i18n-t
        keypath="settings.brandingHelp"
        tag="p"
        class="toc-hint"
        scope="global"
      >
        <a
          class="link"
          target="_blank"
          href="https://filebrowser.org/configuration.html#custom-branding"
          >{{ t("settings.documentation") }}</a
        >
      </i18n-t>

      <div class="toc-section">
        <label class="toc-check">
          <input
            type="checkbox"
            v-model="settings.branding.disableExternal"
            id="branding-links"
          />
          <span>{{ t("settings.disableExternalLinks") }}</span>
        </label>
        <label class="toc-check">
          <input
            type="checkbox"
            v-model="settings.branding.disableUsedPercentage"
            id="branding-used-disk"
          />
          <span>{{ t("settings.disableUsedDiskPercentage") }}</span>
        </label>
      </div>

      <div class="toc-field">
        <label class="toc-label" for="theme">{{
          t("settings.themes.title")
        }}</label>
        <themes
          class="toc-input"
          v-model:theme="settings.branding.theme"
          id="theme"
        ></themes>
      </div>

      <div class="toc-field">
        <label class="toc-label" for="branding-name">{{
          t("settings.instanceName")
        }}</label>
        <input
          class="toc-input"
          type="text"
          v-model="settings.branding.name"
          id="branding-name"
        />
      </div>

      <div class="toc-field">
        <label class="toc-label" for="branding-files">{{
          t("settings.brandingDirectoryPath")
        }}</label>
        <input
          class="toc-input"
          type="text"
          v-model="settings.branding.files"
          id="branding-files"
        />
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit">{{ t("buttons.update") }}</button>
      </div>
    </form>

    <!-- 上传 -->
    <form class="toc-card" @submit.prevent="save">
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.tusUploads") }}</h2>
        <p class="toc-card-sub">{{ t("settings.tusUploadsHelp") }}</p>
      </div>

      <div class="toc-field">
        <label class="toc-label" for="tus-chunkSize">{{
          t("settings.tusUploadsChunkSize")
        }}</label>
        <input
          class="toc-input"
          type="text"
          v-model="formattedChunkSize"
          id="tus-chunkSize"
        />
      </div>

      <div class="toc-field">
        <label class="toc-label" for="tus-retryCount">{{
          t("settings.tusUploadsRetryCount")
        }}</label>
        <vue-number-input
          controls
          v-model.number="settings.tus.retryCount"
          id="tus-retryCount"
          :min="0"
        />
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit">{{ t("buttons.update") }}</button>
      </div>
    </form>

    <!-- 命令执行 -->
    <form
      v-if="enableExec"
      class="toc-card"
      @submit.prevent="save"
    >
      <div class="toc-card-head">
        <h2 class="toc-card-title">{{ t("settings.executeOnShell") }}</h2>
        <p class="toc-card-sub">{{ t("settings.executeOnShellDescription") }}</p>
      </div>

      <div class="toc-field">
        <input
          class="toc-input"
          type="text"
          placeholder="bash -c, cmd /c, ..."
          v-model="shellValue"
        />
      </div>

      <i18n-t
        keypath="settings.commandRunnerHelp"
        tag="p"
        class="toc-hint"
        scope="global"
      >
        <code>FILE</code>
        <code>SCOPE</code>
        <a
          class="link"
          target="_blank"
          href="https://filebrowser.org/configuration.html#command-runner"
          >{{ t("settings.documentation") }}</a
        >
      </i18n-t>

      <div
        v-for="(command, key) in settings.commands"
        :key="key"
        class="collapsible"
      >
        <input :id="key" type="checkbox" />
        <label :for="key">
          <p>{{ capitalize(key) }}</p>
          <i class="material-icons">arrow_drop_down</i>
        </label>
        <div class="collapse">
          <textarea
            class="input input--block input--textarea"
            v-model.trim="commandObject[key]"
          ></textarea>
        </div>
      </div>

      <div class="toc-actions">
        <button class="toc-btn" type="submit">{{ t("buttons.update") }}</button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { settings as api, roles as rolesApi } from "@/api";
import { StatusError } from "@/api/utils";
import Rules from "@/components/settings/Rules.vue";
import Themes from "@/components/settings/Themes.vue";
import Languages from "@/components/settings/Languages.vue";
import { useLayoutStore } from "@/stores/layout";
import { enableExec } from "@/utils/constants";
import { getTheme, setTheme } from "@/utils/theme";
import Errors from "@/views/Errors.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const originalSettings = ref<ISettings | null>(null);
const settings = ref<ISettings | null>(null);
const debounceTimeout = ref<number | null>(null);
const roles = ref<Role[]>([]);
const defaultRoleID = ref<number>(0);

const commandObject = ref<{
  [key: string]: string[] | string;
}>({});
const shellValue = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();

const layoutStore = useLayoutStore();

const formattedChunkSize = computed({
  get() {
    return settings?.value?.tus?.chunkSize
      ? formatBytes(settings?.value?.tus?.chunkSize)
      : "";
  },
  set(value: string) {
    // Use debouncing to allow the user to type freely without
    // interruption by the formatter
    // Clear the previous timeout if it exists
    if (debounceTimeout.value) {
      clearTimeout(debounceTimeout.value);
    }

    // Set a new timeout to apply the format after a short delay
    debounceTimeout.value = window.setTimeout(() => {
      if (settings.value) settings.value.tus.chunkSize = parseBytes(value);
    }, 1500);
  },
});

// Define funcs
const capitalize = (name: string, where: string | RegExp = "_") => {
  if (where === "caps") where = /(?=[A-Z])/;
  const split = name.split(where);
  name = "";

  for (let i = 0; i < split.length; i++) {
    name += split[i].charAt(0).toUpperCase() + split[i].slice(1) + " ";
  }

  return name.slice(0, -1);
};

const save = async () => {
  if (settings.value === null) return false;
  const newSettings: ISettings = {
    ...settings.value,
    shell:
      settings.value?.shell
        .join(" ")
        .trim()
        .split(" ")
        .filter((s: string) => s !== "") ?? [],
    commands: {},
  };

  const keys = Object.keys(settings.value.commands) as Array<
    keyof SettingsCommand
  >;
  for (const key of keys) {
    // not sure if we can safely assume non-null
    const newValue = commandObject.value[key];
    if (!newValue) continue;

    if (Array.isArray(newValue)) {
      newSettings.commands[key] = newValue;
    } else if (key in commandObject.value) {
      newSettings.commands[key] = newValue
        .split("\n")
        .filter((cmd: string) => cmd !== "");
    }
  }
  newSettings.shell = shellValue.value
    .trim()
    .split(" ")
    .filter((s) => s !== "");

  if (newSettings.branding.theme !== getTheme()) {
    setTheme(newSettings.branding.theme);
  }

  // defaultRoleID is not part of the ISettings type yet; attach it so it round
  // trips to the backend (Settings.DefaultRoleID).
  const payload = { ...newSettings, defaultRoleID: defaultRoleID.value };

  try {
    await api.update(payload);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }

  return true;
};
// Parse the user-friendly input (e.g., "20M" or "1T") to bytes
const parseBytes = (input: string) => {
  const regex = /^(\d+)(\.\d+)?(B|K|KB|M|MB|G|GB|T|TB)?$/i;
  const matches = input.match(regex);
  if (matches) {
    const size = parseFloat(matches[1].concat(matches[2] || ""));
    let unit: keyof SettingsUnit =
      matches[3].toUpperCase() as keyof SettingsUnit;
    if (!unit.endsWith("B")) {
      unit += "B";
    }
    const units: SettingsUnit = {
      KB: 1024,
      MB: 1024 ** 2,
      GB: 1024 ** 3,
      TB: 1024 ** 4,
    };
    return size * (units[unit as keyof SettingsUnit] || 1);
  } else {
    return 1024 ** 2;
  }
};
// Format the chunk size in bytes to user-friendly format
const formatBytes = (bytes: number) => {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size}${units[unitIndex]}`;
};

// Define Hooks

onMounted(async () => {
  try {
    layoutStore.loading = true;
    const original: ISettings = await api.get();
    const newSettings: ISettings = { ...original, commands: {} };

    const keys = Object.keys(original.commands) as Array<keyof SettingsCommand>;
    for (const key of keys) {
      newSettings.commands[key] = original.commands[key];
      commandObject.value[key] = original.commands[key]!.join("\n");
    }

    originalSettings.value = original;
    settings.value = newSettings;
    shellValue.value = newSettings.shell.join(" ");
    defaultRoleID.value =
      (original as { defaultRoleID?: number }).defaultRoleID ?? 0;

    try {
      roles.value = await rolesApi.list();
    } catch {
      roles.value = [];
    }
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});

// Clear the debounce timeout when the component is destroyed
onBeforeUnmount(() => {
  if (debounceTimeout.value) {
    clearTimeout(debounceTimeout.value);
  }
});
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
  line-height: 1.6;
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
.toc-hint {
  font-size: 12.5px;
  color: #9a9a9e;
  line-height: 1.6;
  margin: 0 0 10px;
}
.toc-input,
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
