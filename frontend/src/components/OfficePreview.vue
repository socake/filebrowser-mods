<template>
  <div class="office-preview">
    <!-- Loading state -->
    <div v-if="loading" class="office-state">
      <span class="office-spinner" aria-hidden="true"></span>
      <p>{{ loadingText }}</p>
    </div>

    <!-- Error state -->
    <div v-else-if="errorMsg" class="office-state office-state--error">
      <span class="office-icon">!</span>
      <p>{{ errorMsg }}</p>
    </div>

    <!-- Unsupported extension -->
    <div v-else-if="!officeComponent" class="office-state">
      <span class="office-icon">?</span>
      <p>{{ unsupportedText }}</p>
    </div>

    <!-- Document, kept mounted while rendering so it can emit @rendered -->
    <component
      :is="officeComponent"
      v-if="officeComponent && src"
      v-show="!loading && !errorMsg"
      :key="src + ':' + normalizedExt"
      :src="src"
      class="office-doc"
      @rendered="onRendered"
      @error="onError"
    />
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  shallowRef,
  watch,
  type Component,
} from "vue";

interface Props {
  /** Document URL. The underlying vue-office component fetches the ArrayBuffer. */
  src: string;
  /** File extension, e.g. "docx", "xlsx", "xls", "pptx" (case-insensitive, leading dot ok). */
  ext: string;
}

const props = defineProps<Props>();

const loading = shallowRef(true);
const errorMsg = shallowRef<string>("");

const loadingText = "Loading document…";
const unsupportedText = "Unsupported document type";

const normalizedExt = computed(() =>
  (props.ext || "").toLowerCase().replace(/^\./, "").trim()
);

// Lazily load the matching vue-office component (and its stylesheet).
function buildComponent(ext: string): Component | null {
  switch (ext) {
    case "docx":
    case "doc":
      return defineAsyncComponent(async () => {
        await import("@vue-office/docx/lib/index.css");
        return (await import("@vue-office/docx")).default as Component;
      });
    case "xlsx":
    case "xls":
    case "xlsm":
    case "csv":
      return defineAsyncComponent(async () => {
        await import("@vue-office/excel/lib/index.css");
        return (await import("@vue-office/excel")).default as Component;
      });
    case "pptx":
    case "ppt":
      // pptx package ships no stylesheet
      return defineAsyncComponent(
        async () => (await import("@vue-office/pptx")).default as Component
      );
    default:
      return null;
  }
}

const officeComponent = shallowRef<Component | null>(null);

watch(
  () => [normalizedExt.value, props.src] as const,
  () => {
    errorMsg.value = "";
    const comp = buildComponent(normalizedExt.value);
    officeComponent.value = comp;
    if (!comp) {
      loading.value = false;
    } else if (!props.src) {
      loading.value = false;
      errorMsg.value = "No document source provided";
    } else {
      loading.value = true;
    }
  },
  { immediate: true }
);

function onRendered() {
  loading.value = false;
  errorMsg.value = "";
}

function onError(e?: unknown) {
  loading.value = false;
  const detail =
    e instanceof Error
      ? e.message
      : typeof e === "string"
        ? e
        : "Failed to render document";
  errorMsg.value = detail;
  // eslint-disable-next-line no-console
  console.error("[OfficePreview] render error:", e);
}
</script>

<style scoped>
.office-preview {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 240px;
  overflow: auto;
  background: #ffffff;
  color: #141414;
  box-sizing: border-box;
}

.office-doc {
  width: 100%;
  height: 100%;
}

.office-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 2rem;
  text-align: center;
  background: #ffffff;
  color: #141414;
}

.office-state p {
  margin: 0;
  font-size: 1rem;
  opacity: 0.85;
}

.office-state--error {
  color: #b00020;
}

.office-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  font-size: 1.6rem;
  font-weight: 700;
  background: #141414;
  color: #ffffff;
}

.office-state--error .office-icon {
  background: #b00020;
}

.office-spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid rgba(20, 20, 20, 0.15);
  border-top-color: #141414;
  border-radius: 50%;
  animation: office-spin 0.8s linear infinite;
}

@keyframes office-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
