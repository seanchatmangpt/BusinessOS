import { writable } from "svelte/store";
import { getApiBaseUrl, getCSRFToken } from "$lib/api/base";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface OsaMessage {
  id: string;
  role: "user" | "osa";
  content: string;
  timestamp: Date;
}

export interface OsaState {
  conversation: OsaMessage[];
  isStreaming: boolean;
  streamingContent: string;
  isExpanded: boolean;
  error: string | null;
}

// ─── Store ───────────────────────────────────────────────────────────────────

function createOsaStore() {
  const initialState: OsaState = {
    conversation: [],
    isStreaming: false,
    streamingContent: "",
    isExpanded: false,
    error: null,
  };

  const { subscribe, update } = writable<OsaState>(initialState);

  // Internal helper to get current state synchronously
  function getState(): OsaState {
    let current: OsaState = initialState;
    const unsub = subscribe((s) => (current = s));
    unsub();
    return current;
  }

  return {
    subscribe,

    setExpanded(expanded: boolean) {
      update((s) => ({ ...s, isExpanded: expanded }));
    },

    clearConversation() {
      update((s) => ({
        ...s,
        conversation: [],
        streamingContent: "",
        isStreaming: false,
        error: null,
      }));
    },

    async sendMessage(content: string) {
      // Add user message immediately
      const userMessage: OsaMessage = {
        id: crypto.randomUUID(),
        role: "user",
        content,
        timestamp: new Date(),
      };

      update((s) => ({
        ...s,
        conversation: [...s.conversation, userMessage],
        isStreaming: true,
        streamingContent: "",
        isExpanded: true,
        error: null,
      }));

      try {
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
        };
        const csrfToken = getCSRFToken();
        if (csrfToken) {
          headers["X-CSRF-Token"] = csrfToken;
        }

        const response = await fetch(`${getApiBaseUrl()}/chat/message`, {
          method: "POST",
          headers,
          credentials: "include",
          body: JSON.stringify({ message: content }),
        });

        if (!response.ok) {
          const errorData = await response
            .json()
            .catch(() => ({ detail: "Chat failed" }));
          throw new Error(
            errorData.detail || `Chat failed (HTTP ${response.status})`,
          );
        }

        if (!response.body) {
          throw new Error("No response stream");
        }

        // Read the streaming response
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let fullContent = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value, { stream: true });
          fullContent += chunk;
          update((s) => ({ ...s, streamingContent: fullContent }));
        }

        // Finalize: add OSA response to conversation
        const osaMessage: OsaMessage = {
          id: crypto.randomUUID(),
          role: "osa",
          content: fullContent,
          timestamp: new Date(),
        };

        update((s) => ({
          ...s,
          conversation: [...s.conversation, osaMessage],
          isStreaming: false,
          streamingContent: "",
        }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to send message";
        console.error("[OSA Store] Send failed:", err);
        update((s) => ({
          ...s,
          isStreaming: false,
          streamingContent: "",
          error: message,
        }));
      }
    },
  };
}

export const osaStore = createOsaStore();
