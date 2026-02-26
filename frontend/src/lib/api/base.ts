// Backend URLs
const LOCAL_BACKEND_URL = "http://localhost:8001";
const CLOUD_RUN_URL = "https://businessos-api-460433387676.us-central1.run.app";

// API Version (centralized configuration)
const API_VERSION = "v1";

// Shared fetch logic copied from the original ApiClient.request implementation
function getApiBase(): string {
  if (typeof window === "undefined") {
    const result = import.meta.env.VITE_API_URL || `/api/${API_VERSION}`;
    return result;
  }

  const isElectron = "electron" in window;
  const isDev =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";

  if (isElectron) {
    const mode = localStorage.getItem("businessos_mode");
    let cloudUrl = localStorage.getItem("businessos_cloud_url");

    // Auto-configure URL if not set
    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
      localStorage.setItem("businessos_cloud_url", cloudUrl);
    }

    if (mode === "cloud" && cloudUrl) {
      const result = `${cloudUrl}/api/${API_VERSION}`;
      return result;
    } else if (mode === "local") {
      return `http://localhost:18080/api/${API_VERSION}`;
    }
    const result = `${cloudUrl}/api/${API_VERSION}`;
    return result;
  }

  // Web app: use env var, or auto-detect based on environment
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  // In development, use relative URLs through Vite proxy to ensure CSRF cookies work
  // (same origin for cookie set and API calls)
  const result = isDev
    ? `/api/${API_VERSION}`
    : `${CLOUD_RUN_URL}/api/${API_VERSION}`;
  return result;
}

export const getApiBaseUrl = () => getApiBase();
export const API_BASE = getApiBase();

// Get CSRF token from cookie
export function getCSRFToken(): string | null {
  if (typeof document === "undefined") return null;

  const cookies = document.cookie.split(";");
  for (const cookie of cookies) {
    const trimmed = cookie.trim();
    const eqIndex = trimmed.indexOf("=");
    if (eqIndex === -1) continue;

    const name = trimmed.substring(0, eqIndex);
    const value = trimmed.substring(eqIndex + 1); // Get everything after first =

    if (name === "csrf_token") {
      return value;
    }
  }
  return null;
}

// Initialize CSRF token by calling the backend endpoint
// This should be called before any state-changing requests (POST, PUT, DELETE)
export async function initCSRF(): Promise<void> {
  if (typeof window === "undefined") {
    return;
  }

  try {
    // Use relative URL to go through Vite proxy in development
    // This ensures the CSRF cookie is set in the correct domain context (localhost:5173)
    // and can be read by subsequent fetch calls
    const isDev =
      window.location.hostname === "localhost" ||
      window.location.hostname === "127.0.0.1";
    const csrfUrl = isDev ? "/api/auth/csrf" : `${getApiBaseUrl()}/auth/csrf`;

    console.log("[CSRF] Initializing CSRF token from:", csrfUrl);
    const response = await fetch(csrfUrl, {
      method: "GET",
      credentials: "include",
    });

    if (response.ok) {
      const data = await response.json();
      console.log(
        "[CSRF] CSRF token initialized successfully:",
        data.csrf_token?.substring(0, 20) + "...",
      );
      console.log(
        "[CSRF] Cookie should be set now. Document.cookie:",
        document.cookie.includes("csrf_token")
          ? "CSRF cookie found"
          : "CSRF cookie NOT found",
      );
    } else {
      console.error(
        "[CSRF] Failed to initialize CSRF token. Status:",
        response.status,
      );
    }
  } catch (error) {
    console.warn("[CSRF] Failed to initialize CSRF token:", error);
  }
}

// Add CSRF token to headers for state-changing requests
function addCSRFToken(
  method: string,
  headers: Record<string, string>,
): Record<string, string> {
  const stateChangingMethods = ["POST", "PUT", "PATCH", "DELETE"];
  if (stateChangingMethods.includes(method.toUpperCase())) {
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }
  }
  return headers;
}

/**
 * Get the backend server base URL (without /api suffix)
 * Use this for image URLs and other non-API resources
 */
export function getBackendUrl(): string {
  if (typeof window === "undefined") {
    return "";
  }

  const isElectron = "electron" in window;
  const isDev =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";

  if (isElectron) {
    const mode = localStorage.getItem("businessos_mode");
    let cloudUrl = localStorage.getItem("businessos_cloud_url");

    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
    }

    if (mode === "cloud" && cloudUrl) {
      return cloudUrl;
    } else if (mode === "local") {
      return "http://localhost:18080";
    }
    return cloudUrl;
  }

  // Web app: use env var base, or auto-detect based on environment
  if (import.meta.env.VITE_BACKEND_URL) {
    return import.meta.env.VITE_BACKEND_URL;
  }
  return isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
}

export interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
  timeout?: number; // Timeout in milliseconds
}

export async function request<T>(
  endpoint: string,
  options: RequestOptions = {},
): Promise<T> {
  const { method = "GET", body, headers = {}, timeout } = options;

  if (body && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json";
  }

  // Add CSRF token for state-changing requests
  const finalHeaders = addCSRFToken(method, headers);

  const baseUrl = getApiBaseUrl();

  // Create abort controller for timeout
  const controller = new AbortController();
  let timeoutId: NodeJS.Timeout | number | undefined;

  if (timeout && timeout > 0) {
    timeoutId = setTimeout(() => controller.abort(), timeout);
  }

  try {
    const response = await fetch(`${baseUrl}${endpoint}`, {
      method,
      headers: finalHeaders,
      credentials: "include",
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal,
    });

    if (!response.ok) {
      // Handle 401 for non-auth endpoints: session expired, redirect to login
      if (response.status === 401 && !endpoint.includes("/auth/")) {
        const { clearSession } = await import("$lib/auth-client");
        clearSession();
        if (typeof window !== "undefined") {
          window.location.href = "/login";
        }
        throw new Error("Session expired");
      }

      const error = await response
        .json()
        .catch(() => ({ detail: "Request failed" }));
      const errorMessage = error.detail || error.message || "Request failed";
      console.error(
        `[API] ${method} ${endpoint} failed with status ${response.status}: ${errorMessage}`,
      );
      throw new Error(`${errorMessage} (HTTP ${response.status})`);
    }

    return response.json();
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw new Error(`Request timeout after ${timeout}ms`);
    }
    throw error;
  } finally {
    if (timeoutId !== undefined) {
      clearTimeout(timeoutId as number);
    }
  }
}

// For raw response access (like original apiClient)
export const raw = {
  async get(endpoint: string): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "GET",
      credentials: "include",
    });
  },
  async post(endpoint: string, body?: unknown): Promise<Response> {
    const headers = addCSRFToken("POST", {
      "Content-Type": "application/json",
    });
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "POST",
      headers,
      credentials: "include",
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async postFormData(endpoint: string, formData: FormData): Promise<Response> {
    const headers = addCSRFToken("POST", {});
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "POST",
      headers,
      credentials: "include",
      body: formData,
    });
  },
  async put(endpoint: string, body?: unknown): Promise<Response> {
    const headers = addCSRFToken("PUT", { "Content-Type": "application/json" });
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "PUT",
      headers,
      credentials: "include",
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async delete(endpoint: string): Promise<Response> {
    const headers = addCSRFToken("DELETE", {});
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "DELETE",
      headers,
      credentials: "include",
    });
  },
};
