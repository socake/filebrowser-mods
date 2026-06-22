import { fetchURL, fetchJSON, StatusError } from "./utils";

export async function list() {
  return fetchJSON<Role[]>(`/api/roles`, {});
}

export async function get(id: number) {
  return fetchJSON<Role>(`/api/roles/${id}`, {});
}

export async function create(role: Partial<Role>) {
  const res = await fetchURL(`/api/roles`, {
    method: "POST",
    body: JSON.stringify(role),
  });

  if (res.status === 201) {
    return res.headers.get("Location");
  }

  throw new StatusError(await res.text(), res.status);
}

export async function update(role: Role) {
  await fetchURL(`/api/roles/${role.id}`, {
    method: "PUT",
    body: JSON.stringify(role),
  });
}

export async function remove(id: number) {
  await fetchURL(`/api/roles/${id}`, {
    method: "DELETE",
  });
}
