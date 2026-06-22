<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <errors v-if="error" :errorCode="error.status" />

    <main v-else class="users-main">
      <div class="head">
        <div>
          <h1>{{ t("settings.users") }}</h1>
          <div class="sub">共 {{ users.length }} 个用户</div>
        </div>
        <router-link to="/users/new">
          <button class="btn">
            <i class="material-icons">add</i>{{ t("buttons.new") }}
          </button>
        </router-link>
      </div>

      <div class="card">
        <table v-if="users.length > 0">
          <thead>
            <tr>
              <th>{{ t("settings.username") }}</th>
              <th class="w-role">角色</th>
              <th class="w-admin">{{ t("settings.admin") }}</th>
              <th>{{ t("settings.scope") }}</th>
              <th class="w-act">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>
                <div class="uname">
                  <span class="av">{{ initial(user.username) }}</span>
                  <span class="n">{{ user.username }}</span>
                </div>
              </td>
              <td class="muted">{{ roleName(user) }}</td>
              <td>
                <span v-if="user.perm.admin" class="tag admin">
                  {{ t("settings.admin") }}
                </span>
                <span v-else class="muted">—</span>
              </td>
              <td class="scope">{{ user.scope || "/" }}</td>
              <td>
                <router-link :to="'/users/' + user.id">
                  <button class="edit" title="编辑">
                    <i class="material-icons">mode_edit</i>
                  </button>
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty">
          <i class="material-icons">person_off</i>
          <span>暂无用户</span>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { users as api, roles as rolesApi } from "@/api";
import Errors from "@/views/Errors.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<StatusError | null>(null);
const users = ref<IUser[]>([]);
const roleMap = ref<Map<number, string>>(new Map());

const layoutStore = useLayoutStore();
const { t } = useI18n();

onMounted(async () => {
  layoutStore.loading = true;

  try {
    // 角色名映射：拿得到就显示角色名；拿不到则降级为「自定义 / —」
    try {
      const list = await rolesApi.list();
      const map = new Map<number, string>();
      for (const r of list) map.set(r.id, r.name);
      roleMap.value = map;
    } catch {
      roleMap.value = new Map();
    }

    users.value = await api.getAll();
  } catch (err) {
    if (err instanceof Error) {
      error.value = err as StatusError;
    }
  } finally {
    layoutStore.loading = false;
  }
});

const initial = (name: string): string =>
  (name || "?").trim().charAt(0).toUpperCase() || "?";

const roleName = (user: IUser): string => {
  const id = user.roleID;
  if (!id) return "—";
  return roleMap.value.get(id) ?? "自定义";
};
</script>

<style scoped>
.users-main {
  --ink: var(--lumen-accent);
  --t2: #5f5f63;
  --t3: #9a9a9e;
  --line: #ececee;
  --line2: #e2e2e5;
  --soft: #f3f3f4;
  --green: #1e9e5a;
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

.card {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 14px;
  overflow: hidden;
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
thead {
  background: #fafafa;
}
th {
  text-align: left;
  font-size: 12px;
  font-weight: 700;
  color: var(--t3);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  padding: 13px 16px;
}
td {
  padding: 13px 16px;
  border-top: 1px solid var(--line);
  color: var(--ink);
  vertical-align: middle;
}
.w-role {
  width: 180px;
}
.w-admin {
  width: 120px;
}
.w-act {
  width: 90px;
}

.uname {
  display: flex;
  align-items: center;
  gap: 11px;
}
.uname .av {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--soft);
  color: var(--t2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  flex-shrink: 0;
}
.uname .n {
  font-weight: 600;
}

.tag {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 99px;
}
.tag.admin {
  background: rgba(30, 158, 90, 0.14);
  color: #15834a;
}

.muted {
  color: var(--t3);
}
.scope {
  font-family: ui-monospace, Menlo, monospace;
  font-size: 13px;
  color: var(--t2);
}

.edit {
  width: 32px;
  height: 32px;
  border: 0;
  background: transparent;
  border-radius: 8px;
  color: var(--t3);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.edit .material-icons {
  font-size: 18px;
}
.edit:hover {
  background: var(--soft);
  color: var(--ink);
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 56px 24px;
  color: var(--t3);
}
.empty .material-icons {
  font-size: 40px;
}
</style>
