const API_BASE = import.meta.env.VITE_API_BASE_URL || "";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (response.status === 204) {
    return null;
  }

  const contentType = response.headers.get("Content-Type") || "";
  let data = null;
  let text = "";

  if (contentType.includes("application/json")) {
    data = await response.json();
  } else {
    text = await response.text();
  }

  if (!response.ok) {
    const message = data?.error || text.trim() || `Request failed with ${response.status}`;
    const error = new Error(message);
    error.status = response.status;
    throw error;
  }
  return data;
}

export const api = {
  get: (path) => request(path),
  post: (path, body) =>
    request(path, {
      method: "POST",
      body: JSON.stringify(body),
    }),
  put: (path, body) =>
    request(path, {
      method: "PUT",
      body: JSON.stringify(body),
    }),
  delete: (path) =>
    request(path, {
      method: "DELETE",
    }),
};
