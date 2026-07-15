(function initializeBlog() {
  "use strict";

  const $ = (selector, scope = document) => scope.querySelector(selector);
  const $$ = (selector, scope = document) => [...scope.querySelectorAll(selector)];
  const reduceMotion = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;

  function syncModeControls() {
    const currentMode = window.MistMode?.getMode() ?? "light";
    const nextLabel = currentMode === "dark" ? "切换至亮色模式" : "切换至暗色模式";
    const themeColor = $("meta[name='theme-color']");
    const pageColor = getComputedStyle(document.documentElement)
      .getPropertyValue("--bg-primary")
      .trim();
    if (themeColor && pageColor) themeColor.content = pageColor;
    $$('[data-mode-toggle]').forEach((button) => {
      button.setAttribute("aria-label", nextLabel);
      button.setAttribute("title", nextLabel);
      const label = $(".mist-mode-label", button);
      if (label) label.textContent = currentMode === "dark" ? "亮色" : "暗色";
    });
  }

  $$('[data-mode-toggle]').forEach((button) => {
    button.addEventListener("click", () => window.MistMode?.toggleMode());
  });
  document.addEventListener("mistmodechange", syncModeControls);
  syncModeControls();

  const menuButton = $("[data-menu-toggle]");
  const mobileMenu = $("[data-mobile-menu]");
  let lastMenuTrigger = null;

  function setMenu(open, restoreFocus = false) {
    if (!menuButton || !mobileMenu) return;
    menuButton.setAttribute("aria-expanded", String(open));
    mobileMenu.hidden = !open;
    document.body.classList.toggle("is-menu-open", open);
    if (open) {
      lastMenuTrigger = menuButton;
      $("a", mobileMenu)?.focus();
    } else if (restoreFocus) {
      lastMenuTrigger?.focus();
    }
  }

  menuButton?.addEventListener("click", () => {
    setMenu(menuButton.getAttribute("aria-expanded") !== "true");
  });

  mobileMenu?.addEventListener("click", (event) => {
    if (event.target.closest("a")) setMenu(false);
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && menuButton?.getAttribute("aria-expanded") === "true") {
      setMenu(false, true);
    }
  });

  document.addEventListener("click", (event) => {
    if (
      menuButton?.getAttribute("aria-expanded") === "true" &&
      !event.target.closest("[data-menu-toggle]") &&
      !event.target.closest("[data-mobile-menu]")
    ) {
      setMenu(false);
    }
  });

  const toastRegion = $("[data-toast-region]");

  function showToast(title, message = "", tone = "info") {
    if (!toastRegion) return;
    const toast = document.createElement("div");
    toast.className = `mist-toast mist-toast--${tone}`;
    toast.setAttribute("role", tone === "error" ? "alert" : "status");
    toast.tabIndex = -1;

    const indicator = document.createElement("span");
    indicator.className = "mist-toast__indicator";
    indicator.setAttribute("aria-hidden", "true");

    const copy = document.createElement("div");
    const heading = document.createElement("p");
    heading.className = "mist-toast__title";
    heading.textContent = title;
    copy.append(heading);
    if (message) {
      const detail = document.createElement("p");
      detail.className = "mist-toast__message";
      detail.textContent = message;
      copy.append(detail);
    }

    const close = document.createElement("button");
    close.className = "mist-toast__close";
    close.type = "button";
    close.setAttribute("aria-label", "关闭提示");
    close.innerHTML =
      '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18"/></svg>';

    const remove = () => toast.remove();
    close.addEventListener("click", remove);
    toast.addEventListener("keydown", (event) => {
      if (event.key === "Escape") remove();
    });
    toast.append(indicator, copy, close);
    toastRegion.prepend(toast);
    while (toastRegion.children.length > 3) toastRegion.lastElementChild?.remove();

    let timer = window.setTimeout(remove, 4800);
    toast.addEventListener("mouseenter", () => window.clearTimeout(timer), { once: true });
    toast.addEventListener("focusin", () => window.clearTimeout(timer), { once: true });
  }

  $$('[data-newsletter]').forEach((form) => {
    const input = $('input[type="email"]', form);
    const submit = $('button[type="submit"]', form);
    const error = $("[data-field-error]", form);

    form.addEventListener("submit", (event) => {
      event.preventDefault();
      if (!input || !submit) return;

      const valid = input.validity.valid && input.value.trim().length > 0;
      input.setAttribute("aria-invalid", String(!valid));
      if (error) error.textContent = valid ? "" : "请填写一个有效的邮箱地址。";
      if (!valid) {
        input.focus();
        return;
      }

      submit.disabled = true;
      submit.setAttribute("aria-busy", "true");
      const originalLabel = submit.textContent;
      submit.textContent = "正在靠岸…";

      window.setTimeout(() => {
        submit.disabled = false;
        submit.removeAttribute("aria-busy");
        submit.textContent = originalLabel;
        form.reset();
        input.setAttribute("aria-invalid", "false");
        showToast("订阅成功", "下一封漂流信会在两周内抵达。", "success");
      }, reduceMotion ? 0 : 650);
    });
  });

  const backtop = $("[data-backtop]");
  function updateBacktop() {
    backtop?.classList.toggle("is-visible", window.scrollY > 720);
  }
  window.addEventListener("scroll", updateBacktop, { passive: true });
  updateBacktop();
  backtop?.addEventListener("click", () => {
    window.scrollTo({ top: 0, behavior: reduceMotion ? "auto" : "smooth" });
    window.setTimeout(() => $("main")?.focus({ preventScroll: true }), reduceMotion ? 0 : 420);
  });

  const progressBar = $("[data-reading-progress]");
  const articleBody = $("[data-article-body]");
  function updateReadingProgress() {
    if (!progressBar || !articleBody) return;
    const articleTop = articleBody.getBoundingClientRect().top + window.scrollY;
    const articleHeight = articleBody.offsetHeight;
    const viewport = window.innerHeight;
    const traveled = window.scrollY - articleTop + viewport * 0.35;
    const progress = Math.max(0, Math.min(1, traveled / Math.max(1, articleHeight - viewport * 0.45)));
    progressBar.style.setProperty("--reading-progress", `${(progress * 100).toFixed(2)}%`);
  }
  if (progressBar && articleBody) {
    window.addEventListener("scroll", updateReadingProgress, { passive: true });
    window.addEventListener("resize", updateReadingProgress);
    updateReadingProgress();
  }

  const tocLinks = $$('[data-toc] a[href^="#"]');
  const tocSections = tocLinks
    .map((link) => document.getElementById(link.hash.slice(1)))
    .filter(Boolean);
  if (tocSections.length && "IntersectionObserver" in window) {
    const tocObserver = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)[0];
        if (!visible) return;
        tocLinks.forEach((link) => {
          if (link.hash === `#${visible.target.id}`) link.setAttribute("aria-current", "location");
          else link.removeAttribute("aria-current");
        });
      },
      { rootMargin: "-22% 0px -66% 0px", threshold: 0 },
    );
    tocSections.forEach((section) => tocObserver.observe(section));
  }

  [...tocLinks, ...$$('[data-year-nav] a[href^="#"]')].forEach((link) => {
    link.addEventListener("click", (event) => {
      const target = document.getElementById(link.hash.slice(1));
      if (!target) return;
      event.preventDefault();
      target.scrollIntoView({ behavior: reduceMotion ? "auto" : "smooth", block: "start" });
      target.tabIndex = -1;
      window.setTimeout(() => target.focus({ preventScroll: true }), reduceMotion ? 0 : 360);
    });
  });

  const yearLinks = $$('[data-year-nav] a[href^="#"]');
  const yearSections = yearLinks
    .map((link) => document.getElementById(link.hash.slice(1)))
    .filter(Boolean);
  if (yearSections.length && "IntersectionObserver" in window) {
    const yearObserver = new IntersectionObserver(
      (entries) => {
        const visible = entries.find((entry) => entry.isIntersecting);
        if (!visible) return;
        yearLinks.forEach((link) => {
          if (link.hash === `#${visible.target.id}`) link.setAttribute("aria-current", "location");
          else link.removeAttribute("aria-current");
        });
      },
      { rootMargin: "-25% 0px -65% 0px", threshold: 0 },
    );
    yearSections.forEach((section) => yearObserver.observe(section));
  }

  const searchInput = $("[data-search-input]");
  const searchItems = $$('[data-search-item]');
  const resultCount = $("[data-search-count]");
  const resultStatus = $("[data-search-status]");
  const emptyState = $("[data-search-empty]");
  const clearSearch = $("[data-search-clear]");
  const queryButtons = $$('[data-search-query]');

  function normalize(value) {
    return value.trim().toLocaleLowerCase("zh-CN");
  }

  function updateSearch() {
    if (!searchInput || !searchItems.length) return;
    const query = normalize(searchInput.value);
    let visibleCount = 0;

    searchItems.forEach((item) => {
      const matches = !query || normalize(item.dataset.search ?? item.textContent).includes(query);
      item.hidden = !matches;
      if (matches) visibleCount += 1;
    });

    if (resultCount) resultCount.textContent = String(visibleCount);
    if (resultStatus) {
      resultStatus.textContent = query
        ? `关键词“${searchInput.value.trim()}”找到 ${visibleCount} 篇文章。`
        : `当前显示全部 ${visibleCount} 篇文章。`;
    }
    if (emptyState) emptyState.hidden = visibleCount !== 0;
    if (clearSearch) clearSearch.hidden = query.length === 0;
    queryButtons.forEach((button) => {
      button.setAttribute("aria-pressed", String(normalize(button.dataset.searchQuery ?? "") === query));
    });
  }

  if (searchInput) {
    const initialQuery = new URLSearchParams(window.location.search).get("q");
    if (initialQuery) searchInput.value = initialQuery;
    searchInput.addEventListener("input", updateSearch);
    updateSearch();
  }

  queryButtons.forEach((button) => {
    button.addEventListener("click", () => {
      if (!searchInput) return;
      searchInput.value = button.dataset.searchQuery ?? "";
      updateSearch();
      searchInput.focus();
    });
  });

  clearSearch?.addEventListener("click", () => {
    if (!searchInput) return;
    searchInput.value = "";
    updateSearch();
    searchInput.focus();
  });

  $$('[data-copy-link]').forEach((button) => {
    button.addEventListener("click", async () => {
      const value = button.dataset.copyValue || window.location.href;
      try {
        await navigator.clipboard.writeText(value);
        showToast("已复制", button.dataset.copyMessage || "链接已复制到剪贴板。", "success");
      } catch (error) {
        showToast("暂时无法复制", "请从地址栏手动复制。", "error");
      }
    });
  });

  const readingPool = [
    "article.html",
    "article.html#slow-systems",
    "article.html#gentle-interface",
    "archives.html#year-2024",
  ];
  $$('[data-random-reading]').forEach((link) => {
    link.addEventListener("click", (event) => {
      event.preventDefault();
      const current = Number(sessionStorage.getItem("between-tides:random") ?? "-1");
      const next = (current + 1) % readingPool.length;
      sessionStorage.setItem("between-tides:random", String(next));
      window.location.href = readingPool[next];
    });
  });

  $$('[data-current-year]').forEach((node) => {
    node.textContent = String(new Date().getFullYear());
  });
})();
