(() => {
  const state = {
    auth: null,
    events: [],
    selectedEventId: "",
    snippets: [],
  };

  const ui = {
    loginPanel: document.querySelector("#loginPanel"),
    workspace: document.querySelector("#workspace"),
    loginForm: document.querySelector("#loginForm"),
    tenantSlug: document.querySelector("#tenantSlug"),
    email: document.querySelector("#email"),
    loginHint: document.querySelector("#loginHint"),
    loginSubmitBtn: document.querySelector("#loginSubmitBtn"),
    currentUser: document.querySelector("#currentUser"),
    logoutBtn: document.querySelector("#logoutBtn"),
    flash: document.querySelector("#flash"),
    tabs: Array.from(document.querySelectorAll(".tab")),
    panes: Array.from(document.querySelectorAll(".tabpane")),
    statsGrid: document.querySelector("#statsGrid"),
    refreshDashboardBtn: document.querySelector("#refreshDashboardBtn"),
    nextEventsTableBody: document.querySelector("#nextEventsTableBody"),
    eventForm: document.querySelector("#eventForm"),
    eventSubmitBtn: document.querySelector("#eventSubmitBtn"),
    refreshEventsBtn: document.querySelector("#refreshEventsBtn"),
    eventsTableBody: document.querySelector("#eventsTableBody"),
    registrationEventSelect: document.querySelector("#registrationEventSelect"),
    refreshRegistrationsBtn: document.querySelector("#refreshRegistrationsBtn"),
    manualRegistrationForm: document.querySelector("#manualRegistrationForm"),
    manualRegistrationSubmitBtn: document.querySelector("#manualRegistrationSubmitBtn"),
    registrationsTableBody: document.querySelector("#registrationsTableBody"),
    snippetForm: document.querySelector("#snippetForm"),
    snippetSubmitBtn: document.querySelector("#snippetSubmitBtn"),
    refreshSnippetsBtn: document.querySelector("#refreshSnippetsBtn"),
    snippetsTableBody: document.querySelector("#snippetsTableBody"),
    snippetEmbedOutput: document.querySelector("#snippetEmbedOutput"),
  };

  const STORAGE_TENANT_KEY = "eep_admin_tenant_slug";

  bindUI();
  restoreTenantSlug();
  refreshSession();

  function bindUI() {
    ui.loginForm?.addEventListener("submit", onLoginSubmit);
    ui.logoutBtn?.addEventListener("click", onLogoutClick);
    ui.refreshDashboardBtn?.addEventListener("click", () => loadDashboard(true));
    ui.refreshEventsBtn?.addEventListener("click", () => loadEvents(true));
    ui.eventForm?.addEventListener("submit", onEventCreateSubmit);
    ui.refreshRegistrationsBtn?.addEventListener("click", () => loadRegistrations(state.selectedEventId, true));
    ui.manualRegistrationForm?.addEventListener("submit", onManualRegistrationSubmit);
    ui.snippetForm?.addEventListener("submit", onSnippetCreateSubmit);
    ui.refreshSnippetsBtn?.addEventListener("click", () => loadSnippets(true));
    ui.registrationEventSelect?.addEventListener("change", (ev) => {
      const id = String(ev.target.value || "").trim();
      state.selectedEventId = id;
      loadRegistrations(id, false);
    });

    ui.tabs.forEach((tab) => {
      tab.addEventListener("click", () => activateTab(String(tab.dataset.tab || "dashboard")));
    });
  }

  function restoreTenantSlug() {
    const remembered = localStorage.getItem(STORAGE_TENANT_KEY);
    if (remembered && ui.tenantSlug && !ui.tenantSlug.value) {
      ui.tenantSlug.value = remembered;
    }
  }

  function activateTab(tabName) {
    ui.tabs.forEach((tab) => {
      tab.classList.toggle("is-active", tab.dataset.tab === tabName);
    });
    ui.panes.forEach((pane) => {
      pane.classList.toggle("is-active", pane.id === `tab-${tabName}`);
    });
  }

  async function refreshSession() {
    try {
      const payload = await apiRequest("/api/v1/auth/me");
      state.auth = payload;
      showWorkspace(payload);
      await Promise.all([loadDashboard(false), loadEvents(false), loadSnippets(false)]);
    } catch (err) {
      if (err.status === 401) {
        showLogin();
        return;
      }
      showLogin();
      setFlash(`Session konnte nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function showLogin() {
    state.auth = null;
    ui.loginPanel.hidden = false;
    ui.workspace.hidden = true;
    ui.logoutBtn.hidden = true;
    ui.currentUser.textContent = "Nicht angemeldet";
    ui.loginHint.textContent = "";
  }

  function showWorkspace(me) {
    const user = me?.user || {};
    const tenant = me?.tenant || {};

    ui.loginPanel.hidden = true;
    ui.workspace.hidden = false;
    ui.logoutBtn.hidden = false;
    ui.currentUser.textContent = `${user.name || user.email || "User"} @ ${tenant.slug || "tenant"}`;
    activateTab("dashboard");
  }

  async function onLoginSubmit(event) {
    event.preventDefault();
    clearFlash();

    const tenantSlug = String(ui.tenantSlug?.value || "").trim();
    const email = String(ui.email?.value || "").trim().toLowerCase();

    if (!tenantSlug || !email) {
      ui.loginHint.textContent = "Bitte Tenant-Slug und E-Mail eintragen.";
      return;
    }

    localStorage.setItem(STORAGE_TENANT_KEY, tenantSlug);
    setButtonBusy(ui.loginSubmitBtn, true, "Sende...");

    try {
      await apiRequest("/api/v1/auth/magic-link/request", {
        method: "POST",
        body: JSON.stringify({
          tenant_slug: tenantSlug,
          email,
          purpose: "organizer_login",
          redirect_path: "/admin",
        }),
      });
      ui.loginHint.textContent = "Magic Link wurde versendet. Bitte E-Mail oeffnen und den Link klicken.";
    } catch (err) {
      ui.loginHint.textContent = `Login-Link konnte nicht gesendet werden: ${errorMessage(err)}`;
    } finally {
      setButtonBusy(ui.loginSubmitBtn, false, "Magic Link senden");
    }
  }

  async function onLogoutClick() {
    clearFlash();
    try {
      await apiRequest("/api/v1/auth/logout", { method: "POST" });
    } catch (err) {
      setFlash(`Logout-Fehler: ${errorMessage(err)}`, "error");
    }
    showLogin();
  }

  async function loadDashboard(notify) {
    try {
      const payload = await apiRequest("/api/v1/admin/dashboard");
      renderDashboard(payload);
      if (notify) {
        setFlash("Dashboard aktualisiert.");
      }
    } catch (err) {
      setFlash(`Dashboard konnte nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderDashboard(payload) {
    const stats = payload?.stats || {};
    const cards = [
      { label: "Heute", value: stats.today_events },
      { label: "Upcoming", value: stats.upcoming_events },
      { label: "Bestaetigt", value: stats.confirmed_participants },
      { label: "Warteliste", value: stats.waitlist_entries },
      { label: "Freie Plaetze", value: stats.free_seats },
      { label: "Offene E-Mail-Jobs", value: stats.open_email_jobs },
    ];

    ui.statsGrid.innerHTML = cards
      .map((item) => {
        const value = item.value === null || item.value === undefined ? "-" : String(item.value);
        return `<article class="stat"><div class="label">${escapeHTML(item.label)}</div><div class="value">${escapeHTML(value)}</div></article>`;
      })
      .join("");

    const nextEvents = Array.isArray(payload?.next_events) ? payload.next_events : [];
    if (nextEvents.length === 0) {
      ui.nextEventsTableBody.innerHTML = rowMessage("Noch keine kommenden Events.", 5);
      return;
    }

    ui.nextEventsTableBody.innerHTML = nextEvents
      .map((item) => {
        return `
          <tr>
            <td>${escapeHTML(formatDateTime(item.starts_at))}</td>
            <td>${escapeHTML(item.title || "-")}</td>
            <td>${statusPill(item.status)}</td>
            <td>${escapeHTML(String(item.confirmed_participants ?? "-"))}</td>
            <td>${escapeHTML(String(item.waitlist_entries ?? "-"))}</td>
          </tr>
        `;
      })
      .join("");
  }

  async function loadEvents(notify) {
    try {
      const payload = await apiRequest("/api/v1/admin/events");
      const items = Array.isArray(payload?.items) ? payload.items.slice() : [];
      items.sort((a, b) => String(a.starts_at || "").localeCompare(String(b.starts_at || "")));
      state.events = items;
      renderEvents(items);
      fillRegistrationEventSelect(items);
      if (notify) {
        setFlash("Events aktualisiert.");
      }
    } catch (err) {
      setFlash(`Events konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderEvents(items) {
    if (!items.length) {
      ui.eventsTableBody.innerHTML = rowMessage("Noch keine Events vorhanden.", 5);
      return;
    }

    ui.eventsTableBody.innerHTML = items
      .map((item) => {
        const publishBtn = item.status === "draft"
          ? `<button class="btn tiny ok" type="button" data-action="publish" data-id="${escapeAttr(item.id)}">Publish</button>`
          : "";
        const unpublishBtn = item.status === "scheduled" || item.status === "postponed"
          ? `<button class="btn tiny warn" type="button" data-action="unpublish" data-id="${escapeAttr(item.id)}">Unpublish</button>`
          : "";

        return `
          <tr>
            <td>${escapeHTML(formatDateTime(item.starts_at))}</td>
            <td><strong>${escapeHTML(item.title || "-")}</strong><br><span class="muted">${escapeHTML(item.slug || "")}</span></td>
            <td>${statusPill(item.status)}</td>
            <td>${item.is_public ? "Ja" : "Nein"}</td>
            <td>
              <div class="row-actions">
                ${publishBtn}
                ${unpublishBtn}
                <button class="btn tiny light" type="button" data-action="focus-registrations" data-id="${escapeAttr(item.id)}">Teilnehmer</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    ui.eventsTableBody.querySelectorAll("button[data-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.action || "");
        const id = String(btn.dataset.id || "");

        if (action === "focus-registrations") {
          activateTab("registrations");
          state.selectedEventId = id;
          ui.registrationEventSelect.value = id;
          await loadRegistrations(id, false);
          return;
        }

        if (!id || !action) {
          return;
        }

        setButtonBusy(btn, true, "...");
        try {
          await apiRequest(`/api/v1/admin/events/${encodeURIComponent(id)}/${encodeURIComponent(action)}`, {
            method: "POST",
          });
          await Promise.all([loadEvents(false), loadDashboard(false)]);
          setFlash(`Event-Aktion '${action}' wurde ausgefuehrt.`);
        } catch (err) {
          setFlash(`Event-Aktion fehlgeschlagen: ${errorMessage(err)}`, "error");
        } finally {
          setButtonBusy(btn, false);
        }
      });
    });
  }

  async function onEventCreateSubmit(event) {
    event.preventDefault();
    clearFlash();

    const formData = new FormData(ui.eventForm);
    const title = String(formData.get("title") || "").trim();
    const providedSlug = String(formData.get("slug") || "").trim();
    const startsAtLocal = String(formData.get("starts_at") || "").trim();
    const endsAtLocal = String(formData.get("ends_at") || "").trim();

    if (!title || !startsAtLocal) {
      setFlash("Titel und Startzeit sind Pflichtfelder.", "error");
      return;
    }

    const startsAt = toISO(startsAtLocal);
    const endsAt = endsAtLocal ? toISO(endsAtLocal) : "";
    if (!startsAt) {
      setFlash("Startzeit ist ungueltig.", "error");
      return;
    }
    if (endsAtLocal && !endsAt) {
      setFlash("Endzeit ist ungueltig.", "error");
      return;
    }

    const fallbackSlug = `event-${Math.floor(Date.now() / 1000)}`;
    const slug = providedSlug || slugify(title) || fallbackSlug;
    const maxParticipantsRaw = String(formData.get("max_participants") || "").trim();

    const body = {
      slug,
      title,
      starts_at: startsAt,
      ends_at: endsAt,
      timezone: String(formData.get("timezone") || "Europe/Berlin").trim() || "Europe/Berlin",
      participation_mode: String(formData.get("participation_mode") || "onsite").trim() || "onsite",
      location_name: String(formData.get("location_name") || "").trim(),
      online_url: String(formData.get("online_url") || "").trim(),
      is_public: ui.eventForm.querySelector("input[name='is_public']")?.checked === true,
      registration_enabled: ui.eventForm.querySelector("input[name='registration_enabled']")?.checked === true,
      waitlist_enabled: ui.eventForm.querySelector("input[name='waitlist_enabled']")?.checked === true,
      max_participants: maxParticipantsRaw ? Number(maxParticipantsRaw) : null,
    };

    setButtonBusy(ui.eventSubmitBtn, true, "Speichere...");
    try {
      await apiRequest("/api/v1/admin/events", {
        method: "POST",
        body: JSON.stringify(body),
      });
      ui.eventForm.reset();
      const timezoneInput = ui.eventForm.querySelector("input[name='timezone']");
      if (timezoneInput) {
        timezoneInput.value = "Europe/Berlin";
      }
      const checkboxes = ["is_public", "registration_enabled", "waitlist_enabled"];
      checkboxes.forEach((name) => {
        const cb = ui.eventForm.querySelector(`input[name='${name}']`);
        if (cb) {
          cb.checked = true;
        }
      });

      await Promise.all([loadEvents(false), loadDashboard(false)]);
      setFlash("Event wurde angelegt.");
      activateTab("events");
    } catch (err) {
      setFlash(`Event konnte nicht angelegt werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.eventSubmitBtn, false, "Event speichern");
    }
  }

  function fillRegistrationEventSelect(items) {
    const options = items.map((item) => {
      return `<option value="${escapeAttr(item.id)}">${escapeHTML(formatDateTime(item.starts_at))} - ${escapeHTML(item.title || item.slug || item.id)}</option>`;
    });

    if (!options.length) {
      ui.registrationEventSelect.innerHTML = "<option value=''>Keine Events verfuegbar</option>";
      state.selectedEventId = "";
      ui.registrationsTableBody.innerHTML = rowMessage("Bitte zuerst ein Event anlegen.", 6);
      return;
    }

    ui.registrationEventSelect.innerHTML = options.join("");

    const stillExists = items.some((item) => item.id === state.selectedEventId);
    if (!stillExists) {
      state.selectedEventId = items[0].id;
    }
    ui.registrationEventSelect.value = state.selectedEventId;

    if (state.selectedEventId) {
      loadRegistrations(state.selectedEventId, false);
    }
  }

  async function loadRegistrations(eventID, notify) {
    if (!eventID) {
      ui.registrationsTableBody.innerHTML = rowMessage("Kein Event ausgewaehlt.", 6);
      return;
    }

    try {
      const payload = await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventID)}/registrations`);
      const items = Array.isArray(payload?.items) ? payload.items : [];
      renderRegistrations(items);
      if (notify) {
        setFlash("Teilnehmerliste aktualisiert.");
      }
    } catch (err) {
      setFlash(`Teilnehmer konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  async function onManualRegistrationSubmit(event) {
    event.preventDefault();
    clearFlash();

    const eventID = String(state.selectedEventId || ui.registrationEventSelect?.value || "").trim();
    if (!eventID) {
      setFlash("Bitte zuerst ein Event auswaehlen.", "error");
      return;
    }

    const formData = new FormData(ui.manualRegistrationForm);
    const name = String(formData.get("name") || "").trim();
    const email = String(formData.get("email") || "").trim().toLowerCase();
    const phone = String(formData.get("phone") || "").trim();
    const participationType = String(formData.get("participation_type") || "onsite").trim() || "onsite";

    if (!name || !email) {
      setFlash("Name und E-Mail sind Pflichtfelder.", "error");
      return;
    }

    setButtonBusy(ui.manualRegistrationSubmitBtn, true, "Speichere...");
    try {
      const payload = await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventID)}/registrations/manual`, {
        method: "POST",
        body: JSON.stringify({
          name,
          email,
          phone,
          participation_type: participationType,
        }),
      });
      ui.manualRegistrationForm.reset();
      const typeSelect = ui.manualRegistrationForm.querySelector("select[name='participation_type']");
      if (typeSelect) {
        typeSelect.value = "onsite";
      }

      await Promise.all([
        loadRegistrations(eventID, false),
        loadDashboard(false),
      ]);
      const status = String(payload?.item?.status || "confirmed");
      setFlash(`Teilnehmer wurde hinzugefuegt (${status}).`);
    } catch (err) {
      setFlash(`Teilnehmer konnte nicht hinzugefuegt werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.manualRegistrationSubmitBtn, false, "Teilnehmer hinzufuegen");
    }
  }

  function renderRegistrations(items) {
    if (!items.length) {
      ui.registrationsTableBody.innerHTML = rowMessage("Keine Teilnehmer fuer dieses Event.", 6);
      return;
    }

    ui.registrationsTableBody.innerHTML = items
      .map((item) => {
        const markAttendedBtn = (item.status === "confirmed" || item.status === "waitlist")
          ? `<button class="btn tiny ok" type="button" data-reg-action="mark-attended" data-reg-id="${escapeAttr(item.id)}">Anwesend</button>`
          : "";

        const issueCertificateBtn = (item.status === "confirmed" || item.status === "attended")
          ? `<button class="btn tiny light" type="button" data-reg-action="issue-certificate" data-reg-id="${escapeAttr(item.id)}">Zertifikat</button>`
          : "";

        return `
          <tr>
            <td>${escapeHTML(item.participant_name || "-")}</td>
            <td>${escapeHTML(item.participant_email || "-")}</td>
            <td>${statusPill(item.status)}</td>
            <td>${escapeHTML(item.participation_type || "-")}</td>
            <td>${escapeHTML(item.payment_status || "-")}</td>
            <td>
              <div class="row-actions">
                ${markAttendedBtn}
                ${issueCertificateBtn}
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    ui.registrationsTableBody.querySelectorAll("button[data-reg-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.regAction || "");
        const registrationID = String(btn.dataset.regId || "");
        if (!registrationID || !action) {
          return;
        }

        setButtonBusy(btn, true, "...");
        try {
          await apiRequest(`/api/v1/admin/registrations/${encodeURIComponent(registrationID)}/${encodeURIComponent(action)}`, {
            method: "POST",
          });
          await Promise.all([
            loadRegistrations(state.selectedEventId, false),
            loadDashboard(false),
          ]);
          setFlash(`Teilnehmer-Aktion '${action}' erfolgreich.`);
        } catch (err) {
          setFlash(`Teilnehmer-Aktion fehlgeschlagen: ${errorMessage(err)}`, "error");
        } finally {
          setButtonBusy(btn, false);
        }
      });
    });
  }

  async function loadSnippets(notify) {
    try {
      const payload = await apiRequest("/api/v1/admin/snippets");
      const items = Array.isArray(payload?.items) ? payload.items.slice() : [];
      items.sort((a, b) => String(a.name || "").localeCompare(String(b.name || "")));
      state.snippets = items;
      renderSnippets(items);
      if (notify) {
        setFlash("Snippet-Liste aktualisiert.");
      }
    } catch (err) {
      setFlash(`Snippets konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderSnippets(items) {
    if (!items.length) {
      ui.snippetsTableBody.innerHTML = rowMessage("Noch keine Snippets vorhanden.", 5);
      return;
    }

    ui.snippetsTableBody.innerHTML = items
      .map((item) => {
        return `
          <tr>
            <td>${escapeHTML(item.name || "-")}</td>
            <td>${escapeHTML(item.slug || "-")}</td>
            <td>${escapeHTML(item.view_type || "-")}</td>
            <td>${item.is_active ? "Ja" : "Nein"}</td>
            <td>
              <div class="row-actions">
                <button class="btn tiny light" type="button" data-snippet-action="embed" data-snippet-id="${escapeAttr(item.id)}">Embed</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    ui.snippetsTableBody.querySelectorAll("button[data-snippet-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.snippetAction || "");
        const snippetID = String(btn.dataset.snippetId || "");
        if (action !== "embed" || !snippetID) {
          return;
        }

        setButtonBusy(btn, true, "...");
        try {
          await loadSnippetEmbedCode(snippetID);
          setFlash("Embed-Code geladen.");
        } catch (err) {
          setFlash(`Embed-Code konnte nicht geladen werden: ${errorMessage(err)}`, "error");
        } finally {
          setButtonBusy(btn, false);
        }
      });
    });
  }

  async function onSnippetCreateSubmit(event) {
    event.preventDefault();
    clearFlash();

    const formData = new FormData(ui.snippetForm);
    const name = String(formData.get("name") || "").trim();
    const providedSlug = String(formData.get("slug") || "").trim();
    const viewType = String(formData.get("view_type") || "cards").trim() || "cards";
    const series = String(formData.get("series") || "").trim();
    const limitRaw = String(formData.get("limit") || "").trim();
    const includePast = ui.snippetForm.querySelector("input[name='include_past']")?.checked === true;
    const isActive = ui.snippetForm.querySelector("input[name='is_active']")?.checked !== false;

    if (!name) {
      setFlash("Snippet-Name ist ein Pflichtfeld.", "error");
      return;
    }

    const slug = providedSlug || slugify(name) || `snippet-${Math.floor(Date.now() / 1000)}`;
    const eventFilter = {};
    if (series) {
      eventFilter.series = series;
    }
    if (includePast) {
      eventFilter.include_past = true;
    }
    if (limitRaw) {
      const parsedLimit = Number(limitRaw);
      if (!Number.isInteger(parsedLimit) || parsedLimit <= 0) {
        setFlash("Snippet-Limit muss eine ganze Zahl > 0 sein.", "error");
        return;
      }
      eventFilter.limit = parsedLimit;
    }

    const body = {
      name,
      slug,
      view_type: viewType,
      event_filter: eventFilter,
      display_options: {},
      is_active: isActive,
    };

    setButtonBusy(ui.snippetSubmitBtn, true, "Speichere...");
    try {
      await apiRequest("/api/v1/admin/snippets", {
        method: "POST",
        body: JSON.stringify(body),
      });
      ui.snippetForm.reset();
      const activeCheckbox = ui.snippetForm.querySelector("input[name='is_active']");
      if (activeCheckbox) {
        activeCheckbox.checked = true;
      }
      await loadSnippets(false);
      setFlash("Snippet wurde angelegt.");
      activateTab("snippets");
    } catch (err) {
      setFlash(`Snippet konnte nicht angelegt werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.snippetSubmitBtn, false, "Snippet speichern");
    }
  }

  async function loadSnippetEmbedCode(snippetID) {
    const payload = await apiRequest(`/api/v1/admin/snippets/${encodeURIComponent(snippetID)}/embed-code`);
    const embedCode = String(payload?.embed_code || "").trim();
    ui.snippetEmbedOutput.value = embedCode;
    if (embedCode) {
      ui.snippetEmbedOutput.focus();
      ui.snippetEmbedOutput.select();
    }
  }

  function statusPill(value) {
    const status = String(value || "-").trim() || "-";
    return `<span class="status-pill" data-status="${escapeAttr(status)}">${escapeHTML(status)}</span>`;
  }

  function rowMessage(message, columnCount) {
    return `<tr><td colspan="${columnCount}">${escapeHTML(message)}</td></tr>`;
  }

  function setFlash(message, type = "info") {
    if (!ui.flash) {
      return;
    }
    ui.flash.hidden = false;
    ui.flash.classList.toggle("error", type === "error");
    ui.flash.textContent = message;
  }

  function clearFlash() {
    if (!ui.flash) {
      return;
    }
    ui.flash.hidden = true;
    ui.flash.classList.remove("error");
    ui.flash.textContent = "";
  }

  function setButtonBusy(button, busy, busyLabel) {
    if (!button) {
      return;
    }
    if (!button.dataset.idleLabel) {
      button.dataset.idleLabel = button.textContent || "";
    }
    button.disabled = !!busy;
    if (busy) {
      button.textContent = busyLabel || "...";
      return;
    }
    button.textContent = button.dataset.idleLabel || button.textContent || "";
  }

  async function apiRequest(path, options = {}) {
    const init = { ...options };
    init.credentials = "include";

    const headers = { ...(options.headers || {}) };
    if (init.body !== undefined && init.body !== null && !headers["Content-Type"]) {
      headers["Content-Type"] = "application/json";
    }
    init.headers = headers;

    const response = await fetch(path, init);
    const contentType = String(response.headers.get("content-type") || "").toLowerCase();

    let payload = null;
    if (response.status !== 204) {
      if (contentType.includes("application/json")) {
        payload = await response.json().catch(() => null);
      } else {
        payload = await response.text().catch(() => "");
      }
    }

    if (!response.ok) {
      const err = new Error(parseErrorMessage(payload) || `HTTP ${response.status}`);
      err.status = response.status;
      err.payload = payload;
      throw err;
    }

    return payload;
  }

  function parseErrorMessage(payload) {
    if (!payload) {
      return "Unbekannter Fehler";
    }
    if (typeof payload === "string") {
      return payload;
    }
    const nested = payload?.error?.message;
    if (nested) {
      return String(nested);
    }
    if (payload?.message) {
      return String(payload.message);
    }
    return "Unbekannter Fehler";
  }

  function errorMessage(err) {
    if (!err) {
      return "Unbekannter Fehler";
    }
    if (typeof err.message === "string" && err.message.trim() !== "") {
      return err.message;
    }
    return "Unbekannter Fehler";
  }

  function formatDateTime(value) {
    if (!value) {
      return "-";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return String(value);
    }
    return new Intl.DateTimeFormat("de-DE", {
      dateStyle: "medium",
      timeStyle: "short",
    }).format(date);
  }

  function toISO(localDateTimeValue) {
    if (!localDateTimeValue) {
      return "";
    }
    const date = new Date(localDateTimeValue);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    return date.toISOString();
  }

  function slugify(value) {
    return String(value || "")
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9\s-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-")
      .replace(/^-|-$/g, "");
  }

  function escapeHTML(value) {
    return String(value ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function escapeAttr(value) {
    return escapeHTML(value).replaceAll("`", "");
  }
})();
