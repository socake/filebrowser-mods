<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <form @submit="save" class="card">
        <div class="card-title">
          <h2 v-if="user?.id === 0">{{ $t("settings.newUser") }}</h2>
          <h2 v-else>{{ $t("settings.user") }} {{ user?.username }}</h2>
        </div>

        <div class="card-content" v-if="user">
          <p v-if="roles.length">
            <label for="role">{{ t("roles.assignRole") }}</label>
            <select
              id="role"
              class="input input--block"
              v-model.number="user.roleID"
              @change="applyRole"
            >
              <option :value="0">{{ t("roles.noRole") }}</option>
              <option v-for="r in roles" :key="r.id" :value="r.id">
                {{ r.name }}
              </option>
            </select>
            <span class="small">{{ t("roles.assignRoleHint") }}</span>
          </p>

          <user-form
            v-model:user="user"
            v-model:createUserDir="createUserDir"
            :isDefault="false"
            :isNew="isNew"
          />
        </div>

        <div class="card-action">
          <button
            v-if="!isNew"
            @click.prevent="deletePrompt"
            type="button"
            class="button button--flat button--red"
            :aria-label="$t('buttons.delete')"
            :title="$t('buttons.delete')"
          >
            {{ $t("buttons.delete") }}
          </button>
          <router-link to="/users">
            <button
              class="button button--flat button--grey"
              :aria-label="$t('buttons.cancel')"
              :title="$t('buttons.cancel')"
            >
              {{ $t("buttons.cancel") }}
            </button>
          </router-link>
          <input
            class="button button--flat"
            type="submit"
            :value="$t('buttons.save')"
          />
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { users as api, roles as rolesApi, settings } from "@/api";
import UserForm from "@/components/settings/UserForm.vue";
import Errors from "@/views/Errors.vue";
import { computed, inject, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { authMethod } from "@/utils/constants";
import { logout } from "@/utils/auth";

const error = ref<StatusError>();
const originalUser = ref<IUser>();
const user = ref<IUser>();
const createUserDir = ref<boolean>(false);
const isCurrentPasswordRequired = ref<boolean>(false);
const roles = ref<Role[]>([]);

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();

onMounted(() => {
  fetchData();
});

const isNew = computed(() => route.path === "/users/new");

watch(route, () => fetchData());
watch(user, () => {
  if (!user.value?.perm.admin) return;
  user.value.lockPassword = false;
});

const fetchData = async () => {
  layoutStore.loading = true;

  try {
    try {
      roles.value = await rolesApi.list();
      roles.value.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
    } catch {
      roles.value = [];
    }

    if (isNew.value) {
      const { defaults, createUserDir: _createUserDir } = await settings.get();
      isCurrentPasswordRequired.value = authMethod == "json";
      createUserDir.value = _createUserDir;
      user.value = {
        ...defaults,
        username: "",
        password: "",
        rules: [],
        lockPassword: false,
        id: 0,
        roleID: 0,
      };
    } else {
      const { authMethod } = await settings.get();
      isCurrentPasswordRequired.value = authMethod == "json";
      const id = Array.isArray(route.params.id)
        ? route.params.id.join("")
        : route.params.id;
      user.value = { ...(await api.get(parseInt(id))) };
      if (user.value.roleID == null) user.value.roleID = 0;
    }
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

const applyRole = () => {
  if (!user.value) return;
  const role = roles.value.find((r) => r.id === user.value!.roleID);
  if (!role) return;
  // 把角色权限填入页面权限勾选区，让管理员看到套用效果；
  // 用户仍可在下方手动改个别权限字段（= 个人覆盖）。
  user.value.perm = {
    ...user.value.perm,
    admin: role.permissions.admin,
    create: role.permissions.create,
    rename: role.permissions.rename,
    modify: role.permissions.modify,
    delete: role.permissions.delete,
    share: role.permissions.share,
    download: role.permissions.download,
    execute: role.permissions.execute,
  };
};

const deletePrompt = () => {
  if (isCurrentPasswordRequired.value) {
    layoutStore.showHover({
      prompt: "current-password",
      confirm: (event: Event, currentPassword: string) => {
        event.preventDefault();
        layoutStore.closeHovers();
        deleteUser(currentPassword);
      },
    });
  } else {
    layoutStore.showHover({
      prompt: "deleteUser",
      confirm: () => deleteUser(""),
    });
  }
};

const deleteUser = async (currentPassword: string) => {
  if (!user.value) {
    return false;
  }
  try {
    await api.remove(user.value.id, currentPassword);
    if (user.value.id == authStore.user?.id) {
      logout();
    } else {
      router.push({ path: "/users" });
    }
    $showSuccess(t("settings.userDeleted"));
  } catch (err) {
    if (err instanceof StatusError) {
      err.status === 403 ? $showError(t("errors.forbidden")) : $showError(err);
    } else if (err instanceof Error) {
      $showError(err);
    }
  }

  return true;
};

const save = (event: Event) => {
  event.preventDefault();
  if (isCurrentPasswordRequired.value) {
    layoutStore.showHover({
      prompt: "current-password",
      confirm: (event: Event, currentPassword: string) => {
        event.preventDefault();
        layoutStore.closeHovers();
        send(currentPassword);
      },
    });
  } else {
    send("");
  }

  return true;
};

const send = async (currentPassword: string) => {
  if (!user.value) {
    return false;
  }

  try {
    if (isNew.value) {
      const newUser: IUser = {
        ...originalUser?.value,
        ...user.value,
      };

      const loc = await api.create(newUser, currentPassword);
      router.push({ path: loc || "/users" });
      $showSuccess(t("settings.userCreated"));
    } else {
      await api.update(user.value, ["all"], currentPassword);

      if (user.value.id === authStore.user?.id) {
        authStore.updateUser(user.value);
      }

      $showSuccess(t("settings.userUpdated"));
    }
  } catch (e: any) {
    $showError(e);
  }
};
</script>

<style scoped>
/* To C 卡片化：表单收窄居中，不再全宽撑满（不影响 /files 全宽列表） */
.row {
  display: block;
  max-width: 720px;
  margin: 0 auto;
  padding: 28px 24px;
}
.row .column {
  display: block;
  width: 100%;
  padding: 0;
}
.row .card {
  margin: 0;
  background: #fff;
  border: 1px solid #ececee;
  border-radius: 16px;
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.04),
    0 8px 24px rgba(0, 0, 0, 0.05);
  overflow: visible;
}
.row .card .card-title {
  padding: 22px 26px 4px;
  display: block;
}
.row .card .card-title h2 {
  font-size: 21px;
  font-weight: 760;
  letter-spacing: -0.01em;
  color: var(--lumen-accent);
  margin: 0;
}
.row .card .card-content {
  padding: 10px 26px 6px;
}
.row .card .card-action {
  padding: 14px 26px 22px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
}

/* 表单标签 / 段落（含子组件 UserForm 的输入项） */
.card-content :deep(p) {
  margin: 0 0 16px;
}
.card-content :deep(label) {
  display: block;
  font-size: 12.5px;
  font-weight: 600;
  color: #5f5f63;
  margin: 0 0 6px;
}
.card-content :deep(.small) {
  color: #9a9a9e;
}

/* 输入框：收窄，不再 100% 撑满 */
.card-content :deep(.input--block) {
  width: 100%;
  max-width: 440px;
  min-height: 40px;
  border: 1px solid #e2e2e5;
  border-radius: 10px;
  padding: 0 12px;
  background: #fff;
  box-sizing: border-box;
}
.card-content :deep(textarea.input--block) {
  padding: 8px 12px;
  max-width: 100%;
}
.card-content :deep(.input--block:focus) {
  border-color: var(--lumen-accent);
}
</style>
