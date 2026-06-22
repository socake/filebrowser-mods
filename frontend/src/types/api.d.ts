type ApiMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

type ApiContent =
  | Blob
  | File
  | Pick<ReadableStreamDefaultReader<any>, "read">
  | "";

interface ApiOpts {
  method?: ApiMethod;
  headers?: object;
  body?: any;
  signal?: AbortSignal;
}

interface TusSettings {
  retryCount: number;
  chunkSize: number;
}

type ChecksumAlg = "md5" | "sha1" | "sha256" | "sha512";

interface Share {
  hash: string;
  path: string;
  expire?: any;
  userID?: number;
  token?: string;
  username?: string;
  type?: "preview" | "download";
  password_hash?: string;
  createdAt?: number;
}

interface SearchParams {
  [key: string]: string;
}

interface RolePermissions {
  admin: boolean;
  create: boolean;
  rename: boolean;
  modify: boolean;
  delete: boolean;
  share: boolean;
  download: boolean;
  execute: boolean;
}

interface Role {
  id: number;
  name: string;
  description: string;
  permissions: RolePermissions;
  scope: string;
  isPreset: boolean;
  sort: number;
}
