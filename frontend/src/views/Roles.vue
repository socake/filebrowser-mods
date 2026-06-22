<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <main class="roles-main">
      <div class="head">
        <div>
          <h1>{{ t("roles.title") }}</h1>
          <div class="sub">{{ t("roles.subtitle") }}</div>
        </div>
        <button class="btn" @click="openCreate">
          <i class="material-icons">add</i>{{ t("roles.newRole") }}
        </button>
      </div>

      <!-- 角色卡片网格 -->
      <div class="roles-grid">
        <div
          v-for="role in roles"
          :key="role.id"
          class="role"
          :class="{ on: selectedId === role.id, preset: role.isPreset }"
          @click="selectRole(role)"
        >
          <span v-if="role.isPreset" class="badge">{{ t("roles.preset") }}</span>
          <div class="ri" :style="{ background: roleColor(role) }">
            <i class="material-icons">{{ roleIcon(role) }}</i>
          </div>
          <div class="rn">{{ role.name }}</div>
          <div class="rd">{{ role.description }}</div>
          <div class="ru">
            <i class="material-icons">folder</i>{{ role.scope || "/" }}
          </div>
        </div>
      </div>

      <!-- 权限配置面板 -->
      <div v-if="edit" class="panel">
        <div class="panel-head">
          <div class="pt">
            <span class="dot" :style="{ background: roleColor(edit) }">
              <i class="material-icons">{{ roleIcon(edit) }}</i>
            </span>
            {{ edit.name }} · {{ t("roles.permConfig") }}
            <span v-if="edit.isPreset" class="badge">{{ t("roles.preset") }}</span>
          </div>
          <div class="pa">
            <button
              v-if="!edit.isPreset"
              class="danger"
              @click="removeRole"
            >
              {{ t("roles.delete") }}
            </button>
            <button @click="openRename">{{ t("roles.rename") }}</button>
            <button class="save" @click="saveRole">{{ t("roles.save") }}</button>
          </div>
        </div>

        <div class="perm-label">{{ t("roles.permissions") }}</div>
        <div class="perms">
          <div v-for="p in permDefs" :key="p.key" class="perm">
            <div class="pl">
              <div class="pi"><i class="material-icons">{{ p.icon }}</i></div>
              <div>
                <div class="pn">{{ t(`roles.perm.${p.key}`) }}</div>
                <div class="pdesc">{{ t(`roles.perm.${p.key}Desc`) }}</div>
              </div>
            </div>
            <div
              class="toggle"
              :class="{ on: edit.permissions[p.key] }"
              @click="togglePerm(p.key)"
            >
              <i></i>
            </div>
          </div>
        </div>

        <div class="scope-row">
          <span class="sl">{{ t("roles.scope") }}</span>
          <input
            v-model="edit.scope"
            class="sv"
            type="text"
            :placeholder="t('roles.scopeRoot')"
          />
        </div>
      </div>
    </main>

    <!-- 新建 / 重命名 弹窗 -->
    <div v-if="modal" class="modal-mask" @click.self="closeModal">
      <div class="modal">
        <h2>{{ modal === "create" ? t("roles.newRole") : t("roles.rename") }}</h2>

        <label>{{ t("roles.roleName") }}</label>
        <input v-model="form.name" class="m-input" type="text" />

        <label>{{ t("roles.description") }}</label>
        <input v-model="form.description" class="m-input" type="text" />

        <template v-if="modal === 'create'">
          <label>{{ t("roles.scope") }}</label>
          <input
            v-model="form.scope"
            class="m-input"
            type="text"
            :placeholder="t('roles.scopeRoot')"
          />

          <div class="m-perm-label">{{ t("roles.permissions") }}</div>
          <div class="m-perms">
            <label
              v-for="p in permDefs"
              :key="p.key"
              class="m-perm"
            >
              <input type="checkbox" v-model="form.permissions[p.key]" />
              {{ t(`roles.perm.${p.key}`) }}
            </label>
          </div>
        </template>

        <div class="m-actions">
          <button class="m-cancel" @click="closeModal">
            {{ t("roles.cancel") }}
          </button>
          <button class="m-ok" @click="submitModal">
            {{ modal === "create" ? t("roles.create") : t("roles.save") }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import HeaderBar from "@/components/header/HeaderBar.vue";
import { roles as api } from "@/api";
import { StatusError } from "@/api/utils";
import { inject, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

type PermKey = keyof RolePermissions;

const permDefs: { key: PermKey; icon: string }[] = [
  { key: "admin", icon: "shield" },
  { key: "create", icon: "add" },
  { key: "modify", icon: "edit" },
  { key: "delete", icon: "delete" },
  { key: "rename", icon: "drive_file_rename_outline" },
  { key: "share", icon: "share" },
  { key: "download", icon: "download" },
  { key: "execute", icon: "terminal" },
];

const palette = [
  "var(--lumen-accent)",
  "#2f7bf6",
  "#1e9e5a",
  "#9a9a9e",
  "#f59e0b",
  "#8b5cf6",
  "#ef4444",
  "#0ea5e9",
];

const roles = ref<Role[]>([]);
const selectedId = ref<number | null>(null);
const edit = ref<Role | null>(null);

const modal = ref<"" | "create" | "rename">("");
const form = reactive<{
  name: string;
  description: string;
  scope: string;
  permissions: RolePermissions;
}>({
  name: "",
  description: "",
  scope: "",
  permissions: emptyPerms(),
});

function emptyPerms(): RolePermissions {
  return {
    admin: false,
    create: false,
    rename: false,
    modify: false,
    delete: false,
    share: false,
    download: false,
    execute: false,
  };
}

function roleColor(role: Role): string {
  return palette[(role.sort ?? role.id) % palette.length];
}

function roleIcon(role: Role): string {
  const p = role.permissions;
  if (p.admin) return "shield";
  if (p.create || p.modify || p.delete) return "edit";
  if (p.download) return "visibility";
  return "person_outline";
}

const fetchData = async () => {
  try {
    const list = await api.list();
    list.sort((a, b) => (a.sort ?? 0) - (b.sort ?? 0));
    roles.value = list;
    if (list.length) {
      const keep =
        selectedId.value != null
          ? list.find((r) => r.id === selectedId.value)
          : null;
      selectRole(keep ?? list[0]);
    } else {
      selectedId.value = null;
      edit.value = null;
    }
  } catch (e: any) {
    $showError(e);
  }
};

onMounted(fetchData);

function selectRole(role: Role) {
  selectedId.value = role.id;
  // 深拷贝出可编辑副本
  edit.value = {
    ...role,
    permissions: { ...role.permissions },
  };
}

function togglePerm(key: PermKey) {
  if (!edit.value) return;
  const next = !edit.value.permissions[key];
  edit.value.permissions[key] = next;
  // 选中管理员时联动开启全部，关闭时仅关闭管理员
  if (key === "admin" && next) {
    for (const p of permDefs) edit.value.permissions[p.key] = true;
  }
}

async function saveRole() {
  if (!edit.value) return;
  try {
    await api.update(edit.value);
    $showSuccess(t("roles.updated"));
    await fetchData();
  } catch (e: any) {
    $showError(e);
  }
}

async function removeRole() {
  if (!edit.value) return;
  if (!window.confirm(t("roles.deleteConfirm"))) return;
  try {
    await api.remove(edit.value.id);
    $showSuccess(t("roles.deleted"));
    selectedId.value = null;
    await fetchData();
  } catch (e: any) {
    if (e instanceof StatusError && e.status === 403) {
      $showError(t("roles.presetCannotDelete"));
    } else {
      $showError(e);
    }
  }
}

function openCreate() {
  modal.value = "create";
  form.name = "";
  form.description = "";
  form.scope = "";
  form.permissions = emptyPerms();
}

function openRename() {
  if (!edit.value) return;
  modal.value = "rename";
  form.name = edit.value.name;
  form.description = edit.value.description;
}

function closeModal() {
  modal.value = "";
}

async function submitModal() {
  if (!form.name.trim()) {
    $showError(t("roles.nameRequired"));
    return;
  }
  try {
    if (modal.value === "create") {
      const loc = await api.create({
        name: form.name.trim(),
        description: form.description,
        scope: form.scope,
        permissions: { ...form.permissions },
        isPreset: false,
        sort: roles.value.length,
      });
      $showSuccess(t("roles.created"));
      // Location 形如 /api/roles/{id}
      const newId = loc ? parseInt(loc.split("/").pop() || "", 10) : NaN;
      if (!Number.isNaN(newId)) selectedId.value = newId;
    } else if (modal.value === "rename" && edit.value) {
      const updated: Role = {
        ...edit.value,
        name: form.name.trim(),
        description: form.description,
      };
      await api.update(updated);
      $showSuccess(t("roles.updated"));
      selectedId.value = updated.id;
    }
    closeModal();
    await fetchData();
  } catch (e: any) {
    $showError(e);
  }
}
</script>

<style scoped>
.roles-main {
  --ink: var(--lumen-accent);
  --t2: #5f5f63;
  --t3: #9a9a9e;
  --line: #ececee;
  --line2: #e2e2e5;
  --soft: #f3f3f4;
  max-width: 1120px;
  margin: 0 auto;
  padding: 28px 34px;
  color: var(--ink);
}
.head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 22px;
}
.head h1 {
  font-size: 25px;
  font-weight: 780;
  letter-spacing: -0.02em;
  margin: 0;
}
.sub {
  font-size: 13.5px;
  color: var(--t3);
  margin-top: 4px;
}
.btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 38px;
  padding: 0 16px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  background: var(--ink);
  color: #fff;
  cursor: pointer;
  border: 0;
}
.btn .material-icons {
  font-size: 18px;
}

/* 角色卡片网格 */
.roles-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin-bottom: 26px;
}
@media (max-width: 900px) {
  .roles-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
.role {
  position: relative;
  background: #fff;
  border: 1.5px solid var(--line);
  border-radius: 14px;
  padding: 16px;
  cursor: pointer;
  transition: 0.15s;
}
.role:hover {
  border-color: var(--line2);
}
.role.on {
  border-color: var(--ink);
  box-shadow: 0 0 0 3px rgba(20, 20, 20, 0.06);
}
.role .ri {
  width: 38px;
  height: 38px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  margin-bottom: 11px;
}
.role .ri .material-icons {
  font-size: 20px;
  color: #fff;
}
.role .rn {
  font-size: 15px;
  font-weight: 700;
}
.role .rd {
  font-size: 12px;
  color: var(--t3);
  margin-top: 3px;
  line-height: 1.45;
  min-height: 34px;
}
.role .ru {
  font-size: 12px;
  color: var(--t2);
  font-weight: 600;
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.role .ru .material-icons {
  font-size: 14px;
  color: var(--t3);
}
.role .badge {
  position: absolute;
  top: 14px;
  right: 14px;
  font-size: 10px;
  font-weight: 700;
  background: var(--soft);
  color: var(--t3);
  padding: 1px 7px;
  border-radius: 99px;
}

/* 权限矩阵面板 */
.panel {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 14px;
  overflow: hidden;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--line);
  flex-wrap: wrap;
  gap: 10px;
}
.panel-head .pt {
  font-size: 16px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 9px;
}
.panel-head .pt .dot {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.panel-head .pt .dot .material-icons {
  font-size: 15px;
  color: #fff;
}
.panel-head .pt .badge {
  font-size: 10px;
  font-weight: 700;
  background: var(--soft);
  color: var(--t3);
  padding: 1px 7px;
  border-radius: 99px;
}
.panel-head .pa {
  display: flex;
  gap: 8px;
}
.panel-head .pa button {
  height: 32px;
  padding: 0 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--line2);
  background: #fff;
  color: var(--t2);
}
.panel-head .pa button:hover {
  border-color: var(--ink);
  color: var(--ink);
}
.panel-head .pa .save {
  background: var(--ink);
  color: #fff;
  border-color: var(--ink);
}
.panel-head .pa .danger {
  color: #ef4444;
  border-color: #f4c6c6;
}
.panel-head .pa .danger:hover {
  border-color: #ef4444;
  color: #ef4444;
}

.perm-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--t3);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 16px 20px 4px;
}
.perms {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px 24px;
  padding: 6px 20px 18px;
}
@media (max-width: 700px) {
  .perms {
    grid-template-columns: 1fr;
  }
}
.perm {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 11px 0;
  border-bottom: 1px solid var(--soft);
}
.perm .pl {
  display: flex;
  align-items: center;
  gap: 11px;
}
.perm .pi {
  width: 32px;
  height: 32px;
  border-radius: 9px;
  background: var(--soft);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--t2);
}
.perm .pi .material-icons {
  font-size: 17px;
  color: var(--t2);
}
.perm .pn {
  font-size: 14px;
  font-weight: 600;
}
.perm .pdesc {
  font-size: 11.5px;
  color: var(--t3);
}
.toggle {
  width: 40px;
  height: 23px;
  border-radius: 99px;
  background: #d8d8dc;
  position: relative;
  cursor: pointer;
  transition: 0.15s;
  flex-shrink: 0;
}
.toggle.on {
  background: var(--ink);
}
.toggle i {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 19px;
  height: 19px;
  border-radius: 50%;
  background: #fff;
  transition: 0.15s;
}
.toggle.on i {
  left: 19px;
}
.scope-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid var(--line);
  background: #fafafa;
}
.scope-row .sl {
  font-size: 13.5px;
  color: var(--t2);
  font-weight: 600;
}
.scope-row .sv {
  font-size: 13px;
  color: var(--ink);
  font-family: ui-monospace, Menlo, monospace;
  background: #fff;
  border: 1px solid var(--line2);
  border-radius: 8px;
  padding: 6px 12px;
  min-width: 220px;
}

/* 弹窗 */
.modal-mask {
  --ink: var(--lumen-accent);
  --t2: #5f5f63;
  --t3: #9a9a9e;
  --line: #ececee;
  --line2: #e2e2e5;
  --soft: #f3f3f4;
  position: fixed;
  inset: 0;
  background: rgba(20, 20, 20, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.modal {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  width: 440px;
  max-width: calc(100vw - 32px);
  max-height: calc(100vh - 64px);
  overflow: auto;
  color: var(--ink);
}
.modal h2 {
  font-size: 19px;
  font-weight: 760;
  margin: 0 0 16px;
}
.modal label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--t2);
  margin: 12px 0 6px;
}
.m-input {
  width: 100%;
  height: 38px;
  border: 1px solid var(--line2);
  border-radius: 9px;
  padding: 0 12px;
  font-size: 14px;
  color: var(--ink);
  outline: none;
}
.m-input:focus {
  border-color: var(--ink);
}
.m-perm-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--t3);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 18px 0 8px;
}
.m-perms {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 14px;
}
.m-perm {
  display: flex !important;
  align-items: center;
  gap: 7px;
  font-size: 13.5px !important;
  font-weight: 500 !important;
  color: var(--ink) !important;
  margin: 0 !important;
  cursor: pointer;
}
.m-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 22px;
}
.m-actions button {
  height: 38px;
  padding: 0 18px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}
.m-cancel {
  background: #fff;
  border: 1px solid var(--line2);
  color: var(--t2);
}
.m-ok {
  background: var(--ink);
  border: 1px solid var(--ink);
  color: #fff;
}
</style>
