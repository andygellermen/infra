(() => {
  const state = {
    auth: null,
    events: [],
    series: [],
    selectedEventId: "",
    snippets: [],
    editingEventId: "",
    editingSeriesId: "",
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
    eventFormHeading: document.querySelector("#eventFormHeading"),
    eventFormHint: document.querySelector("#eventFormHint"),
    eventCancelEditBtn: document.querySelector("#eventCancelEditBtn"),
    eventSeriesSelect: document.querySelector("#eventSeriesSelect"),
    eventSubmitBtn: document.querySelector("#eventSubmitBtn"),
    refreshEventsBtn: document.querySelector("#refreshEventsBtn"),
    eventsTableBody: document.querySelector("#eventsTableBody"),
    seriesForm: document.querySelector("#seriesForm"),
    seriesFormHeading: document.querySelector("#seriesFormHeading"),
    seriesFormHint: document.querySelector("#seriesFormHint"),
    seriesCancelEditBtn: document.querySelector("#seriesCancelEditBtn"),
    seriesSubmitBtn: document.querySelector("#seriesSubmitBtn"),
    refreshSeriesBtn: document.querySelector("#refreshSeriesBtn"),
    seriesTableBody: document.querySelector("#seriesTableBody"),
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
  resetEventForm();
  resetSeriesForm();
  refreshSession();

  function bindUI() {
    if (ui.loginForm) {
      ui.loginForm.addEventListener("submit", onLoginSubmit);
    }
    if (ui.logoutBtn) {
      ui.logoutBtn.addEventListener("click", onLogoutClick);
    }
    if (ui.refreshDashboardBtn) {
      ui.refreshDashboardBtn.addEventListener("click", () => loadDashboard(true));
    }
    if (ui.refreshEventsBtn) {
      ui.refreshEventsBtn.addEventListener("click", () => loadEvents(true));
    }
    if (ui.eventForm) {
      ui.eventForm.addEventListener("submit", onEventSubmit);
    }
    if (ui.eventCancelEditBtn) {
      ui.eventCancelEditBtn.addEventListener("click", () => {
        resetEventForm();
        setFlash("Event-Bearbeitung abgebrochen.");
      });
    }
    if (ui.refreshSeriesBtn) {
      ui.refreshSeriesBtn.addEventListener("click", () => loadSeries(true));
    }
    if (ui.seriesForm) {
      ui.seriesForm.addEventListener("submit", onSeriesSubmit);
    }
    if (ui.seriesCancelEditBtn) {
      ui.seriesCancelEditBtn.addEventListener("click", () => {
        resetSeriesForm();
        setFlash("Serien-Bearbeitung abgebrochen.");
      });
    }
    if (ui.refreshRegistrationsBtn) {
      ui.refreshRegistrationsBtn.addEventListener("click", () => loadRegistrations(state.selectedEventId, true));
    }
    if (ui.manualRegistrationForm) {
      ui.manualRegistrationForm.addEventListener("submit", onManualRegistrationSubmit);
    }
    if (ui.snippetForm) {
      ui.snippetForm.addEventListener("submit", onSnippetCreateSubmit);
    }
    if (ui.refreshSnippetsBtn) {
      ui.refreshSnippetsBtn.addEventListener("click", () => loadSnippets(true));
    }
    if (ui.registrationEventSelect) {
      ui.registrationEventSelect.addEventListener("change", (ev) => {
        const id = String(ev.target.value || "").trim();
        state.selectedEventId = id;
        loadRegistrations(id, false);
      });
    }

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
      await loadSeries(false, { rerenderEvents: false });
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
    const user = me && me.user ? me.user : {};
    const tenant = me && me.tenant ? me.tenant : {};

    ui.loginPanel.hidden = true;
    ui.workspace.hidden = false;
    ui.logoutBtn.hidden = false;
    ui.currentUser.textContent = `${user.name || user.email || "User"} @ ${tenant.slug || "tenant"}`;
    activateTab("dashboard");
  }

  async function onLoginSubmit(event) {
    event.preventDefault();
    clearFlash();

    const tenantSlug = String(ui.tenantSlug ? ui.tenantSlug.value : "").trim();
    const email = String(ui.email ? ui.email.value : "").trim().toLowerCase();

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
    const stats = payload && payload.stats ? payload.stats : {};
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

    const nextEvents = Array.isArray(payload && payload.next_events) ? payload.next_events : [];
    if (nextEvents.length === 0) {
      ui.nextEventsTableBody.innerHTML = rowMessage("Noch keine kommenden Events.", 5);
      return;
    }

    ui.nextEventsTableBody.innerHTML = nextEvents
      .map((item) => {
        const confirmed = item.confirmed_participants === null || item.confirmed_participants === undefined
          ? "-"
          : String(item.confirmed_participants);
        const waitlist = item.waitlist_entries === null || item.waitlist_entries === undefined
          ? "-"
          : String(item.waitlist_entries);
        return `
          <tr>
            <td>${escapeHTML(formatDateTime(item.starts_at))}</td>
            <td>${escapeHTML(item.title || "-")}</td>
            <td>${statusPill(item.status)}</td>
            <td>${escapeHTML(confirmed)}</td>
            <td>${escapeHTML(waitlist)}</td>
          </tr>
        `;
      })
      .join("");
  }

  async function loadSeries(notify, options) {
    const config = options || {};
    try {
      const payload = await apiRequest("/api/v1/admin/event-series");
      const items = Array.isArray(payload && payload.items) ? payload.items.slice() : [];
      items.sort((a, b) => String(a.title || "").localeCompare(String(b.title || "")));
      state.series = items;
      renderSeries(items);
      fillEventSeriesSelect(currentEventSeriesSelection());
      maybeResetSeriesEditorAfterReload(items);
      maybeResetEventSeriesSelectionAfterReload();
      if (config.rerenderEvents !== false) {
        renderEvents(state.events);
      }
      if (notify) {
        setFlash("Event-Serien aktualisiert.");
      }
    } catch (err) {
      setFlash(`Event-Serien konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderSeries(items) {
    if (!items.length) {
      ui.seriesTableBody.innerHTML = rowMessage("Noch keine Event-Serien vorhanden.", 4);
      return;
    }

    ui.seriesTableBody.innerHTML = items
      .map((item) => {
        const defaults = [
          item.default_location_name ? `Ort: ${item.default_location_name}` : "",
          item.default_address ? `Adresse: ${item.default_address}` : "",
          item.default_online_url ? `Online: ${item.default_online_url}` : "",
        ].filter(Boolean);

        return `
          <tr>
            <td>
              <strong>${escapeHTML(item.title || "-")}</strong><br>
              <span class="muted">${escapeHTML(item.slug || "")}</span>
              ${item.description ? `<div class="table-subline">${escapeHTML(item.description)}</div>` : ""}
            </td>
            <td>${defaults.length ? defaults.map((entry) => `<div class="meta-stack">${escapeHTML(entry)}</div>`).join("") : "<span class='muted'>Keine Standards gesetzt</span>"}</td>
            <td>${item.is_public ? "Ja" : "Nein"}</td>
            <td>
              <div class="row-actions">
                <button class="btn tiny light" type="button" data-series-action="create-event" data-series-id="${escapeAttr(item.id)}">Termin anlegen</button>
                <button class="btn tiny light" type="button" data-series-action="edit" data-series-id="${escapeAttr(item.id)}">Bearbeiten</button>
                <button class="btn tiny warn" type="button" data-series-action="delete" data-series-id="${escapeAttr(item.id)}">Loeschen</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    ui.seriesTableBody.querySelectorAll("button[data-series-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.seriesAction || "");
        const seriesID = String(btn.dataset.seriesId || "");
        if (!action || !seriesID) {
          return;
        }

        if (action === "edit") {
          const item = findSeriesByID(seriesID);
          if (!item) {
            setFlash("Event-Serie nicht gefunden.", "error");
            return;
          }
          populateSeriesFormForEdit(item);
          activateTab("series");
          return;
        }

        if (action === "create-event") {
          startEventCreationFromSeries(seriesID);
          return;
        }

        if (action === "delete") {
          const item = findSeriesByID(seriesID);
          const label = item ? item.title : "diese Serie";
          const confirmed = window.confirm(`Soll ${label} wirklich geloescht werden?`);
          if (!confirmed) {
            return;
          }

          setButtonBusy(btn, true, "...");
          try {
            await apiRequest(`/api/v1/admin/event-series/${encodeURIComponent(seriesID)}`, {
              method: "DELETE",
            });
            if (state.editingSeriesId === seriesID) {
              resetSeriesForm();
            }
            await Promise.all([loadSeries(false), loadEvents(false)]);
            setFlash("Event-Serie wurde geloescht.");
          } catch (err) {
            setFlash(`Event-Serie konnte nicht geloescht werden: ${errorMessage(err)}`, "error");
          } finally {
            setButtonBusy(btn, false);
          }
        }
      });
    });
  }

  async function onSeriesSubmit(event) {
    event.preventDefault();
    clearFlash();

    const formData = new FormData(ui.seriesForm);
    const title = String(formData.get("title") || "").trim();
    if (!title) {
      setFlash("Serien-Titel ist ein Pflichtfeld.", "error");
      return;
    }

    const providedSlug = String(formData.get("slug") || "").trim();
    const body = {
      slug: providedSlug || slugify(title) || `series-${Math.floor(Date.now() / 1000)}`,
      title,
      description: String(formData.get("description") || "").trim(),
      default_location_name: String(formData.get("default_location_name") || "").trim(),
      default_address: String(formData.get("default_address") || "").trim(),
      default_online_url: String(formData.get("default_online_url") || "").trim(),
      is_public: isChecked(ui.seriesForm, "is_public"),
    };

    const isEdit = !!state.editingSeriesId;
    const targetPath = isEdit
      ? `/api/v1/admin/event-series/${encodeURIComponent(state.editingSeriesId)}`
      : "/api/v1/admin/event-series";
    const method = isEdit ? "PATCH" : "POST";
    const successMessage = isEdit ? "Event-Serie wurde aktualisiert." : "Event-Serie wurde angelegt.";

    setButtonBusy(ui.seriesSubmitBtn, true, "Speichere...");
    try {
      await apiRequest(targetPath, {
        method,
        body: JSON.stringify(body),
      });
      resetSeriesForm();
      await loadSeries(false);
      setFlash(successMessage);
      activateTab("series");
    } catch (err) {
      setFlash(`Event-Serie konnte nicht gespeichert werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.seriesSubmitBtn, false, isEdit ? "Aenderungen speichern" : "Serie speichern");
    }
  }

  async function loadEvents(notify) {
    try {
      const payload = await apiRequest("/api/v1/admin/events");
      const items = Array.isArray(payload && payload.items) ? payload.items.slice() : [];
      items.sort((a, b) => String(a.starts_at || "").localeCompare(String(b.starts_at || "")));
      state.events = items;
      renderEvents(items);
      fillRegistrationEventSelect(items);
      maybeResetEventEditorAfterReload(items);
      if (notify) {
        setFlash("Events aktualisiert.");
      }
    } catch (err) {
      setFlash(`Events konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderEvents(items) {
    if (!items.length) {
      ui.eventsTableBody.innerHTML = rowMessage("Noch keine Events vorhanden.", 6);
      return;
    }

    ui.eventsTableBody.innerHTML = items
      .map((item) => {
        const publishBtn = item.status === "draft"
          ? `<button class="btn tiny ok" type="button" data-event-action="publish" data-event-id="${escapeAttr(item.id)}">Publish</button>`
          : "";
        const unpublishBtn = item.status === "scheduled" || item.status === "postponed"
          ? `<button class="btn tiny warn" type="button" data-event-action="unpublish" data-event-id="${escapeAttr(item.id)}">Unpublish</button>`
          : "";
        const seriesLabel = renderSeriesBadge(item.series_id);
        const subtitle = item.subtitle ? `<div class="table-subline">${escapeHTML(item.subtitle)}</div>` : "";

        return `
          <tr>
            <td>${escapeHTML(formatDateTime(item.starts_at))}</td>
            <td>
              <strong>${escapeHTML(item.title || "-")}</strong><br>
              <span class="muted">${escapeHTML(item.slug || "")}</span>
              ${subtitle}
            </td>
            <td>${seriesLabel}</td>
            <td>${statusPill(item.status)}</td>
            <td>${item.is_public ? "Ja" : "Nein"}</td>
            <td>
              <div class="row-actions">
                <button class="btn tiny light" type="button" data-event-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>
                ${publishBtn}
                ${unpublishBtn}
                <button class="btn tiny light" type="button" data-event-action="focus-registrations" data-event-id="${escapeAttr(item.id)}">Teilnehmer</button>
                <button class="btn tiny warn" type="button" data-event-action="delete" data-event-id="${escapeAttr(item.id)}">Loeschen</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    ui.eventsTableBody.querySelectorAll("button[data-event-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.eventAction || "");
        const id = String(btn.dataset.eventId || "");
        if (!action || !id) {
          return;
        }

        if (action === "edit") {
          const item = findEventByID(id);
          if (!item) {
            setFlash("Event nicht gefunden.", "error");
            return;
          }
          populateEventFormForEdit(item);
          activateTab("events");
          return;
        }

        if (action === "focus-registrations") {
          activateTab("registrations");
          state.selectedEventId = id;
          ui.registrationEventSelect.value = id;
          await loadRegistrations(id, false);
          return;
        }

        if (action === "delete") {
          const item = findEventByID(id);
          const label = item ? item.title : "dieses Event";
          const confirmed = window.confirm(`Soll ${label} wirklich geloescht werden?`);
          if (!confirmed) {
            return;
          }

          setButtonBusy(btn, true, "...");
          try {
            await apiRequest(`/api/v1/admin/events/${encodeURIComponent(id)}`, {
              method: "DELETE",
            });
            if (state.editingEventId === id) {
              resetEventForm();
            }
            await Promise.all([loadEvents(false), loadDashboard(false)]);
            setFlash("Event wurde geloescht.");
          } catch (err) {
            setFlash(`Event konnte nicht geloescht werden: ${errorMessage(err)}`, "error");
          } finally {
            setButtonBusy(btn, false);
          }
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

  async function onEventSubmit(event) {
    event.preventDefault();
    clearFlash();

    const current = currentEditingEvent();
    const isEdit = !!state.editingEventId;
    let body;
    try {
      body = buildEventRequestBody(isEdit, current);
    } catch (err) {
      setFlash(errorMessage(err), "error");
      return;
    }

    const targetPath = isEdit
      ? `/api/v1/admin/events/${encodeURIComponent(state.editingEventId)}`
      : "/api/v1/admin/events";
    const method = isEdit ? "PATCH" : "POST";
    const successMessage = isEdit ? "Event wurde aktualisiert." : "Event wurde angelegt.";

    setButtonBusy(ui.eventSubmitBtn, true, "Speichere...");
    try {
      await apiRequest(targetPath, {
        method,
        body: JSON.stringify(body),
      });
      resetEventForm();
      await Promise.all([loadEvents(false), loadDashboard(false)]);
      setFlash(successMessage);
      activateTab("events");
    } catch (err) {
      setFlash(`Event konnte nicht gespeichert werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.eventSubmitBtn, false, isEdit ? "Aenderungen speichern" : "Event speichern");
    }
  }

  function buildEventRequestBody(isEdit, current) {
    const formData = new FormData(ui.eventForm);
    const title = String(formData.get("title") || "").trim();
    const startsAtLocal = String(formData.get("starts_at") || "").trim();

    if (!title || !startsAtLocal) {
      throw new Error("Titel und Startzeit sind Pflichtfelder.");
    }

    const startsAt = toISO(startsAtLocal);
    if (!startsAt) {
      throw new Error("Startzeit ist ungueltig.");
    }

    const endsAtLocal = String(formData.get("ends_at") || "").trim();
    const endsAt = endsAtLocal ? toISO(endsAtLocal) : "";
    if (endsAtLocal && !endsAt) {
      throw new Error("Endzeit ist ungueltig.");
    }

    const providedSlug = String(formData.get("slug") || "").trim();
    const fallbackSlug = `event-${Math.floor(Date.now() / 1000)}`;
    const slug = providedSlug || slugify(title) || fallbackSlug;
    const maxParticipantsRaw = String(formData.get("max_participants") || "").trim();

    let maxParticipants = null;
    if (maxParticipantsRaw) {
      const parsed = Number(maxParticipantsRaw);
      if (!Number.isInteger(parsed) || parsed <= 0) {
        throw new Error("Maximale Teilnehmerzahl muss eine ganze Zahl > 0 sein.");
      }
      maxParticipants = parsed;
    }

    const body = {
      series_id: String(formData.get("series_id") || "").trim(),
      slug,
      title,
      subtitle: String(formData.get("subtitle") || "").trim(),
      description: String(formData.get("description") || "").trim(),
      starts_at: startsAt,
      ends_at: endsAt,
      timezone: String(formData.get("timezone") || "Europe/Berlin").trim() || "Europe/Berlin",
      location_name: String(formData.get("location_name") || "").trim(),
      address: String(formData.get("address") || "").trim(),
      online_url: String(formData.get("online_url") || "").trim(),
      participation_mode: String(formData.get("participation_mode") || "onsite").trim() || "onsite",
      is_public: isChecked(ui.eventForm, "is_public"),
      registration_enabled: isChecked(ui.eventForm, "registration_enabled"),
      waitlist_enabled: isChecked(ui.eventForm, "waitlist_enabled"),
      max_participants: maxParticipants,
      change_note: String(formData.get("change_note") || "").trim(),
    };

    if (isEdit) {
      const hadMaxParticipants = current && current.max_participants !== null && current.max_participants !== undefined;
      body.clear_max_participants = !maxParticipantsRaw && hadMaxParticipants;
    }

    return body;
  }

  function resetEventForm(options) {
    const config = options || {};
    const prefill = config.prefill || null;

    state.editingEventId = "";
    if (ui.eventForm) {
      ui.eventForm.reset();
    }
    fillEventSeriesSelect(String(config.seriesID || "").trim());
    setEventFormDefaults();

    if (prefill) {
      setFieldValue(ui.eventForm, "title", prefill.title || "");
      setFieldValue(ui.eventForm, "description", prefill.description || "");
      setFieldValue(ui.eventForm, "location_name", prefill.location_name || "");
      setFieldValue(ui.eventForm, "address", prefill.address || "");
      setFieldValue(ui.eventForm, "online_url", prefill.online_url || "");
      setFieldValue(ui.eventForm, "series_id", prefill.series_id || "");
      setFieldValue(ui.eventForm, "participation_mode", prefill.participation_mode || "onsite");
      setCheckboxValue(ui.eventForm, "is_public", prefill.is_public !== false);
      if (ui.eventFormHint) {
        ui.eventFormHint.textContent = prefill.hint || "Serien-Standards wurden in das Event-Formular uebernommen.";
      }
    } else if (ui.eventFormHint) {
      ui.eventFormHint.textContent = "Neue Einzeltermine oder Termine aus einer Serie anlegen.";
    }

    if (ui.eventFormHeading) {
      ui.eventFormHeading.textContent = "Event anlegen";
    }
    if (ui.eventSubmitBtn) {
      ui.eventSubmitBtn.textContent = "Event speichern";
      ui.eventSubmitBtn.dataset.idleLabel = "Event speichern";
    }
    if (ui.eventCancelEditBtn) {
      ui.eventCancelEditBtn.hidden = true;
    }
  }

  function populateEventFormForEdit(item) {
    state.editingEventId = item.id;
    if (ui.eventForm) {
      ui.eventForm.reset();
    }
    fillEventSeriesSelect(String(item.series_id || "").trim());
    setEventFormDefaults();
    setFieldValue(ui.eventForm, "series_id", String(item.series_id || "").trim());
    setFieldValue(ui.eventForm, "title", item.title || "");
    setFieldValue(ui.eventForm, "slug", item.slug || "");
    setFieldValue(ui.eventForm, "subtitle", item.subtitle || "");
    setFieldValue(ui.eventForm, "description", item.description || "");
    setFieldValue(ui.eventForm, "starts_at", toLocalDateTimeInputValue(item.starts_at));
    setFieldValue(ui.eventForm, "ends_at", toLocalDateTimeInputValue(item.ends_at));
    setFieldValue(ui.eventForm, "timezone", item.timezone || "Europe/Berlin");
    setFieldValue(ui.eventForm, "location_name", item.location_name || "");
    setFieldValue(ui.eventForm, "address", item.address || "");
    setFieldValue(ui.eventForm, "online_url", item.online_url || "");
    setFieldValue(ui.eventForm, "participation_mode", item.participation_mode || "onsite");
    setFieldValue(ui.eventForm, "max_participants", item.max_participants === null || item.max_participants === undefined ? "" : item.max_participants);
    setFieldValue(ui.eventForm, "change_note", item.change_note || "");
    setCheckboxValue(ui.eventForm, "is_public", item.is_public === true);
    setCheckboxValue(ui.eventForm, "registration_enabled", item.registration_enabled !== false);
    setCheckboxValue(ui.eventForm, "waitlist_enabled", item.waitlist_enabled !== false);

    if (ui.eventFormHeading) {
      ui.eventFormHeading.textContent = "Event bearbeiten";
    }
    if (ui.eventFormHint) {
      const seriesTitle = resolveSeriesTitle(item.series_id);
      ui.eventFormHint.textContent = seriesTitle
        ? `Aktuell zugeordnet zu ${seriesTitle}. Leere Felder wie Ende, Reihe oder max. Teilnehmerzahl werden beim Speichern sauber entfernt.`
        : "Bestehendes Event bearbeiten. Leere Felder wie Ende oder max. Teilnehmerzahl werden beim Speichern sauber entfernt.";
    }
    if (ui.eventSubmitBtn) {
      ui.eventSubmitBtn.textContent = "Aenderungen speichern";
      ui.eventSubmitBtn.dataset.idleLabel = "Aenderungen speichern";
    }
    if (ui.eventCancelEditBtn) {
      ui.eventCancelEditBtn.hidden = false;
    }
  }

  function startEventCreationFromSeries(seriesID) {
    const item = findSeriesByID(seriesID);
    if (!item) {
      setFlash("Event-Serie nicht gefunden.", "error");
      return;
    }

    resetEventForm({
      seriesID,
      prefill: {
        series_id: item.id,
        title: item.title || "",
        description: item.description || "",
        location_name: item.default_location_name || "",
        address: item.default_address || "",
        online_url: item.default_online_url || "",
        participation_mode: deriveParticipationMode(item),
        is_public: item.is_public !== false,
        hint: `Neuer Termin auf Basis der Serie '${item.title || item.slug || item.id}'. Bitte Datum, Titel und ggf. Paketinhalte ergaenzen.`,
      },
    });
    activateTab("events");
    const startsAtField = ui.eventForm ? ui.eventForm.querySelector("input[name='starts_at']") : null;
    if (startsAtField) {
      startsAtField.focus();
    }
    setFlash("Event-Formular wurde mit Serien-Standards vorbelegt.");
  }

  function setEventFormDefaults() {
    setFieldValue(ui.eventForm, "timezone", "Europe/Berlin");
    setFieldValue(ui.eventForm, "participation_mode", "onsite");
    setCheckboxValue(ui.eventForm, "is_public", true);
    setCheckboxValue(ui.eventForm, "registration_enabled", true);
    setCheckboxValue(ui.eventForm, "waitlist_enabled", true);
    setFieldValue(ui.eventForm, "change_note", "");
  }

  function currentEventSeriesSelection() {
    if (state.editingEventId) {
      const current = currentEditingEvent();
      if (current) {
        return String(current.series_id || "").trim();
      }
    }
    return String(ui.eventSeriesSelect ? ui.eventSeriesSelect.value : "").trim();
  }

  function maybeResetEventEditorAfterReload(items) {
    if (!state.editingEventId) {
      return;
    }
    const current = items.find((item) => item.id === state.editingEventId);
    if (!current) {
      resetEventForm();
      return;
    }
    populateEventFormForEdit(current);
  }

  function maybeResetSeriesEditorAfterReload(items) {
    if (!state.editingSeriesId) {
      return;
    }
    const current = items.find((item) => item.id === state.editingSeriesId);
    if (!current) {
      resetSeriesForm();
      return;
    }
    populateSeriesFormForEdit(current);
  }

  function maybeResetEventSeriesSelectionAfterReload() {
    if (state.editingEventId) {
      const current = currentEditingEvent();
      fillEventSeriesSelect(current ? String(current.series_id || "").trim() : "");
      return;
    }
    const selected = String(ui.eventSeriesSelect ? ui.eventSeriesSelect.value : "").trim();
    const exists = !selected || state.series.some((item) => item.id === selected);
    if (!exists) {
      fillEventSeriesSelect("");
    }
  }

  function fillEventSeriesSelect(selectedValue) {
    if (!ui.eventSeriesSelect) {
      return;
    }
    const options = ["<option value=''>Keine Serie</option>"];
    state.series.forEach((item) => {
      options.push(`<option value="${escapeAttr(item.id)}">${escapeHTML(item.title || item.slug || item.id)}</option>`);
    });
    ui.eventSeriesSelect.innerHTML = options.join("");
    ui.eventSeriesSelect.value = selectedValue || "";
  }

  function renderSeriesBadge(seriesID) {
    const title = resolveSeriesTitle(seriesID);
    if (!title) {
      return "<span class='muted'>—</span>";
    }
    return `<span class="series-chip">${escapeHTML(title)}</span>`;
  }

  function resolveSeriesTitle(seriesID) {
    const item = findSeriesByID(seriesID);
    return item ? item.title || item.slug || "" : "";
  }

  function deriveParticipationMode(seriesItem) {
    const hasLocation = !!String(seriesItem && seriesItem.default_location_name || "").trim();
    const hasOnlineURL = !!String(seriesItem && seriesItem.default_online_url || "").trim();
    if (hasLocation && hasOnlineURL) {
      return "hybrid";
    }
    if (hasOnlineURL) {
      return "online";
    }
    return "onsite";
  }

  function findSeriesByID(seriesID) {
    return state.series.find((item) => item.id === String(seriesID || "").trim()) || null;
  }

  function findEventByID(eventID) {
    return state.events.find((item) => item.id === String(eventID || "").trim()) || null;
  }

  function currentEditingEvent() {
    return findEventByID(state.editingEventId);
  }

  function resetSeriesForm() {
    state.editingSeriesId = "";
    if (ui.seriesForm) {
      ui.seriesForm.reset();
    }
    setCheckboxValue(ui.seriesForm, "is_public", true);
    if (ui.seriesFormHeading) {
      ui.seriesFormHeading.textContent = "Event-Serie anlegen";
    }
    if (ui.seriesFormHint) {
      ui.seriesFormHint.textContent = "Vorlagen fuer mehrtaegige oder wiederkehrende Formate pflegen.";
    }
    if (ui.seriesSubmitBtn) {
      ui.seriesSubmitBtn.textContent = "Serie speichern";
      ui.seriesSubmitBtn.dataset.idleLabel = "Serie speichern";
    }
    if (ui.seriesCancelEditBtn) {
      ui.seriesCancelEditBtn.hidden = true;
    }
  }

  function populateSeriesFormForEdit(item) {
    state.editingSeriesId = item.id;
    if (ui.seriesForm) {
      ui.seriesForm.reset();
    }
    setFieldValue(ui.seriesForm, "title", item.title || "");
    setFieldValue(ui.seriesForm, "slug", item.slug || "");
    setFieldValue(ui.seriesForm, "description", item.description || "");
    setFieldValue(ui.seriesForm, "default_location_name", item.default_location_name || "");
    setFieldValue(ui.seriesForm, "default_address", item.default_address || "");
    setFieldValue(ui.seriesForm, "default_online_url", item.default_online_url || "");
    setCheckboxValue(ui.seriesForm, "is_public", item.is_public !== false);
    if (ui.seriesFormHeading) {
      ui.seriesFormHeading.textContent = "Event-Serie bearbeiten";
    }
    if (ui.seriesFormHint) {
      ui.seriesFormHint.textContent = "Die Serienwerte dienen als praktische Vorlage fuer neue Event-Termine.";
    }
    if (ui.seriesSubmitBtn) {
      ui.seriesSubmitBtn.textContent = "Aenderungen speichern";
      ui.seriesSubmitBtn.dataset.idleLabel = "Aenderungen speichern";
    }
    if (ui.seriesCancelEditBtn) {
      ui.seriesCancelEditBtn.hidden = false;
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
      const items = Array.isArray(payload && payload.items) ? payload.items : [];
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

    const eventID = String(state.selectedEventId || (ui.registrationEventSelect ? ui.registrationEventSelect.value : "")).trim();
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
      const item = payload && payload.item ? payload.item : null;
      const status = String(item && item.status ? item.status : "confirmed");
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
        const markAttendedBtn = item.status === "confirmed" || item.status === "waitlist"
          ? `<button class="btn tiny ok" type="button" data-reg-action="mark-attended" data-reg-id="${escapeAttr(item.id)}">Anwesend</button>`
          : "";

        const issueCertificateBtn = item.status === "confirmed" || item.status === "attended"
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
      const items = Array.isArray(payload && payload.items) ? payload.items.slice() : [];
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
    const includePast = isChecked(ui.snippetForm, "include_past");
    const isActive = isChecked(ui.snippetForm, "is_active");

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
      setCheckboxValue(ui.snippetForm, "is_active", true);
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
    const embedCode = String(payload && payload.embed_code ? payload.embed_code : "").trim();
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

  function setFlash(message, type) {
    const flashType = type || "info";
    if (!ui.flash) {
      return;
    }
    ui.flash.hidden = false;
    ui.flash.classList.toggle("error", flashType === "error");
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

  async function apiRequest(path, options) {
    const init = Object.assign({}, options || {});
    init.credentials = "include";

    const headers = Object.assign({}, (options && options.headers) || {});
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
    if (payload.error && payload.error.message) {
      return String(payload.error.message);
    }
    if (payload.message) {
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

  function toLocalDateTimeInputValue(value) {
    if (!value) {
      return "";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "";
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  }

  function setFieldValue(form, fieldName, value) {
    const field = form ? form.querySelector(`[name='${fieldName}']`) : null;
    if (field) {
      field.value = value === null || value === undefined ? "" : value;
    }
  }

  function setCheckboxValue(form, fieldName, checked) {
    const field = form ? form.querySelector(`[name='${fieldName}']`) : null;
    if (field) {
      field.checked = !!checked;
    }
  }

  function isChecked(form, fieldName) {
    const field = form ? form.querySelector(`[name='${fieldName}']`) : null;
    return !!(field && field.checked);
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
    return String(value === null || value === undefined ? "" : value)
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
