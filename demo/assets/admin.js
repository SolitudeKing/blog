(function initializeAdminWorkbench() {
  "use strict";

  const $ = (selector, scope = document) => scope.querySelector(selector);
  const $$ = (selector, scope = document) => [...scope.querySelectorAll(selector)];
  const sidebar = $("[data-admin-sidebar]");
  if (!sidebar) return;

  const page = document.body;
  const navToggle = $("[data-admin-nav-toggle]");
  const navClose = $("[data-admin-nav-close]");
  const scrim = $("[data-admin-scrim]");
  const mobileNav = window.matchMedia("(max-width: 860px)");
  const reduceMotion = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
  let navOpen = false;
  let lastNavTrigger = null;

  function focusableElements() {
    return $$(
      'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
      sidebar,
    ).filter((element) => !element.hidden && element.getAttribute("aria-hidden") !== "true");
  }

  function applyNavigationState(open, restoreFocus = false) {
    navOpen = mobileNav.matches && open;
    sidebar.classList.toggle("is-open", navOpen);
    navToggle?.setAttribute("aria-expanded", String(navOpen));
    page.classList.toggle("is-admin-nav-open", navOpen);

    if (mobileNav.matches) {
      sidebar.setAttribute("aria-hidden", String(!navOpen));
      sidebar.inert = !navOpen;
      if (scrim) scrim.hidden = !navOpen;
    } else {
      sidebar.removeAttribute("aria-hidden");
      sidebar.inert = false;
      if (scrim) scrim.hidden = true;
    }

    if (navOpen) {
      lastNavTrigger = navToggle;
      window.requestAnimationFrame(() => {
        $("[aria-current='page']", sidebar)?.focus();
      });
    } else if (restoreFocus) {
      lastNavTrigger?.focus();
    }
  }

  navToggle?.addEventListener("click", () => applyNavigationState(!navOpen));
  navClose?.addEventListener("click", () => applyNavigationState(false, true));
  scrim?.addEventListener("click", () => applyNavigationState(false, true));

  sidebar.addEventListener("click", (event) => {
    const link = event.target.closest("a");
    if (!link || !mobileNav.matches) return;
    const target = link.hash ? document.getElementById(link.hash.slice(1)) : null;
    applyNavigationState(false);
    if (target && link.pathname === window.location.pathname) {
      window.setTimeout(() => {
        target.tabIndex = -1;
        target.focus({ preventScroll: true });
      }, reduceMotion ? 0 : 360);
    }
  });

  document.addEventListener("keydown", (event) => {
    if (!navOpen) return;
    if (event.key === "Escape") {
      event.preventDefault();
      applyNavigationState(false, true);
      return;
    }
    if (event.key !== "Tab") return;

    const focusable = focusableElements();
    const first = focusable[0];
    const last = focusable.at(-1);
    if (!first || !last) return;

    if (!sidebar.contains(document.activeElement)) {
      event.preventDefault();
      (event.shiftKey ? last : first).focus();
      return;
    }

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  });

  const handleViewportChange = () => applyNavigationState(false);
  if (typeof mobileNav.addEventListener === "function") {
    mobileNav.addEventListener("change", handleViewportChange);
  } else {
    mobileNav.addListener(handleViewportChange);
  }
  applyNavigationState(false);

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
    const pause = () => window.clearTimeout(timer);
    toast.addEventListener("mouseenter", pause, { once: true });
    toast.addEventListener("focusin", pause, { once: true });
  }

  $$('[data-admin-action]').forEach((control) => {
    control.addEventListener("click", () => {
      showToast(
        control.dataset.toastTitle || "操作已记录",
        control.dataset.toastMessage || "静态工作台已响应当前操作。",
        control.dataset.toastTone || "info",
      );
    });
  });

  const searchForm = $("[data-admin-search-form]");
  const searchInput = $("[data-admin-search]");
  const filterButtons = $$('[data-admin-filter]');
  const articleRows = $$('[data-admin-article]');
  const resultsStatus = $("[data-admin-results]");
  const emptyState = $("[data-admin-empty]");
  let activeFilter = "all";

  function normalize(value) {
    return value.trim().toLocaleLowerCase("zh-CN");
  }

  function updateArticles() {
    const query = normalize(searchInput?.value ?? "");
    let visibleCount = 0;

    articleRows.forEach((row) => {
      const matchesFilter = activeFilter === "all" || row.dataset.status === activeFilter;
      const matchesQuery = !query || normalize(row.dataset.search ?? row.textContent).includes(query);
      row.hidden = !(matchesFilter && matchesQuery);
      if (!row.hidden) visibleCount += 1;
    });

    if (resultsStatus) {
      const suffix = query ? `，关键词“${searchInput.value.trim()}”` : "";
      resultsStatus.textContent = `显示 ${visibleCount} 篇文章${suffix}`;
    }
    if (emptyState) emptyState.hidden = visibleCount !== 0;
  }

  filterButtons.forEach((button) => {
    button.addEventListener("click", () => {
      activeFilter = button.dataset.adminFilter || "all";
      filterButtons.forEach((item) => {
        item.setAttribute("aria-pressed", String(item === button));
      });
      updateArticles();
    });
  });

  searchForm?.addEventListener("submit", (event) => {
    event.preventDefault();
    updateArticles();
  });
  searchInput?.addEventListener("input", updateArticles);
  document.addEventListener("keydown", (event) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase() === "k") {
      event.preventDefault();
      searchInput?.focus();
    }
  });
  updateArticles();

  const taskInputs = $$('[data-admin-task]');
  const taskCount = $("[data-admin-task-count]");

  function updateTasks() {
    let remaining = 0;
    taskInputs.forEach((input) => {
      input.closest("li")?.classList.toggle("is-complete", input.checked);
      if (!input.checked) remaining += 1;
    });
    if (taskCount) taskCount.textContent = String(remaining);
  }

  taskInputs.forEach((input) => input.addEventListener("change", updateTasks));
  updateTasks();
})();
