<template>
  <div class="card floating lumen-share" id="share">
    <!-- ===== 创建态 ===== -->
    <template v-if="state === 'create'">
      <div class="ls-head">
        <h2>{{ $t("prompts.shareTitle") }}</h2>
        <button class="ls-x" @click="closeHovers" :aria-label="$t('buttons.close')">
          <i class="material-icons">close</i>
        </button>
      </div>

      <div class="ls-body">
        <div class="ls-file" v-if="fileItem">
          <div class="ls-file-ic">{{ fileBadge }}</div>
          <div class="ls-file-nm">{{ fileItem.name }}</div>
          <div class="ls-file-mt">{{ fileMeta }}</div>
        </div>

        <div class="ls-label">{{ $t("prompts.shareMethod") }}</div>
        <div class="ls-ways">
          <div
            class="ls-way preview"
            :class="{ on: shareType === 'preview' }"
            @click="shareType = 'preview'"
          >
            <div class="ls-way-top">
              <div class="ls-way-ic"><i class="material-icons">visibility</i></div>
              <div class="ls-chk">
                <i v-if="shareType === 'preview'" class="material-icons">check</i>
              </div>
            </div>
            <div class="ls-way-ti">{{ $t("prompts.sharePreview") }}</div>
            <div class="ls-way-de">{{ $t("prompts.sharePreviewDesc") }}</div>
          </div>
          <div
            class="ls-way dl"
            :class="{ on: shareType === 'download' }"
            @click="shareType = 'download'"
          >
            <div class="ls-way-top">
              <div class="ls-way-ic"><i class="material-icons">file_download</i></div>
              <div class="ls-chk">
                <i v-if="shareType === 'download'" class="material-icons">check</i>
              </div>
            </div>
            <div class="ls-way-ti">{{ $t("prompts.shareDownload") }}</div>
            <div class="ls-way-de">{{ $t("prompts.shareDownloadDesc") }}</div>
          </div>
        </div>

        <div class="ls-label">{{ $t("settings.shareDuration") }}</div>
        <div class="ls-chips">
          <div
            v-for="opt in expireOptions"
            :key="opt.key"
            class="ls-chip"
            :class="{ on: expireChoice === opt.key }"
            @click="expireChoice = opt.key"
          >
            {{ opt.label }}
          </div>
        </div>

        <div v-if="expireChoice === 'custom'" class="ls-custom">
          <vue-number-input
            center
            controls
            size="small"
            :max="2147483647"
            :min="1"
            v-model="customTime"
          />
          <select class="ls-select" v-model="customUnit" :aria-label="$t('time.unit')">
            <option value="seconds">{{ $t("time.seconds") }}</option>
            <option value="minutes">{{ $t("time.minutes") }}</option>
            <option value="hours">{{ $t("time.hours") }}</option>
            <option value="days">{{ $t("time.days") }}</option>
          </select>
        </div>

        <div class="ls-label">{{ $t("prompts.shareAccessPassword") }}</div>
        <div class="ls-pw-row">
          <span class="ls-pw-t">{{ $t("prompts.sharePasswordToggle") }}</span>
          <div
            class="ls-toggle"
            :class="{ on: usePassword }"
            @click="usePassword = !usePassword"
            role="switch"
            :aria-checked="usePassword"
          >
            <i></i>
          </div>
        </div>
        <input
          v-if="usePassword"
          class="ls-pw-input"
          type="password"
          v-model.trim="password"
          :placeholder="$t('prompts.sharePasswordPlaceholder')"
        />
      </div>

      <div class="ls-foot">
        <button class="ls-btn ls-ghost" @click="closeHovers">
          {{ $t("buttons.cancel") }}
        </button>
        <button class="ls-btn ls-primary" @click="submit">
          {{ $t("prompts.shareCreate") }}
        </button>
      </div>
    </template>

    <!-- ===== 已创建态 ===== -->
    <template v-else>
      <div class="ls-head">
        <h2>{{ $t("prompts.shareCreatedTitle") }}</h2>
        <button class="ls-x" @click="closeHovers" :aria-label="$t('buttons.close')">
          <i class="material-icons">close</i>
        </button>
      </div>

      <div class="ls-body">
        <div class="ls-done-ic"><i class="material-icons">check</i></div>
        <div class="ls-done-t">{{ $t("prompts.shareLinkReady") }}</div>
        <div class="ls-done-s">
          {{ createdName }} ·
          {{
            created.type === "download"
              ? $t("prompts.shareDownload")
              : $t("prompts.sharePreview")
          }}
        </div>

        <div class="ls-link-box">
          <div class="ls-lt">
            <span class="ls-tag" :class="created.type === 'download' ? 'dl' : 'preview'">
              {{
                created.type === "download"
                  ? $t("prompts.shareTagDownload")
                  : $t("prompts.shareTagPreview")
              }}
            </span>
            <span class="ls-lt-sub">
              {{
                created.type === "download"
                  ? $t("prompts.shareDownloadLinkLabel")
                  : $t("prompts.sharePreviewLinkLabel")
              }}
            </span>
            <span class="ls-ex">{{ expireLabel }}</span>
          </div>
          <div class="ls-link-row">
            <span class="ls-u">{{ createdURL }}</span>
            <button class="ls-cp" @click="copyToClipboard(createdURL)">
              <i class="material-icons">content_copy</i>{{ $t("buttons.copyToClipboard") }}
            </button>
          </div>
        </div>

        <div class="ls-meta">
          {{
            created.password_hash
              ? $t("prompts.sharePasswordRequired")
              : $t("prompts.shareNoPassword")
          }}
        </div>
      </div>

      <div class="ls-foot">
        <button class="ls-btn ls-ghost" @click="createAnother">
          {{ $t("prompts.shareCreateAnother") }}
        </button>
        <button class="ls-btn ls-primary" @click="closeHovers">
          {{ $t("buttons.done") }}
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useFileStore } from "@/stores/file";
import * as api from "@/api/index";
import { filesize } from "@/utils";
import { useLayoutStore } from "@/stores/layout";
import { copyToClipboardWithFallback } from "@/utils/clipboard";
import { getExtBadge } from "@/utils/fileType";
import { formatTimestamp } from "@/utils/share";

export default {
  name: "share",
  data: function () {
    return {
      state: "create", // "create" | "created"
      shareType: "preview",
      expireChoice: "permanent",
      customTime: 7,
      customUnit: "days",
      usePassword: false,
      password: "",
      created: null,
      createdURL: "",
      createdName: "",
    };
  },
  inject: ["$showError", "$showSuccess"],
  computed: {
    ...mapState(useFileStore, ["req", "selected", "selectedCount", "isListing"]),
    expireOptions() {
      return [
        { key: "1", label: this.$t("prompts.shareExpire1Day") },
        { key: "7", label: this.$t("prompts.shareExpire7Days") },
        { key: "30", label: this.$t("prompts.shareExpire30Days") },
        { key: "permanent", label: this.$t("permanent") },
        { key: "custom", label: this.$t("prompts.shareExpireCustom") },
      ];
    },
    fileItem() {
      if (!this.isListing) {
        return this.req;
      }
      if (this.selectedCount === 1) {
        return this.req.items[this.selected[0]];
      }
      return null;
    },
    fileBadge() {
      const it = this.fileItem;
      if (!it) return "—";
      if (it.isDir) return "DIR";
      return getExtBadge(it.extension);
    },
    fileMeta() {
      const it = this.fileItem;
      if (!it) return "";
      if (it.isDir) return this.$t("prompts.shareFolderMeta");
      return filesize(it.size || 0);
    },
    expireLabel() {
      if (this.created && this.created.expire) {
        return this.$t("prompts.shareExpiresAt", {
          date: formatTimestamp(this.created.expire),
        });
      }
      return this.$t("permanent");
    },
    url() {
      if (!this.isListing) {
        return this.$route.path;
      }
      if (this.selectedCount === 0 || this.selectedCount > 1) {
        return;
      }
      return this.req.items[this.selected[0]].url;
    },
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    copyToClipboard: function (text) {
      copyToClipboardWithFallback(text).then(
        () => this.$showSuccess(this.$t("success.linkCopied")),
        (e) => this.$showError(e)
      );
    },
    expireParams() {
      switch (this.expireChoice) {
        case "permanent":
          return { expires: "", unit: "hours" };
        case "custom":
          return { expires: String(this.customTime || 0), unit: this.customUnit };
        default:
          // "1" / "7" / "30" days
          return { expires: this.expireChoice, unit: "days" };
      }
    },
    submit: async function () {
      try {
        const { expires, unit } = this.expireParams();
        const pw = this.usePassword ? this.password : "";
        const res = await api.share.create(
          this.url,
          pw,
          expires,
          unit,
          this.shareType
        );

        this.created = res;
        this.createdName = this.fileItem ? this.fileItem.name : res.path;
        this.createdURL = api.share.getShareURL(res);
        this.state = "created";
      } catch (e) {
        this.$showError(e);
      }
    },
    createAnother() {
      this.created = null;
      this.createdURL = "";
      this.password = "";
      this.usePassword = false;
      this.state = "create";
    },
  },
};
</script>

<style scoped>
.lumen-share {
  --ink: var(--lumen-accent);
  --t2: #5f5f63;
  --t3: #9a9a9e;
  --line: #ececee;
  --line2: #e2e2e5;
  --soft: #f3f3f4;
  --blue: #2f7bf6;
  --green: #1e9e5a;
  width: 400px;
  max-width: 94vw;
  padding: 0;
  border-radius: 18px;
  overflow: hidden;
  color: var(--ink);
}

.ls-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 22px 14px;
}
.ls-head h2 {
  font-size: 18px;
  font-weight: 750;
  letter-spacing: -0.01em;
  margin: 0;
}
.ls-x {
  width: 28px;
  height: 28px;
  border: 0;
  background: transparent;
  border-radius: 8px;
  color: var(--t3);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.ls-x:hover {
  background: var(--soft);
  color: var(--ink);
}
.ls-x .material-icons {
  font-size: 18px;
}

.ls-body {
  padding: 4px 22px 8px;
}

.ls-file {
  display: flex;
  align-items: center;
  gap: 10px;
  background: var(--soft);
  border-radius: 10px;
  padding: 10px 12px;
  margin-bottom: 8px;
}
.ls-file-ic {
  width: 30px;
  height: 30px;
  border-radius: 7px;
  background: #fde8e8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #e5484d;
  flex-shrink: 0;
}
.ls-file-nm {
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ls-file-mt {
  font-size: 12px;
  color: var(--t3);
  margin-left: auto;
  flex-shrink: 0;
}

.ls-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--t2);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 18px 0 9px;
}

.ls-ways {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.ls-way {
  border: 1.5px solid var(--line2);
  border-radius: 12px;
  padding: 13px;
  cursor: pointer;
  transition: 0.14s;
}
.ls-way:hover {
  border-color: #c8c8cc;
}
.ls-way.on {
  border-color: var(--ink);
  background: #fafafa;
}
.ls-way-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 7px;
}
.ls-way-ic {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.ls-way-ic .material-icons {
  font-size: 17px;
}
.ls-way.preview .ls-way-ic {
  background: var(--blue);
}
.ls-way.dl .ls-way-ic {
  background: var(--green);
}
.ls-chk {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1.5px solid var(--line2);
  display: flex;
  align-items: center;
  justify-content: center;
}
.ls-way.on .ls-chk {
  background: var(--ink);
  border-color: var(--ink);
}
.ls-chk .material-icons {
  font-size: 12px;
  color: #fff;
}
.ls-way-ti {
  font-size: 14px;
  font-weight: 650;
}
.ls-way-de {
  font-size: 11.5px;
  color: var(--t3);
  margin-top: 2px;
  line-height: 1.4;
}

.ls-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.ls-chip {
  padding: 7px 14px;
  border: 1px solid var(--line2);
  border-radius: 99px;
  font-size: 13px;
  font-weight: 600;
  color: var(--t2);
  cursor: pointer;
}
.ls-chip:hover {
  border-color: #c8c8cc;
}
.ls-chip.on {
  background: var(--ink);
  color: #fff;
  border-color: var(--ink);
}

.ls-custom {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 10px;
}
.ls-select {
  height: 34px;
  border: 1px solid var(--line2);
  border-radius: 8px;
  padding: 0 8px;
  background: #fff;
  color: var(--ink);
}

.ls-pw-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2px 0;
}
.ls-pw-t {
  font-size: 13.5px;
  color: var(--t2);
}
.ls-toggle {
  width: 40px;
  height: 23px;
  border-radius: 99px;
  background: #d8d8dc;
  position: relative;
  cursor: pointer;
  transition: 0.15s;
  flex-shrink: 0;
}
.ls-toggle.on {
  background: var(--ink);
}
.ls-toggle i {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 19px;
  height: 19px;
  border-radius: 50%;
  background: #fff;
  transition: 0.15s;
}
.ls-toggle.on i {
  left: 19px;
}
.ls-pw-input {
  width: 100%;
  margin-top: 10px;
  height: 38px;
  border: 1px solid var(--line2);
  border-radius: 10px;
  padding: 0 12px;
  font-size: 14px;
  box-sizing: border-box;
}

.ls-foot {
  display: flex;
  gap: 10px;
  padding: 16px 22px 20px;
}
.ls-btn {
  flex: 1;
  height: 42px;
  border-radius: 11px;
  font-size: 14px;
  font-weight: 650;
  cursor: pointer;
  border: 1px solid transparent;
}
.ls-ghost {
  background: #fff;
  border-color: var(--line2);
  color: var(--t2);
}
.ls-ghost:hover {
  border-color: var(--ink);
  color: var(--ink);
}
.ls-primary {
  background: var(--ink);
  color: #fff;
}

/* 已创建态 */
.ls-done-ic {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: #eafaf0;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 6px auto 12px;
}
.ls-done-ic .material-icons {
  color: var(--green);
  font-size: 24px;
}
.ls-done-t {
  text-align: center;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 4px;
}
.ls-done-s {
  text-align: center;
  font-size: 13px;
  color: var(--t3);
  margin-bottom: 18px;
}
.ls-link-box {
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px 14px;
  margin-bottom: 12px;
}
.ls-lt {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-bottom: 8px;
}
.ls-tag {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 9px;
  border-radius: 99px;
}
.ls-tag.preview {
  background: rgba(47, 123, 246, 0.12);
  color: #1f5fd0;
}
.ls-tag.dl {
  background: rgba(30, 158, 90, 0.14);
  color: #15834a;
}
.ls-lt-sub {
  font-size: 12.5px;
  color: var(--t2);
}
.ls-ex {
  font-size: 11px;
  color: var(--t3);
  margin-left: auto;
}
.ls-link-row {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--soft);
  border-radius: 8px;
  padding: 8px 10px;
}
.ls-u {
  flex: 1;
  font-size: 12.5px;
  color: var(--t2);
  font-family: ui-monospace, Menlo, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ls-cp {
  display: flex;
  align-items: center;
  gap: 5px;
  background: var(--ink);
  color: #fff;
  border: 0;
  border-radius: 7px;
  padding: 6px 11px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  flex-shrink: 0;
}
.ls-cp .material-icons {
  font-size: 13px;
}
.ls-meta {
  font-size: 12px;
  color: var(--t3);
  text-align: center;
  margin: 8px 0 4px;
}
</style>
