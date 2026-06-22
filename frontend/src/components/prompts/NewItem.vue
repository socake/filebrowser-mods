<template>
  <div class="card floating new-item">
    <div class="card-title">
      <h2>{{ t("prompts.newItem") }}</h2>
    </div>

    <!-- Step 1: choose folder or file -->
    <div v-if="kind === null" class="card-content">
      <p>{{ t("prompts.newItemMessage") }}</p>
      <div class="ni-kinds">
        <button type="button" class="ni-kind" @click="kind = 'dir'">
          <i class="material-icons">create_new_folder</i>
          <span>{{ t("prompts.newItemFolder") }}</span>
        </button>
        <button type="button" class="ni-kind" @click="kind = 'file'">
          <i class="material-icons">note_add</i>
          <span>{{ t("prompts.newItemFile") }}</span>
        </button>
      </div>
    </div>

    <!-- Step 2a: folder name -->
    <div v-else-if="kind === 'dir'" class="card-content">
      <p>{{ t("prompts.newDirMessage") }}</p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
        tabindex="1"
      />
      <CreateFilePath :name="name" :is-dir="true" />
    </div>

    <!-- Step 2b: file type + name -->
    <div v-else class="card-content">
      <p>{{ t("prompts.newItemFileType") }}</p>
      <div class="ni-types">
        <button
          v-for="(ft, idx) in fileTypes"
          :key="ft.ext || 'custom'"
          type="button"
          class="ni-type"
          :class="{ 'ni-type--active': selectedIndex === idx }"
          @click="selectedIndex = idx"
        >
          <span class="ni-type__label">{{ ft.label }}</span>
          <span class="ni-type__ext">{{ ft.ext || t("prompts.newItemCustomExt") }}</span>
        </button>
      </div>

      <p style="margin-top: 1em">{{ t("prompts.newFileMessage") }}</p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
        :placeholder="selectedType.ext ? t('prompts.newItemNamePlaceholder') : t('prompts.newItemCustomPlaceholder')"
        tabindex="1"
      />
      <CreateFilePath :name="fileName" />
    </div>

    <div class="card-action">
      <button
        v-if="kind !== null"
        class="button button--flat button--grey"
        @click="back"
        :aria-label="t('buttons.back')"
        :title="t('buttons.back')"
        tabindex="3"
      >
        {{ t("buttons.back") }}
      </button>
      <button
        v-else
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
        tabindex="3"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        v-if="kind !== null"
        class="button button--flat"
        :aria-label="t('buttons.create')"
        :title="t('buttons.create')"
        @click="submit"
        tabindex="2"
      >
        {{ t("buttons.create") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { inject } from "vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { files as api } from "@/api";
import url from "@/utils/url";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import CreateFilePath from "@/components/prompts/CreateFilePath.vue";

const $showError = inject<IToastError>("$showError")!;

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

type FileType = { label: string; ext: string };

const fileTypes: FileType[] = [
  { label: t("prompts.fileTypes.txt"), ext: ".txt" },
  { label: "Markdown", ext: ".md" },
  { label: "JSON", ext: ".json" },
  { label: "YAML", ext: ".yaml" },
  { label: "CSV", ext: ".csv" },
  { label: "HTML", ext: ".html" },
  { label: "CSS", ext: ".css" },
  { label: "JavaScript", ext: ".js" },
  { label: "Python", ext: ".py" },
  { label: t("prompts.fileTypes.ini"), ext: ".ini" },
  { label: t("prompts.fileTypes.custom"), ext: "" },
];

const kind = ref<null | "dir" | "file">(null);
const name = ref<string>("");
const selectedIndex = ref<number>(0);
const selectedType = computed<FileType>(() => fileTypes[selectedIndex.value]);

// Final file name shown in the preview / used for creation.
const fileName = computed(() => {
  if (!name.value) return "";
  return name.value + (selectedType.value.ext || "");
});

const back = () => {
  kind.value = null;
  name.value = "";
  selectedIndex.value = 0;
};

// Resolve the base directory uri for the current /files listing.
const baseUri = () => {
  let uri = fileStore.isFiles ? route.path + "/" : "/";
  if (!fileStore.isListing) {
    uri = url.removeLastDir(uri) + "/";
  }
  return uri;
};

const createDir = async () => {
  let uri = baseUri();
  uri += encodeURIComponent(name.value) + "/";
  uri = uri.replace("//", "/");

  await api.post(uri);
  const res = await api.fetch(url.removeLastDir(uri) + "/");
  fileStore.updateRequest(res);
};

const createFile = async () => {
  let uri = baseUri();
  uri += encodeURIComponent(fileName.value);
  uri = uri.replace("//", "/");

  await api.post(uri);
  router.push({ path: uri });
};

const submit = async (event: Event) => {
  event.preventDefault();
  if (name.value === "") return;

  try {
    if (kind.value === "dir") {
      await createDir();
    } else if (kind.value === "file") {
      await createFile();
    }
  } catch (e) {
    if (e instanceof Error) {
      $showError(e);
    }
  }

  layoutStore.closeHovers();
};
</script>

<style scoped>
.new-item {
  max-width: 28rem;
}

.ni-kinds {
  display: flex;
  gap: 1em;
  margin-top: 0.5em;
}

.ni-kind {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5em;
  padding: 1.4em 1em;
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 0.6em;
  background: #fff;
  cursor: pointer;
  color: #333;
  transition: border-color 0.15s, box-shadow 0.15s, color 0.15s;
}

.ni-kind:hover {
  border-color: var(--lumen-accent);
  color: var(--lumen-accent);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.06);
}

.ni-kind i {
  font-size: 2em;
}

.ni-kind span {
  font-size: 0.95em;
}

.ni-types {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(7rem, 1fr));
  gap: 0.5em;
  margin-top: 0.5em;
}

.ni-type {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.15em;
  padding: 0.6em 0.7em;
  border: 1px solid rgba(0, 0, 0, 0.12);
  border-radius: 0.5em;
  background: #fff;
  cursor: pointer;
  text-align: left;
  transition: border-color 0.15s, background 0.15s, color 0.15s;
}

.ni-type:hover {
  border-color: var(--lumen-accent);
}

.ni-type--active {
  border-color: var(--lumen-accent);
  background: var(--lumen-accent);
  color: #fff;
}

.ni-type__label {
  font-size: 0.9em;
  font-weight: 500;
}

.ni-type__ext {
  font-size: 0.78em;
  opacity: 0.7;
}
</style>
