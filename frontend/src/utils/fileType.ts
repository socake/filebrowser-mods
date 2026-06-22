// File type / badge helpers shared by the share components.
//
// The extension -> category map below is the single source of truth (previously
// duplicated inline in views/SharesManage.vue). All category-based logic derives
// from it.

export const EXT_MAP: Record<string, string> = {
  png: "image",
  jpg: "image",
  jpeg: "image",
  gif: "image",
  webp: "image",
  bmp: "image",
  svg: "image",
  ico: "image",
  tiff: "image",
  heic: "image",
  pdf: "pdf",
  doc: "office",
  docx: "office",
  xls: "office",
  xlsx: "office",
  ppt: "office",
  pptx: "office",
  txt: "document",
  md: "document",
  rtf: "document",
  odt: "document",
  csv: "document",
  zip: "archive",
  rar: "archive",
  "7z": "archive",
  tar: "archive",
  gz: "archive",
  bz2: "archive",
  xz: "archive",
};

/** Last path segment of a path-like string (drops trailing slashes). */
export function baseName(path: string): string {
  const clean = path.replace(/\/+$/, "");
  const parts = clean.split("/");
  return parts[parts.length - 1] || path;
}

/** Lower-cased extension (no leading dot) extracted from a path or name. */
export function getFileExtension(path: string): string {
  const name = baseName(path);
  const idx = name.lastIndexOf(".");
  if (idx < 0 || idx === name.length - 1) return "";
  return name.slice(idx + 1).toLowerCase();
}

/** Coarse category for a path/name, falling back to "other". */
export function getFileCategory(path: string): string {
  const ext = getFileExtension(path);
  return EXT_MAP[ext] || "other";
}

/**
 * Category-based file badge (label + color), as used by the shares table.
 * e.g. "PDF" / "IMG" / "DOC" / "XLS" / "PPT" / "ZIP" / "TXT" / raw ext.
 */
export function getFileBadge(nameOrPath: string): {
  label: string;
  color: string;
} {
  const ext = getFileExtension(nameOrPath);
  const cat = getFileCategory(nameOrPath);
  if (cat === "image") return { label: "IMG", color: "#5f5f63" };
  if (cat === "pdf") return { label: "PDF", color: "#e5484d" };
  if (cat === "archive") return { label: "ZIP", color: "#f59e0b" };
  if (cat === "office") {
    if (ext.startsWith("xls")) return { label: "XLS", color: "#1e9e5a" };
    if (ext.startsWith("ppt")) return { label: "PPT", color: "#f59e0b" };
    return { label: "DOC", color: "#2b7cd3" };
  }
  if (cat === "document") return { label: "TXT", color: "#5f5f63" };
  return {
    label: (ext || "?").slice(0, 3).toUpperCase(),
    color: "#5f5f63",
  };
}

/**
 * Short uppercase badge derived from a raw extension string (with or without a
 * leading dot), as used by the single-file preview cards. Returns the first 4
 * chars uppercased, or "FILE" when there is no extension.
 */
export function getExtBadge(extension?: string | null): string {
  const ext = (extension || "").replace(/^\./, "");
  return ext ? ext.slice(0, 4).toUpperCase() : "FILE";
}
