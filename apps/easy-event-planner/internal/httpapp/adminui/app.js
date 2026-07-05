(() => {
  const state = {
    auth: null,
    dashboard: null,
    events: [],
    series: [],
    registrationsByEvent: {},
    selectedEventId: "",
    selectedParticipantEmail: "",
    selectedParticipantName: "",
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
    eventRecurrenceMode: document.querySelector("#eventRecurrenceMode"),
    eventRecurrenceEndMode: document.querySelector("#eventRecurrenceEndMode"),
    eventSubmitBtn: document.querySelector("#eventSubmitBtn"),
    refreshEventsBtn: document.querySelector("#refreshEventsBtn"),
    eventsList: document.querySelector("#eventsList"),
    seriesForm: document.querySelector("#seriesForm"),
    seriesFormHeading: document.querySelector("#seriesFormHeading"),
    seriesFormHint: document.querySelector("#seriesFormHint"),
    seriesCancelEditBtn: document.querySelector("#seriesCancelEditBtn"),
    newSeriesBtn: document.querySelector("#newSeriesBtn"),
    seriesCreateEventBtn: document.querySelector("#seriesCreateEventBtn"),
    seriesSubmitBtn: document.querySelector("#seriesSubmitBtn"),
    refreshSeriesBtn: document.querySelector("#refreshSeriesBtn"),
    seriesList: document.querySelector("#seriesList"),
    seriesEventTabs: document.querySelector("#seriesEventTabs"),
    registrationEventSelect: document.querySelector("#registrationEventSelect"),
    registrationsEventList: document.querySelector("#registrationsEventList"),
    registrationEventTitle: document.querySelector("#registrationEventTitle"),
    registrationEventHint: document.querySelector("#registrationEventHint"),
    refreshRegistrationsBtn: document.querySelector("#refreshRegistrationsBtn"),
    manualRegistrationForm: document.querySelector("#manualRegistrationForm"),
    manualRegistrationSubmitBtn: document.querySelector("#manualRegistrationSubmitBtn"),
    registrationsTableBody: document.querySelector("#registrationsTableBody"),
    participantBookingsTitle: document.querySelector("#participantBookingsTitle"),
    participantBookingsHint: document.querySelector("#participantBookingsHint"),
    participantBookingsSummary: document.querySelector("#participantBookingsSummary"),
    snippetForm: document.querySelector("#snippetForm"),
    snippetSubmitBtn: document.querySelector("#snippetSubmitBtn"),
    refreshSnippetsBtn: document.querySelector("#refreshSnippetsBtn"),
    snippetsTableBody: document.querySelector("#snippetsTableBody"),
    snippetEmbedOutput: document.querySelector("#snippetEmbedOutput"),
    snippetScriptSrcOutput: document.querySelector("#snippetScriptSrcOutput"),
    snippetDataUrlOutput: document.querySelector("#snippetDataUrlOutput"),
    snippetCopyEmbedBtn: document.querySelector("#snippetCopyEmbedBtn"),
    snippetOpenScriptBtn: document.querySelector("#snippetOpenScriptBtn"),
    snippetOpenDataBtn: document.querySelector("#snippetOpenDataBtn"),
    snippetPreviewHint: document.querySelector("#snippetPreviewHint"),
    snippetPreviewFrame: document.querySelector("#snippetPreviewFrame"),
  };

  const STORAGE_TENANT_KEY = "eep_admin_tenant_slug";
  const SCHEDULE_MIN_HOUR = 8;
  const SCHEDULE_MAX_HOUR = 22;
  const DEFAULT_START_TIME = "09:00";

  bindUI();
  fillTimeSelectOptions();
  restoreTenantSlug();
  resetEventForm();
  resetSeriesForm();
  resetParticipantBookings();
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
      bindDateTimeFieldValidation("starts_at", "Startzeit");
      bindDateTimeFieldValidation("ends_at", "Endzeit", true);
    }
    if (ui.eventRecurrenceMode) {
      ui.eventRecurrenceMode.addEventListener("change", syncRecurrenceFields);
    }
    if (ui.eventRecurrenceEndMode) {
      ui.eventRecurrenceEndMode.addEventListener("change", syncRecurrenceFields);
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
    if (ui.newSeriesBtn) {
      ui.newSeriesBtn.addEventListener("click", () => {
        resetSeriesForm();
        renderSeries(state.series);
        setFlash("Neue Serie vorbereitet.");
      });
    }
    if (ui.seriesCreateEventBtn) {
      ui.seriesCreateEventBtn.addEventListener("click", () => {
        if (!state.editingSeriesId) {
          setFlash("Bitte zuerst links eine Serie auswaehlen oder speichern.", "error");
          return;
        }
        startEventCreationFromSeries(state.editingSeriesId);
      });
    }
    if (ui.seriesCancelEditBtn) {
      ui.seriesCancelEditBtn.addEventListener("click", () => {
        resetSeriesForm();
        renderSeries(state.series);
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
    if (ui.snippetCopyEmbedBtn) {
      ui.snippetCopyEmbedBtn.addEventListener("click", onSnippetCopyEmbed);
    }
    if (ui.snippetOpenScriptBtn) {
      ui.snippetOpenScriptBtn.addEventListener("click", () => openSnippetURL(ui.snippetScriptSrcOutput));
    }
    if (ui.snippetOpenDataBtn) {
      ui.snippetOpenDataBtn.addEventListener("click", () => openSnippetURL(ui.snippetDataUrlOutput));
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
    state.dashboard = payload;
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
      ui.nextEventsTableBody.innerHTML = rowMessage("Noch keine kommenden Events.", 6);
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
            <td>
              <div class="row-actions">
                <button class="btn tiny light" type="button" data-dashboard-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>
                <button class="btn tiny ${dashboardVisibilityButtonClass(item.id)}" type="button" data-dashboard-action="toggle-visibility" data-event-id="${escapeAttr(item.id)}"${dashboardVisibilityDisabled(item.id) ? " disabled" : ""}>${escapeHTML(dashboardVisibilityLabel(item.id))}</button>
                <button class="btn tiny light" type="button" data-dashboard-action="focus-registrations" data-event-id="${escapeAttr(item.id)}">Teilnehmer</button>
              </div>
            </td>
          </tr>
        `;
      })
      .join("");

    bindDashboardEventActions();
  }

  async function loadSeries(notify, options) {
    const config = options || {};
    try {
      const payload = await apiRequest("/api/v1/admin/event-series");
      const items = Array.isArray(payload && payload.items) ? payload.items.slice() : [];
      items.sort((a, b) => String(a.title || "").localeCompare(String(b.title || "")));
      state.series = items;
      ensureSeriesEditorSelection(items);
      renderSeries(items);
      fillEventSeriesSelect(currentEventSeriesSelection());
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
    if (!ui.seriesList) {
      return;
    }

    if (!items.length) {
      ui.seriesList.innerHTML = emptyStackMessage("Noch keine Event-Serien vorhanden.");
      renderSeriesEventTabs("");
      updateSeriesActionButtons();
      return;
    }

    ui.seriesList.innerHTML = items
      .map((item) => {
        return renderSeriesCard(item);
      })
      .join("");

    renderSeriesEventTabs(state.editingSeriesId);
    updateSeriesActionButtons();

    ui.seriesList.querySelectorAll("[data-series-card]").forEach((card) => {
      card.addEventListener("click", (ev) => {
        if (ev.target.closest("button")) {
          return;
        }
        const seriesID = String(card.dataset.seriesCard || "");
        const item = findSeriesByID(seriesID);
        if (!item) {
          return;
        }
        populateSeriesFormForEdit(item);
        renderSeries(state.series);
        activateTab("series");
      });
    });

    ui.seriesList.querySelectorAll("button[data-series-action]").forEach((btn) => {
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
          renderSeries(state.series);
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

    ui.seriesList.querySelectorAll("button[data-series-event-id]").forEach((btn) => {
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const eventID = String(btn.dataset.seriesEventId || "");
        openSeriesEvent(eventID);
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
      maybeResetEventEditorAfterReload(items);
      renderEvents(items);
      renderSeries(state.series);
      if (state.dashboard) {
        renderDashboard(state.dashboard);
      }
      fillRegistrationEventSelect(items);
      syncRegistrationSelectionSummary();
      if (notify) {
        setFlash("Events aktualisiert.");
      }
    } catch (err) {
      setFlash(`Events konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderEvents(items) {
    if (!items.length) {
      if (ui.eventsList) {
        ui.eventsList.innerHTML = emptyStackMessage("Noch keine Events vorhanden.");
      }
      if (ui.registrationsEventList) {
        ui.registrationsEventList.innerHTML = emptyStackMessage("Noch keine Events vorhanden.");
      }
      return;
    }

    if (ui.eventsList) {
      ui.eventsList.innerHTML = items.map((item) => renderEventCard(item, "events")).join("");
      bindEventCardInteractions(ui.eventsList, "events");
    }
    if (ui.registrationsEventList) {
      ui.registrationsEventList.innerHTML = items.map((item) => renderEventCard(item, "registrations")).join("");
      bindEventCardInteractions(ui.registrationsEventList, "registrations");
    }
  }

  function renderEventCard(item, context) {
    const isRegistrationContext = context === "registrations";
    const isActive = isRegistrationContext
      ? state.selectedEventId === item.id
      : state.editingEventId === item.id;
    const subtitle = item.subtitle ? `<div class="event-card-subline">${escapeHTML(item.subtitle)}</div>` : "";
    const counts = `${Number(item.confirmed_participants || 0)} Teilnehmer · ${Number(item.waitlist_entries || 0)} Warteliste`;
    const visibilityToggleAllowed = canToggleVisibility(item);
    const visibilityLabel = item.is_public ? "Aktiv" : "Inaktiv";
    const visibilityClass = item.is_public ? "ok" : "light";
    const visibilityAction = item.is_public ? "unpublish" : "publish";
    const seriesTitle = resolveSeriesTitle(item.series_id);

    return `
      <article class="event-card${isActive ? " is-active" : ""}" data-event-card="${escapeAttr(item.id)}" data-event-context="${escapeAttr(context)}">
        <div class="event-card-meta-row">
          <span class="event-card-date">${escapeHTML(formatDateTime(item.starts_at))}</span>
          <button class="btn tiny ${visibilityClass}" type="button" data-event-action="${escapeAttr(visibilityAction)}" data-event-id="${escapeAttr(item.id)}"${visibilityToggleAllowed ? "" : " disabled"}>${escapeHTML(visibilityLabel)}</button>
        </div>
        <div class="event-card-title">${escapeHTML(item.title || "-")}</div>
        ${subtitle}
        <div class="event-card-badges">
          ${statusPill(item.status)}
          ${seriesTitle ? `<span class="series-chip">${escapeHTML(seriesTitle)}</span>` : ""}
        </div>
        <div class="event-card-foot">
          <span class="muted">${escapeHTML(counts)}</span>
          <span class="muted">${escapeHTML(item.slug || "")}</span>
        </div>
        <div class="event-card-actions">
          ${isRegistrationContext
            ? `<button class="btn tiny light" type="button" data-event-action="focus-registrations" data-event-id="${escapeAttr(item.id)}">Teilnehmer</button>
               <button class="btn tiny light" type="button" data-event-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>`
            : `<button class="btn tiny light" type="button" data-event-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>
               <button class="btn tiny light" type="button" data-event-action="focus-registrations" data-event-id="${escapeAttr(item.id)}">Teilnehmer</button>
               <button class="btn tiny warn" type="button" data-event-action="archive" data-event-id="${escapeAttr(item.id)}">Archivieren</button>`}
        </div>
      </article>
    `;
  }

  function bindEventCardInteractions(container, context) {
    container.querySelectorAll("[data-event-card]").forEach((card) => {
      card.addEventListener("click", async (ev) => {
        if (ev.target.closest("button")) {
          return;
        }
        const eventID = String(card.dataset.eventCard || "");
        if (!eventID) {
          return;
        }
        if (context === "registrations") {
          await focusRegistrationsForEvent(eventID);
          return;
        }
        const item = findEventByID(eventID);
        if (item) {
          populateEventFormForEdit(item);
          renderEvents(state.events);
        }
      });
    });

    container.querySelectorAll("button[data-event-action]").forEach((btn) => {
      btn.addEventListener("click", async (ev) => {
        ev.stopPropagation();
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
          renderEvents(state.events);
          return;
        }

        if (action === "focus-registrations") {
          await focusRegistrationsForEvent(id);
          return;
        }

        if (action === "archive") {
          const item = findEventByID(id);
          const label = item ? item.title : "dieses Event";
          const confirmed = window.confirm(`Soll ${label} archiviert werden?`);
          if (!confirmed) {
            return;
          }

          setButtonBusy(btn, true, "...");
          try {
            await apiRequest(`/api/v1/admin/events/${encodeURIComponent(id)}/archive`, {
              method: "POST",
            });
            if (state.editingEventId === id) {
              resetEventForm();
            }
            await Promise.all([loadEvents(false), loadDashboard(false)]);
            setFlash("Event wurde archiviert.");
          } catch (err) {
            setFlash(`Event konnte nicht archiviert werden: ${errorMessage(err)}`, "error");
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
          {
            const currentItem = findEventByID(id);
            setFlash(`Sichtbarkeit fuer '${currentItem ? currentItem.title : "Event"}' wurde aktualisiert.`);
          }
        } catch (err) {
          setFlash(`Event-Aktion fehlgeschlagen: ${errorMessage(err)}`, "error");
        } finally {
          setButtonBusy(btn, false);
        }
      });
    });
  }

  async function focusRegistrationsForEvent(eventID) {
    activateTab("registrations");
    state.selectedEventId = eventID;
    if (ui.registrationEventSelect) {
      ui.registrationEventSelect.value = eventID;
    }
    renderEvents(state.events);
    syncRegistrationSelectionSummary();
    await loadRegistrations(eventID, false);
  }

  function canToggleVisibility(item) {
    return ["cancelled", "completed", "archived"].indexOf(String(item.status || "").trim()) === -1;
  }

  async function onEventSubmit(event) {
    event.preventDefault();
    clearFlash();

    const current = currentEditingEvent();
    const isEdit = !!state.editingEventId;
    let body;
    let recurrencePlan;
    try {
      body = buildEventRequestBody(isEdit, current);
      recurrencePlan = buildRecurrencePlan(isEdit, body);
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
      if (isEdit) {
        await apiRequest(targetPath, {
          method,
          body: JSON.stringify(body),
        });
      } else {
        for (const eventBody of recurrencePlan) {
          await apiRequest("/api/v1/admin/events", {
            method: "POST",
            body: JSON.stringify(eventBody),
          });
        }
      }
      resetEventForm();
      await Promise.all([loadEvents(false), loadDashboard(false)]);
      if (isEdit) {
        setFlash(successMessage);
      } else {
        setFlash(recurrencePlan.length > 1 ? `${recurrencePlan.length} Events wurden angelegt.` : successMessage);
      }
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
    const startsAtLocal = composeLocalDateTimeValue(formData, "starts_at");

    if (!title || !startsAtLocal) {
      throw new Error("Titel und Startzeit sind Pflichtfelder.");
    }

    validateScheduleDateTime(startsAtLocal, "Startzeit");
    const startsAt = toISO(startsAtLocal);
    if (!startsAt) {
      throw new Error("Startzeit ist ungueltig.");
    }

    const endsAtLocal = composeLocalDateTimeValue(formData, "ends_at");
    validateScheduleDateTime(endsAtLocal, "Endzeit", true);
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

  function buildRecurrencePlan(isEdit, baseBody) {
    if (isEdit) {
      return [baseBody];
    }

    const formData = new FormData(ui.eventForm);
    const mode = String(formData.get("recurrence_mode") || "none").trim();
    if (mode === "none") {
      return [baseBody];
    }

    const startsAtLocal = composeLocalDateTimeValue(formData, "starts_at");
    const endsAtLocal = composeLocalDateTimeValue(formData, "ends_at");
    const interval = Number(String(formData.get("recurrence_interval") || "1").trim() || "1");
    if (!Number.isInteger(interval) || interval <= 0) {
      throw new Error("Wiederholungsintervall muss eine ganze Zahl > 0 sein.");
    }

    const endMode = String(formData.get("recurrence_end_mode") || "count").trim();
    const untilRaw = String(formData.get("recurrence_until") || "").trim();
    const countRaw = String(formData.get("recurrence_count") || "").trim();
    const durationMs = calculateDurationMs(startsAtLocal, endsAtLocal);
    const localStarts = [startsAtLocal];
    const maxOccurrences = 120;

    if (endMode === "until") {
      if (!untilRaw) {
        throw new Error("Bitte ein Enddatum fuer die Wiederholung auswaehlen.");
      }
      const untilDate = new Date(`${untilRaw}T23:59:59`);
      if (Number.isNaN(untilDate.getTime())) {
        throw new Error("Das Wiederholungs-Enddatum ist ungueltig.");
      }
      let iteration = 1;
      while (localStarts.length < maxOccurrences) {
        const nextStart = addRecurringLocalDate(startsAtLocal, mode, interval, iteration);
        const nextDate = new Date(nextStart);
        if (nextDate.getTime() > untilDate.getTime()) {
          break;
        }
        localStarts.push(nextStart);
        iteration += 1;
      }
      if (localStarts.length === 1) {
        throw new Error("Die Wiederholung erzeugt mit diesem Enddatum keine zusaetzlichen Termine.");
      }
    } else {
      const count = Number(countRaw || "0");
      if (!Number.isInteger(count) || count < 2) {
        throw new Error("Bitte fuer Wiederholungen mindestens 2 Termine angeben.");
      }
      if (count > maxOccurrences) {
        throw new Error(`Bitte hoechstens ${maxOccurrences} Termine pro Wiederholungsserie anlegen.`);
      }
      for (let iteration = 1; iteration < count; iteration += 1) {
        localStarts.push(addRecurringLocalDate(startsAtLocal, mode, interval, iteration));
      }
    }

    return localStarts.map((startLocal, index) => {
      const nextBody = Object.assign({}, baseBody);
      nextBody.starts_at = toISO(startLocal);
      nextBody.ends_at = durationMs === null ? "" : new Date(new Date(startLocal).getTime() + durationMs).toISOString();
      nextBody.slug = buildRecurringSlug(baseBody.slug, startLocal, index);
      if (index > 0 && !nextBody.change_note) {
        nextBody.change_note = "Automatisch ueber Wiederholungsmodus angelegt.";
      }
      return nextBody;
    });
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
    syncRecurrenceFields();

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
    syncRecurrenceFields();
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
    setDateTimeFieldValue(ui.eventForm, "starts_at", item.starts_at);
    setDateTimeFieldValue(ui.eventForm, "ends_at", item.ends_at);
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
    setFieldValue(ui.eventForm, "recurrence_mode", "none");
    setFieldValue(ui.eventForm, "recurrence_interval", "1");
    setFieldValue(ui.eventForm, "recurrence_end_mode", "count");
    setFieldValue(ui.eventForm, "recurrence_count", "4");
    setFieldValue(ui.eventForm, "recurrence_until", "");

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
    syncRecurrenceFields();
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
    const startsAtField = ui.eventForm ? ui.eventForm.querySelector("input[name='starts_at_date']") : null;
    if (startsAtField) {
      startsAtField.focus();
    }
    setFlash("Event-Formular wurde mit Serien-Standards vorbelegt.");
  }

  function setEventFormDefaults() {
    setFieldValue(ui.eventForm, "timezone", "Europe/Berlin");
    setFieldValue(ui.eventForm, "participation_mode", "onsite");
    setFieldValue(ui.eventForm, "recurrence_mode", "none");
    setFieldValue(ui.eventForm, "recurrence_interval", "1");
    setFieldValue(ui.eventForm, "recurrence_end_mode", "count");
    setFieldValue(ui.eventForm, "recurrence_count", "4");
    setFieldValue(ui.eventForm, "recurrence_until", "");
    setCheckboxValue(ui.eventForm, "is_public", true);
    setCheckboxValue(ui.eventForm, "registration_enabled", true);
    setCheckboxValue(ui.eventForm, "waitlist_enabled", true);
    setFieldValue(ui.eventForm, "change_note", "");
    setFieldValue(ui.eventForm, "starts_at_time", DEFAULT_START_TIME);
    setFieldValue(ui.eventForm, "ends_at_time", "");
  }

  function syncRecurrenceFields() {
    const recurrenceMode = String(ui.eventRecurrenceMode ? ui.eventRecurrenceMode.value : "none").trim();
    const endMode = String(ui.eventRecurrenceEndMode ? ui.eventRecurrenceEndMode.value : "count").trim();
    const isEdit = !!state.editingEventId;
    const recurrenceEnabled = !isEdit && recurrenceMode !== "none";

    setFieldDisabled(ui.eventForm, "recurrence_interval", !recurrenceEnabled);
    setFieldDisabled(ui.eventForm, "recurrence_end_mode", !recurrenceEnabled);
    setFieldDisabled(ui.eventForm, "recurrence_count", !recurrenceEnabled || endMode !== "count");
    setFieldDisabled(ui.eventForm, "recurrence_until", !recurrenceEnabled || endMode !== "until");

    if (isEdit && ui.eventFormHint) {
      const prefix = ui.eventFormHint.textContent || "";
      if (!prefix.includes("Wiederholungsmodus")) {
        ui.eventFormHint.textContent = `${prefix} Wiederholungsmodus wird nur fuer neu angelegte Termine verwendet.`;
      }
    }
  }

  function calculateDurationMs(startsAtLocal, endsAtLocal) {
    if (!endsAtLocal) {
      return null;
    }
    const startsAt = new Date(startsAtLocal);
    const endsAt = new Date(endsAtLocal);
    if (Number.isNaN(startsAt.getTime()) || Number.isNaN(endsAt.getTime())) {
      throw new Error("Start- oder Endzeit ist ungueltig.");
    }
    if (endsAt.getTime() < startsAt.getTime()) {
      throw new Error("Endzeit muss nach der Startzeit liegen.");
    }
    return endsAt.getTime() - startsAt.getTime();
  }

  function addRecurringLocalDate(localDateTimeValue, mode, interval, iteration) {
    const nextDate = new Date(localDateTimeValue);
    if (Number.isNaN(nextDate.getTime())) {
      throw new Error("Die Wiederholungsbasis ist ungueltig.");
    }
    if (mode === "weekly") {
      nextDate.setDate(nextDate.getDate() + (7 * interval * iteration));
    } else if (mode === "monthly") {
      nextDate.setMonth(nextDate.getMonth() + (interval * iteration));
    } else {
      throw new Error("Unbekannter Wiederholungsmodus.");
    }
    return toLocalDateTimeInputValue(nextDate);
  }

  function buildRecurringSlug(baseSlug, localStartValue, index) {
    if (index === 0) {
      return baseSlug;
    }
    return `${baseSlug}-${formatSlugDate(localStartValue)}`;
  }

  function formatSlugDate(localDateTimeValue) {
    const date = new Date(localDateTimeValue);
    if (Number.isNaN(date.getTime())) {
      return `serie-${Date.now()}`;
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    return `${year}-${month}-${day}`;
  }

  function syncRegistrationSelectionSummary() {
    const item = findEventByID(state.selectedEventId);
    if (!item) {
      if (ui.registrationEventTitle) {
        ui.registrationEventTitle.textContent = "Teilnehmer";
      }
      if (ui.registrationEventHint) {
        ui.registrationEventHint.textContent = "Bitte links ein Event auswaehlen.";
      }
      return;
    }

    if (ui.registrationEventTitle) {
      ui.registrationEventTitle.textContent = item.title || "Teilnehmer";
    }
    if (ui.registrationEventHint) {
      ui.registrationEventHint.textContent = `${formatDateTime(item.starts_at)} · ${Number(item.confirmed_participants || 0)} bestaetigt · ${Number(item.waitlist_entries || 0)} Warteliste`;
    }
  }

  function ensureSeriesEditorSelection(items) {
    if (!items.length) {
      resetSeriesForm();
      return;
    }
    const current = state.editingSeriesId
      ? items.find((item) => item.id === state.editingSeriesId)
      : null;
    populateSeriesFormForEdit(current || items[0]);
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
    renderSeriesEventTabs("");
    updateSeriesActionButtons();
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
    renderSeriesEventTabs(item.id);
    updateSeriesActionButtons();
  }

  function fillRegistrationEventSelect(items) {
    const options = items.map((item) => {
      return `<option value="${escapeAttr(item.id)}">${escapeHTML(formatDateTime(item.starts_at))} - ${escapeHTML(item.title || item.slug || item.id)}</option>`;
    });

    if (!options.length) {
      ui.registrationEventSelect.innerHTML = "<option value=''>Keine Events verfuegbar</option>";
      state.selectedEventId = "";
      ui.registrationsTableBody.innerHTML = rowMessage("Bitte zuerst ein Event anlegen.", 6);
      syncRegistrationSelectionSummary();
      return;
    }

    ui.registrationEventSelect.innerHTML = options.join("");

    const stillExists = items.some((item) => item.id === state.selectedEventId);
    if (!stillExists) {
      state.selectedEventId = items[0].id;
    }
    ui.registrationEventSelect.value = state.selectedEventId;

    if (state.selectedEventId) {
      syncRegistrationSelectionSummary();
      loadRegistrations(state.selectedEventId, false);
    }
  }

  async function loadRegistrations(eventID, notify) {
    if (!eventID) {
      ui.registrationsTableBody.innerHTML = rowMessage("Kein Event ausgewaehlt.", 6);
      syncRegistrationSelectionSummary();
      resetParticipantBookings();
      return;
    }

    try {
      state.selectedEventId = eventID;
      syncRegistrationSelectionSummary();
      const payload = await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventID)}/registrations`);
      const items = Array.isArray(payload && payload.items) ? payload.items : [];
      state.registrationsByEvent[eventID] = items;
      renderRegistrations(items);
      refreshParticipantBookingsFromCache();
      renderEvents(state.events);
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
        loadEvents(false),
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
      resetParticipantBookings();
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
          <tr class="registration-row${isSelectedParticipant(item.participant_email) ? " is-active" : ""}" data-reg-row data-reg-email="${escapeAttr(item.participant_email || "")}" data-reg-name="${escapeAttr(item.participant_name || "")}">
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

    ui.registrationsTableBody.querySelectorAll("tr[data-reg-row]").forEach((row) => {
      row.addEventListener("click", async (ev) => {
        if (ev.target.closest("button")) {
          return;
        }
        const email = String(row.dataset.regEmail || "").trim().toLowerCase();
        const name = String(row.dataset.regName || "").trim();
        if (!email) {
          return;
        }
        state.selectedParticipantEmail = email;
        state.selectedParticipantName = name;
        renderRegistrations(items);
        await hydrateParticipantBookings(email, name);
      });
    });

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
            loadEvents(false),
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
            <td>${escapeHTML(item.view_type || "-")}${renderSnippetMeta(item)}</td>
            <td>${item.is_active ? "Ja" : "Nein"}</td>
            <td>
              <div class="row-actions">
                <button class="btn tiny light" type="button" data-snippet-action="embed" data-snippet-id="${escapeAttr(item.id)}">Embed</button>
                <button class="btn tiny light" type="button" data-snippet-action="preview" data-snippet-id="${escapeAttr(item.id)}">Preview</button>
                <button class="btn tiny ${item.is_active ? "warn" : "ok"}" type="button" data-snippet-action="toggle-active" data-snippet-id="${escapeAttr(item.id)}">${item.is_active ? "Deaktivieren" : "Aktivieren"}</button>
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
        if (!action || !snippetID) {
          return;
        }

        setButtonBusy(btn, true, "...");
        try {
          if (action === "toggle-active") {
            await toggleSnippetActive(snippetID);
            setFlash("Snippet-Status wurde aktualisiert.");
          } else {
            const preview = action === "preview";
            await loadSnippetEmbedCode(snippetID, { preview });
            setFlash(preview ? "Snippet-Vorschau geladen." : "Embed-Code geladen.");
          }
        } catch (err) {
          setFlash(`Snippet-Aktion fehlgeschlagen: ${errorMessage(err)}`, "error");
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
    const theme = String(formData.get("theme") || "light").trim() || "light";
    const registerCTA = isChecked(ui.snippetForm, "register");

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
      display_options: {
        theme,
        register: registerCTA,
      },
      is_active: isActive,
    };

    setButtonBusy(ui.snippetSubmitBtn, true, "Speichere...");
    try {
      await apiRequest("/api/v1/admin/snippets", {
        method: "POST",
        body: JSON.stringify(body),
      });
      ui.snippetForm.reset();
      setFieldValue(ui.snippetForm, "theme", "light");
      setCheckboxValue(ui.snippetForm, "register", true);
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

  async function loadSnippetEmbedCode(snippetID, options) {
    const config = options || {};
    const payload = await apiRequest(`/api/v1/admin/snippets/${encodeURIComponent(snippetID)}/embed-code`);
    const embedCode = String(payload && payload.embed_code ? payload.embed_code : "").trim();
    const scriptSrc = String(payload && payload.script_src ? payload.script_src : "").trim();
    const item = payload && payload.item ? payload.item : null;
    const dataURL = scriptSrc ? buildSnippetDataURL(scriptSrc) : "";
    ui.snippetEmbedOutput.value = embedCode;
    if (ui.snippetScriptSrcOutput) {
      ui.snippetScriptSrcOutput.value = scriptSrc;
    }
    if (ui.snippetDataUrlOutput) {
      ui.snippetDataUrlOutput.value = dataURL;
    }
    if (embedCode) {
      ui.snippetEmbedOutput.focus();
      ui.snippetEmbedOutput.select();
    }
    renderSnippetPreview(item, scriptSrc, !!config.preview);
    return payload;
  }

  function statusPill(value) {
    const status = String(value || "-").trim() || "-";
    return `<span class="status-pill" data-status="${escapeAttr(status)}">${escapeHTML(status)}</span>`;
  }

  function rowMessage(message, columnCount) {
    return `<tr><td colspan="${columnCount}">${escapeHTML(message)}</td></tr>`;
  }

  function emptyStackMessage(message) {
    return `<div class="empty-stack">${escapeHTML(message)}</div>`;
  }

  function renderSeriesCard(item) {
    const isActive = state.editingSeriesId === item.id;
    const defaults = [
      item.default_location_name ? `Ort: ${item.default_location_name}` : "",
      item.default_address ? `Adresse: ${item.default_address}` : "",
      item.default_online_url ? `Online: ${item.default_online_url}` : "",
    ].filter(Boolean);
    const seriesEvents = getSeriesEvents(item.id).slice(0, 5);
    const appointmentButtons = seriesEvents.length
      ? `
        <div class="series-appointment-buttons">
          ${seriesEvents.map((eventItem) => `
            <button class="btn tiny ghost" type="button" data-series-event-id="${escapeAttr(eventItem.id)}">${escapeHTML(formatShortDateTime(eventItem.starts_at))}</button>
          `).join("")}
        </div>
      `
      : `<div class="muted compact">Noch keine Termine angelegt.</div>`;

    return `
      <article class="event-card series-card${isActive ? " is-active" : ""}" data-series-card="${escapeAttr(item.id)}">
        <div class="event-card-meta-row">
          <span class="event-card-date">${escapeHTML(item.slug || "")}</span>
          <span class="series-chip">${item.is_public ? "Sichtbar" : "Intern"}</span>
        </div>
        <div class="event-card-title">${escapeHTML(item.title || "-")}</div>
        ${item.description ? `<div class="event-card-subline">${escapeHTML(item.description)}</div>` : ""}
        <div class="series-defaults">
          ${defaults.length ? defaults.map((entry) => `<div class="meta-stack">${escapeHTML(entry)}</div>`).join("") : "<span class='muted'>Keine Standards gesetzt</span>"}
        </div>
        <div class="series-appointments">
          <strong>Naechste Termine</strong>
          ${appointmentButtons}
        </div>
        <div class="event-card-actions">
          <button class="btn tiny light" type="button" data-series-action="edit" data-series-id="${escapeAttr(item.id)}">Bearbeiten</button>
          <button class="btn tiny light" type="button" data-series-action="create-event" data-series-id="${escapeAttr(item.id)}">Termin anlegen</button>
          <button class="btn tiny warn" type="button" data-series-action="delete" data-series-id="${escapeAttr(item.id)}">Loeschen</button>
        </div>
      </article>
    `;
  }

  function getSeriesEvents(seriesID) {
    return state.events
      .filter((item) => String(item.series_id || "").trim() === String(seriesID || "").trim())
      .slice()
      .sort((a, b) => String(a.starts_at || "").localeCompare(String(b.starts_at || "")));
  }

  function renderSeriesEventTabs(seriesID) {
    if (!ui.seriesEventTabs) {
      return;
    }
    if (!seriesID) {
      ui.seriesEventTabs.innerHTML = emptyInlineHint("Serie speichern oder links auswaehlen, um zugehoerige Termine hier als Tabs anzuzeigen.");
      return;
    }
    const seriesEvents = getSeriesEvents(seriesID);
    if (!seriesEvents.length) {
      ui.seriesEventTabs.innerHTML = emptyInlineHint("Zu dieser Serie gibt es noch keine Termine.");
      return;
    }
    ui.seriesEventTabs.innerHTML = seriesEvents.map((item) => {
      const activeClass = state.editingEventId === item.id ? " is-active" : "";
      return `<button class="inline-tab${activeClass}" type="button" data-series-event-id="${escapeAttr(item.id)}">${escapeHTML(formatShortDateTime(item.starts_at))}</button>`;
    }).join("");
    ui.seriesEventTabs.querySelectorAll("button[data-series-event-id]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const eventID = String(btn.dataset.seriesEventId || "");
        openSeriesEvent(eventID);
      });
    });
  }

  function emptyInlineHint(message) {
    return `<div class="inline-hint">${escapeHTML(message)}</div>`;
  }

  function updateSeriesActionButtons() {
    if (ui.seriesCreateEventBtn) {
      ui.seriesCreateEventBtn.disabled = !state.editingSeriesId;
    }
  }

  function openSeriesEvent(eventID) {
    const item = findEventByID(eventID);
    if (!item) {
      setFlash("Termin konnte nicht geladen werden.", "error");
      return;
    }
    populateEventFormForEdit(item);
    renderEvents(state.events);
    renderSeriesEventTabs(String(item.series_id || "").trim());
    activateTab("events");
    setFlash(`Termin '${item.title || item.slug || item.id}' zur Bearbeitung geoeffnet.`);
  }

  async function hydrateParticipantBookings(email, name) {
    renderParticipantBookingsLoading(name || email);
    for (const eventItem of state.events) {
      if (state.registrationsByEvent[eventItem.id]) {
        continue;
      }
      try {
        const payload = await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventItem.id)}/registrations`);
        state.registrationsByEvent[eventItem.id] = Array.isArray(payload && payload.items) ? payload.items : [];
      } catch (err) {
        state.registrationsByEvent[eventItem.id] = [];
        setFlash(`Teilnehmer-Historie konnte nicht vollstaendig geladen werden: ${errorMessage(err)}`, "error");
        break;
      }
    }
    renderParticipantBookings(email, name);
  }

  function refreshParticipantBookingsFromCache() {
    if (!state.selectedParticipantEmail) {
      return;
    }
    renderParticipantBookings(state.selectedParticipantEmail, state.selectedParticipantName);
  }

  function resetParticipantBookings() {
    state.selectedParticipantEmail = "";
    state.selectedParticipantName = "";
    if (ui.participantBookingsTitle) {
      ui.participantBookingsTitle.textContent = "Teilnehmer-Historie";
    }
    if (ui.participantBookingsHint) {
      ui.participantBookingsHint.textContent = "Waehle oben einen Teilnehmer aus, um weitere gebuchte Events zu sehen.";
    }
    if (ui.participantBookingsSummary) {
      ui.participantBookingsSummary.innerHTML = emptyStackMessage("Noch kein Teilnehmer ausgewaehlt.");
    }
  }

  function renderParticipantBookingsLoading(label) {
    if (ui.participantBookingsTitle) {
      ui.participantBookingsTitle.textContent = label || "Teilnehmer-Historie";
    }
    if (ui.participantBookingsHint) {
      ui.participantBookingsHint.textContent = "Weitere Buchungen werden geladen ...";
    }
    if (ui.participantBookingsSummary) {
      ui.participantBookingsSummary.innerHTML = emptyStackMessage("Buchungen werden zusammengesucht ...");
    }
  }

  function renderParticipantBookings(email, name) {
    if (!email) {
      resetParticipantBookings();
      return;
    }
    const bookings = buildParticipantBookings(email);
    if (ui.participantBookingsTitle) {
      ui.participantBookingsTitle.textContent = name || email;
    }
    if (ui.participantBookingsHint) {
      ui.participantBookingsHint.textContent = bookings.length
        ? `${bookings.length} Buchung(en) ueber alle geladenen Events hinweg. Klick auf einen Eintrag springt direkt zum Termin.`
        : "Keine weiteren Buchungen fuer diesen Teilnehmer gefunden.";
    }
    if (!ui.participantBookingsSummary) {
      return;
    }
    if (!bookings.length) {
      ui.participantBookingsSummary.innerHTML = emptyStackMessage("Keine weiteren Buchungen gefunden.");
      return;
    }
    ui.participantBookingsSummary.innerHTML = bookings.map((entry) => {
      return `
        <article class="event-card participant-booking-card" data-booking-event-id="${escapeAttr(entry.event.id)}">
          <div class="event-card-meta-row">
            <span class="event-card-date">${escapeHTML(formatDateTime(entry.event.starts_at))}</span>
            ${statusPill(entry.registration.status)}
          </div>
          <div class="event-card-title">${escapeHTML(entry.event.title || entry.event.slug || entry.event.id)}</div>
          <div class="event-card-foot">
            <span class="muted">${escapeHTML(entry.registration.participation_type || "-")} · Zahlung ${escapeHTML(entry.registration.payment_status || "-")}</span>
            <button class="btn tiny light" type="button" data-booking-open-event="${escapeAttr(entry.event.id)}">Zum Event</button>
          </div>
        </article>
      `;
    }).join("");
    ui.participantBookingsSummary.querySelectorAll("[data-booking-open-event]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const eventID = String(btn.dataset.bookingOpenEvent || "");
        await focusRegistrationsForEvent(eventID);
      });
    });
  }

  function buildParticipantBookings(email) {
    const needle = String(email || "").trim().toLowerCase();
    if (!needle) {
      return [];
    }
    return state.events
      .map((eventItem) => {
        const registrations = state.registrationsByEvent[eventItem.id] || [];
        const registration = registrations.find((item) => String(item.participant_email || "").trim().toLowerCase() === needle);
        return registration ? { event: eventItem, registration } : null;
      })
      .filter(Boolean)
      .sort((a, b) => String(a.event.starts_at || "").localeCompare(String(b.event.starts_at || "")));
  }

  function isSelectedParticipant(email) {
    return !!email && String(email).trim().toLowerCase() === state.selectedParticipantEmail;
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

  function bindDashboardEventActions() {
    if (!ui.nextEventsTableBody) {
      return;
    }
    ui.nextEventsTableBody.querySelectorAll("button[data-dashboard-action]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const action = String(btn.dataset.dashboardAction || "");
        const eventID = String(btn.dataset.eventId || "");
        if (!action || !eventID) {
          return;
        }

        if (action === "edit") {
          await openDashboardEventEdit(eventID);
          return;
        }
        if (action === "focus-registrations") {
          await focusRegistrationsForEvent(eventID);
          return;
        }
        if (action === "toggle-visibility") {
          const item = findEventByID(eventID);
          if (!item) {
            setFlash("Event fuer Sichtbarkeitswechsel nicht gefunden.", "error");
            return;
          }
          const visibilityAction = item.is_public ? "unpublish" : "publish";
          setButtonBusy(btn, true, "...");
          try {
            await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventID)}/${encodeURIComponent(visibilityAction)}`, {
              method: "POST",
            });
            await Promise.all([loadEvents(false), loadDashboard(false)]);
            setFlash(`Sichtbarkeit fuer '${item.title || "Event"}' wurde aktualisiert.`);
          } catch (err) {
            setFlash(`Sichtbarkeitswechsel fehlgeschlagen: ${errorMessage(err)}`, "error");
          } finally {
            setButtonBusy(btn, false);
          }
        }
      });
    });
  }

  async function openDashboardEventEdit(eventID) {
    let item = findEventByID(eventID);
    if (!item) {
      await loadEvents(false);
      item = findEventByID(eventID);
    }
    if (!item) {
      setFlash("Event konnte nicht fuer die Bearbeitung geladen werden.", "error");
      return;
    }
    populateEventFormForEdit(item);
    renderEvents(state.events);
    activateTab("events");
  }

  function dashboardVisibilityLabel(eventID) {
    const item = findEventByID(eventID);
    if (!item) {
      return "Aktiv/Inaktiv";
    }
    return item.is_public ? "Aktiv" : "Inaktiv";
  }

  function dashboardVisibilityButtonClass(eventID) {
    const item = findEventByID(eventID);
    if (!item) {
      return "light";
    }
    return item.is_public ? "ok" : "light";
  }

  function dashboardVisibilityDisabled(eventID) {
    const item = findEventByID(eventID);
    if (!item) {
      return false;
    }
    return !canToggleVisibility(item);
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

  function formatShortDateTime(value) {
    if (!value) {
      return "-";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return String(value);
    }
    return new Intl.DateTimeFormat("de-DE", {
      day: "2-digit",
      month: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
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

  function splitLocalDateTimeValue(value) {
    const localValue = toLocalDateTimeInputValue(value);
    if (!localValue || !localValue.includes("T")) {
      return { date: "", time: "" };
    }
    const parts = localValue.split("T");
    return {
      date: parts[0] || "",
      time: normalizeQuarterHourTime(parts[1] || ""),
    };
  }

  function composeLocalDateTimeValue(formData, fieldBaseName) {
    const date = String(formData.get(`${fieldBaseName}_date`) || "").trim();
    const time = String(formData.get(`${fieldBaseName}_time`) || "").trim();
    if (!date) {
      return "";
    }
    if (!time) {
      return `${date}T`;
    }
    return `${date}T${normalizeQuarterHourTime(time)}`;
  }

  function setDateTimeFieldValue(form, fieldBaseName, value) {
    const parts = splitLocalDateTimeValue(value);
    setFieldValue(form, `${fieldBaseName}_date`, parts.date);
    setFieldValue(form, `${fieldBaseName}_time`, parts.time);
  }

  function fillTimeSelectOptions() {
    fillTimeSelect("starts_at_time", DEFAULT_START_TIME);
    fillTimeSelect("ends_at_time", "", true);
  }

  function fillTimeSelect(fieldName, selectedValue, allowEmpty) {
    const field = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}']`) : null;
    if (!field) {
      return;
    }
    const options = allowEmpty ? ["<option value=''>Keine Endzeit</option>"] : [];
    for (let hour = SCHEDULE_MIN_HOUR; hour <= SCHEDULE_MAX_HOUR; hour += 1) {
      for (const minute of [0, 15, 30, 45]) {
        if (hour === SCHEDULE_MAX_HOUR && minute > 0) {
          continue;
        }
        const value = `${String(hour).padStart(2, "0")}:${String(minute).padStart(2, "0")}`;
        options.push(`<option value="${value}">${value} Uhr</option>`);
      }
    }
    field.innerHTML = options.join("");
    const fallbackValue = allowEmpty ? "" : DEFAULT_START_TIME;
    field.value = selectedValue === "" ? "" : normalizeQuarterHourTime(selectedValue || field.value || fallbackValue);
  }

  function normalizeQuarterHourTime(value) {
    const raw = String(value || "").trim();
    if (!raw) {
      return DEFAULT_START_TIME;
    }
    const match = raw.match(/^(\d{2}):(\d{2})/);
    if (!match) {
      return DEFAULT_START_TIME;
    }
    const hour = Number(match[1]);
    const minute = Number(match[2]);
    const roundedMinute = [0, 15, 30, 45].reduce((best, current) => {
      return Math.abs(current - minute) < Math.abs(best - minute) ? current : best;
    }, 0);
    const clampedHour = Math.max(SCHEDULE_MIN_HOUR, Math.min(SCHEDULE_MAX_HOUR, hour));
    if (clampedHour === SCHEDULE_MAX_HOUR) {
      return "22:00";
    }
    return `${String(clampedHour).padStart(2, "0")}:${String(roundedMinute).padStart(2, "0")}`;
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

  function setFieldDisabled(form, fieldName, disabled) {
    const field = form ? form.querySelector(`[name='${fieldName}']`) : null;
    if (field) {
      field.disabled = !!disabled;
    }
  }

  function bindDateTimeFieldValidation(fieldName, label, allowEmpty) {
    const dateField = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}_date']`) : null;
    const timeField = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}_time']`) : null;
    if (!dateField || !timeField) {
      return;
    }
    const handler = () => {
      const formData = new FormData(ui.eventForm);
      const composedValue = composeLocalDateTimeValue(formData, fieldName);
      const message = getScheduleValidationMessage(composedValue, label, !!allowEmpty);
      dateField.setCustomValidity(message);
      timeField.setCustomValidity(message);
    };
    dateField.addEventListener("change", handler);
    dateField.addEventListener("input", handler);
    timeField.addEventListener("change", handler);
    handler();
  }

  function validateScheduleDateTime(localDateTimeValue, label, allowEmpty) {
    const message = getScheduleValidationMessage(localDateTimeValue, label, !!allowEmpty);
    if (message) {
      throw new Error(message);
    }
  }

  function getScheduleValidationMessage(localDateTimeValue, label, allowEmpty) {
    const value = String(localDateTimeValue || "").trim();
    if (!value) {
      return allowEmpty ? "" : `${label} ist erforderlich.`;
    }
    if (/T$/.test(value)) {
      return `${label} muss mit Datum und Uhrzeit ausgewaehlt werden.`;
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return `${label} ist ungueltig.`;
    }
    const hours = date.getHours();
    const minutes = date.getMinutes();
    if ([0, 15, 30, 45].indexOf(minutes) === -1) {
      return `${label} muss auf Viertelstunden liegen (:00, :15, :30, :45).`;
    }
    if (hours < SCHEDULE_MIN_HOUR || hours > SCHEDULE_MAX_HOUR || (hours === SCHEDULE_MAX_HOUR && minutes > 0)) {
      return `${label} muss zwischen 08:00 und 22:00 Uhr liegen.`;
    }
    return "";
  }

  function isChecked(form, fieldName) {
    const field = form ? form.querySelector(`[name='${fieldName}']`) : null;
    return !!(field && field.checked);
  }

  function renderSnippetMeta(item) {
    const eventFilter = item && item.event_filter ? item.event_filter : {};
    const displayOptions = item && item.display_options ? item.display_options : {};
    const chips = [];
    if (eventFilter.series) {
      chips.push(`Serie ${eventFilter.series}`);
    }
    if (eventFilter.limit) {
      chips.push(`Limit ${eventFilter.limit}`);
    }
    if (displayOptions.theme) {
      chips.push(`Theme ${displayOptions.theme}`);
    }
    if (displayOptions.register) {
      chips.push("CTA");
    }
    if (!chips.length) {
      return "";
    }
    return `<div class="table-subline">${escapeHTML(chips.join(" · "))}</div>`;
  }

  async function toggleSnippetActive(snippetID) {
    const item = state.snippets.find((entry) => entry.id === snippetID);
    if (!item) {
      throw new Error("Snippet nicht gefunden.");
    }
    await apiRequest(`/api/v1/admin/snippets/${encodeURIComponent(snippetID)}`, {
      method: "PATCH",
      body: JSON.stringify({ is_active: !item.is_active }),
    });
    await loadSnippets(false);
  }

  function buildSnippetDataURL(scriptSrc) {
    if (!scriptSrc) {
      return "";
    }
    try {
      const source = new URL(scriptSrc, window.location.origin);
      const tenantSlug = source.pathname.replace(/\/include\.js$/, "").split("/").filter(Boolean).pop();
      const config = source.searchParams.get("config");
      if (!tenantSlug || !config) {
        return "";
      }
      return `${source.origin}/api/v1/public/${encodeURIComponent(tenantSlug)}/snippet/events?config=${encodeURIComponent(config)}`;
    } catch (err) {
      return "";
    }
  }

  function renderSnippetPreview(item, scriptSrc, forcePreview) {
    if (!ui.snippetPreviewFrame || !ui.snippetPreviewHint) {
      return;
    }
    if (!scriptSrc) {
      ui.snippetPreviewHint.textContent = "Lade ein Snippet, um die Einbindung wie in Ghost direkt zu pruefen.";
      ui.snippetPreviewFrame.srcdoc = "";
      return;
    }
    const label = item && item.name ? item.name : "Snippet";
    ui.snippetPreviewHint.textContent = forcePreview
      ? `Live-Vorschau fuer '${label}' geladen.`
      : `Embed fuer '${label}' geladen. Die Vorschau ist darunter sofort einsatzbereit.`;
    ui.snippetPreviewFrame.srcdoc = buildSnippetPreviewDocument(scriptSrc);
  }

  function buildSnippetPreviewDocument(scriptSrc) {
    const safeScriptSrc = escapeAttr(scriptSrc);
    return `<!doctype html>
<html lang="de">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>body{margin:0;padding:16px;background:#faf7f0;font-family:Arial,sans-serif}#eep-preview{min-height:80px}</style>
  </head>
  <body>
    <div id="eep-preview"></div>
    <script src="${safeScriptSrc}" data-target="#eep-preview" defer></script>
  </body>
</html>`;
  }

  async function onSnippetCopyEmbed() {
    const value = String(ui.snippetEmbedOutput ? ui.snippetEmbedOutput.value : "").trim();
    if (!value) {
      setFlash("Bitte zuerst ein Snippet laden.", "error");
      return;
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(value);
      setFlash("Embed-Code wurde in die Zwischenablage kopiert.");
      return;
    }
    ui.snippetEmbedOutput.focus();
    ui.snippetEmbedOutput.select();
    document.execCommand("copy");
    setFlash("Embed-Code wurde in die Zwischenablage kopiert.");
  }

  function openSnippetURL(field) {
    const value = String(field && field.value ? field.value : "").trim();
    if (!value) {
      setFlash("Bitte zuerst ein Snippet laden.", "error");
      return;
    }
    window.open(value, "_blank", "noopener");
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
