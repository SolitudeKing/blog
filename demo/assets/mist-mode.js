(function initializeMistMode(globalScope) {
  "use strict";

  const STORAGE_KEY = "between-tides:mode";
  const MODES = new Set(["light", "dark"]);
  const root = document.documentElement;

  function isMode(value) {
    return MODES.has(value);
  }

  function storedMode() {
    try {
      const value = globalScope.localStorage.getItem(STORAGE_KEY);
      if (isMode(value)) return value;
      if (value !== null) globalScope.localStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      return null;
    }
    return null;
  }

  function preferredMode() {
    if (isMode(root.dataset.mode)) return root.dataset.mode;
    return globalScope.matchMedia?.("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }

  function applyMode(mode, persist = false) {
    if (!isMode(mode)) throw new TypeError(`Unsupported mode: ${String(mode)}`);
    root.dataset.mode = mode;
    if (persist) {
      try {
        globalScope.localStorage.setItem(STORAGE_KEY, mode);
      } catch (error) {
        // Mode still works when storage is unavailable.
      }
    }
    document.dispatchEvent(new CustomEvent("mistmodechange", { detail: { mode } }));
    return mode;
  }

  const api = Object.freeze({
    getMode() {
      return isMode(root.dataset.mode) ? root.dataset.mode : applyMode(preferredMode());
    },
    setMode(mode) {
      return applyMode(mode, true);
    },
    toggleMode() {
      return applyMode(api.getMode() === "dark" ? "light" : "dark", true);
    },
    clearPreference() {
      try {
        globalScope.localStorage.removeItem(STORAGE_KEY);
      } catch (error) {
        // Ignore unavailable storage.
      }
      return applyMode(preferredMode());
    },
  });

  globalScope.addEventListener("storage", (event) => {
    if (event.key !== STORAGE_KEY) return;
    applyMode(isMode(event.newValue) ? event.newValue : preferredMode());
  });

  globalScope.MistMode = api;
  applyMode(storedMode() ?? preferredMode());
})(window);
