(() => {
  const state = {
    auth: null,
    tenantProfile: null,
    tenantSettings: null,
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
    editingSnippetId: "",
  };

  const ui = {
    layout: document.querySelector(".layout"),
    topbar: document.querySelector("#topbar"),
    loginPanel: document.querySelector("#loginPanel"),
    workspace: document.querySelector("#workspace"),
    loginForm: document.querySelector("#loginForm"),
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
    eventPublicationHint: document.querySelector("#eventPublicationHint"),
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
    snippetFormHeading: document.querySelector("#snippetFormHeading"),
    snippetCancelEditBtn: document.querySelector("#snippetCancelEditBtn"),
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
    settingsProfileForm: document.querySelector("#settingsProfileForm"),
    settingsProfileSubmitBtn: document.querySelector("#settingsProfileSubmitBtn"),
    settingsRulesForm: document.querySelector("#settingsRulesForm"),
    settingsRulesSubmitBtn: document.querySelector("#settingsRulesSubmitBtn"),
    settingsRulesHint: document.querySelector("#settingsRulesHint"),
    eventDetailBaseUrlHint: document.querySelector("#eventDetailBaseUrlHint"),
  };

  bindUI();
  fillTimeSelectOptions();
  resetEventForm();
  resetSeriesForm();
  resetSnippetForm();
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
      const eventPublicField = ui.eventForm.querySelector("input[name='is_public']");
      if (eventPublicField) {
        eventPublicField.addEventListener("change", () => {
          updateEventPublicationHint(currentEditingEvent());
        });
      }
      const registrationEnabledField = ui.eventForm.querySelector("input[name='registration_enabled']");
      if (registrationEnabledField) {
        registrationEnabledField.addEventListener("change", () => {
          applyRegistrationActivationShortcut();
        });
      }
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
    if (ui.snippetCancelEditBtn) {
      ui.snippetCancelEditBtn.addEventListener("click", () => {
        resetSnippetForm();
        setFlash("Snippet-Bearbeitung abgebrochen.");
      });
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
    if (ui.settingsProfileForm) {
      ui.settingsProfileForm.addEventListener("submit", onTenantProfileSubmit);
      const publicBaseField = ui.settingsProfileForm.querySelector("[name='public_base_url']");
      if (publicBaseField) {
        publicBaseField.addEventListener("input", updateEventDetailBaseURLHint);
      }
    }
    if (ui.settingsRulesForm) {
      ui.settingsRulesForm.addEventListener("submit", onTenantSettingsSubmit);
      const detailBaseField = ui.settingsRulesForm.querySelector("[name='event_detail_base_url']");
      if (detailBaseField) {
        detailBaseField.addEventListener("input", updateEventDetailBaseURLHint);
      }
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
      await loadTenantSettings(false);
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
    state.tenantProfile = null;
    state.tenantSettings = null;
    ui.loginPanel.hidden = false;
    ui.workspace.hidden = true;
    ui.logoutBtn.hidden = true;
    if (ui.topbar) {
      ui.topbar.hidden = true;
    }
    if (ui.layout) {
      ui.layout.classList.add("is-login");
    }
    ui.currentUser.textContent = "Nicht angemeldet";
    setLoginHint("");
  }

  function showWorkspace(me) {
    const user = me && me.user ? me.user : {};
    const tenant = me && me.tenant ? me.tenant : {};

    ui.loginPanel.hidden = true;
    ui.workspace.hidden = false;
    ui.logoutBtn.hidden = false;
    if (ui.topbar) {
      ui.topbar.hidden = false;
    }
    if (ui.layout) {
      ui.layout.classList.remove("is-login");
    }
    ui.currentUser.textContent = `${user.name || user.email || "User"} @ ${tenant.slug || "tenant"}`;
    activateTab("dashboard");
  }

  async function onLoginSubmit(event) {
    event.preventDefault();
    clearFlash();

    const email = String(ui.email ? ui.email.value : "").trim().toLowerCase();

    if (!email) {
      setLoginHint("Bitte E-Mail eingeben.", "error");
      return;
    }

    setLoginHint("");
    setButtonBusy(ui.loginSubmitBtn, true, "Sende...");

    try {
      await apiRequest("/api/v1/auth/magic-link/request", {
        method: "POST",
        body: JSON.stringify({
          email,
          purpose: "organizer_login",
          redirect_path: "/admin",
        }),
      });
      setLoginHint("Link ist unterwegs.", "success");
    } catch (err) {
      setLoginHint(`Nicht gesendet: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.loginSubmitBtn, false, "Link senden");
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

  async function loadTenantSettings(notify) {
    try {
      const [tenantPayload, settingsPayload] = await Promise.all([
        apiRequest("/api/v1/admin/tenant"),
        apiRequest("/api/v1/admin/tenant/settings"),
      ]);
      state.tenantProfile = tenantPayload && tenantPayload.item ? tenantPayload.item : null;
      state.tenantSettings = settingsPayload && settingsPayload.item ? settingsPayload.item : null;
      renderTenantSettings();
      if (notify) {
        setFlash("Settings aktualisiert.");
      }
    } catch (err) {
      setFlash(`Settings konnten nicht geladen werden: ${errorMessage(err)}`, "error");
    }
  }

  function renderTenantSettings() {
    const tenantItem = state.tenantProfile || {};
    const settingsItem = state.tenantSettings || {};
    const appSettings = getAppSettings();

    setFieldValue(ui.settingsProfileForm, "slug", tenantItem.slug || "");
    setFieldValue(ui.settingsProfileForm, "name", tenantItem.name || "");
    setFieldValue(ui.settingsProfileForm, "public_base_url", tenantItem.public_base_url || "");
    setFieldValue(ui.settingsProfileForm, "default_timezone", tenantItem.default_timezone || "Europe/Berlin");
    setFieldValue(ui.settingsProfileForm, "default_locale", tenantItem.default_locale || "de-DE");

    setFieldValue(ui.settingsRulesForm, "event_time_start", appSettings.event_time_start);
    setFieldValue(ui.settingsRulesForm, "event_time_end", appSettings.event_time_end);
    setFieldValue(ui.settingsRulesForm, "event_time_step_minutes", String(appSettings.event_time_step_minutes));
    setFieldValue(ui.settingsRulesForm, "event_slug_mode", appSettings.event_slug_mode);
    setFieldValue(ui.settingsRulesForm, "event_detail_base_url", appSettings.event_detail_base_url || "");
    setFieldValue(ui.settingsRulesForm, "participant_cancel_deadline_hours", String(appSettings.participant_cancel_deadline_hours));
    setFieldValue(ui.settingsRulesForm, "sender_email", settingsItem.sender_email || "");
    setFieldValue(ui.settingsRulesForm, "sender_name", settingsItem.sender_name || "");
    setFieldValue(ui.settingsRulesForm, "default_retention_days", settingsItem.default_retention_days || 30);
    setFieldValue(ui.settingsRulesForm, "allowed_embed_origins", (appSettings.allowed_embed_origins || []).join("\n"));

    updateEventDetailBaseURLHint();
    fillTimeSelectOptions();
    applySteppedDateTimeInputConfig();
    applyEventSlugMode();
    validateEventScheduleFields();
  }

  async function onTenantProfileSubmit(event) {
    event.preventDefault();
    clearFlash();

    const formData = new FormData(ui.settingsProfileForm);
    const body = {
      name: String(formData.get("name") || "").trim(),
      public_base_url: String(formData.get("public_base_url") || "").trim(),
      default_timezone: String(formData.get("default_timezone") || "").trim(),
      default_locale: String(formData.get("default_locale") || "").trim(),
    };

    if (!body.name || !body.public_base_url || !body.default_timezone || !body.default_locale) {
      setFlash("Bitte alle Profilfelder ausfuellen.", "error");
      return;
    }

    setButtonBusy(ui.settingsProfileSubmitBtn, true, "Speichere...");
    try {
      await apiRequest("/api/v1/admin/tenant", {
        method: "PATCH",
        body: JSON.stringify(body),
      });
      await loadTenantSettings(false);
      setFlash("Mandantenprofil wurde aktualisiert.");
    } catch (err) {
      setFlash(`Mandantenprofil konnte nicht gespeichert werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.settingsProfileSubmitBtn, false, "Profil speichern");
    }
  }

  async function onTenantSettingsSubmit(event) {
    event.preventDefault();
    clearFlash();

    const formData = new FormData(ui.settingsRulesForm);
    const retention = Number(String(formData.get("default_retention_days") || "0").trim() || "0");
    const cancelDeadlineHours = Number(String(formData.get("participant_cancel_deadline_hours") || "24").trim() || "24");
    const appSettings = {
      event_time_start: String(formData.get("event_time_start") || "").trim(),
      event_time_end: String(formData.get("event_time_end") || "").trim(),
      event_time_step_minutes: Number(String(formData.get("event_time_step_minutes") || "15").trim() || "15"),
      event_slug_mode: String(formData.get("event_slug_mode") || "optional").trim(),
      event_detail_base_url: String(formData.get("event_detail_base_url") || "").trim(),
      participant_cancel_deadline_hours: cancelDeadlineHours,
      allowed_embed_origins: parseOriginsTextarea(String(formData.get("allowed_embed_origins") || "")),
    };

    try {
      validateSettingsSchedule(appSettings);
    } catch (err) {
      setFlash(errorMessage(err), "error");
      return;
    }
    if (!Number.isInteger(retention) || retention <= 0) {
      setFlash("Aufbewahrungstage muessen eine ganze Zahl > 0 sein.", "error");
      return;
    }
    if (!Number.isInteger(cancelDeadlineHours) || cancelDeadlineHours < 0) {
      setFlash("Die Abmeldefrist muss eine ganze Zahl >= 0 sein.", "error");
      return;
    }

    setButtonBusy(ui.settingsRulesSubmitBtn, true, "Speichere...");
    try {
      await apiRequest("/api/v1/admin/tenant/settings", {
        method: "PATCH",
        body: JSON.stringify({
          sender_email: String(formData.get("sender_email") || "").trim(),
          sender_name: String(formData.get("sender_name") || "").trim(),
          default_retention_days: retention,
          app_settings: appSettings,
        }),
      });
      await loadTenantSettings(false);
      setFlash("EEP-Regeln wurden aktualisiert.");
    } catch (err) {
      setFlash(`EEP-Regeln konnten nicht gespeichert werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.settingsRulesSubmitBtn, false, "Settings speichern");
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
    const publicationMeta = getPublicationMeta(item);
    const visibilityLabel = item.is_published ? "Verbergen" : "Freigeben";
    const visibilityClass = item.is_published ? "ok" : "light";
    const visibilityAction = item.is_published ? "unpublish" : "publish";
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
          <span class="series-chip">${escapeHTML(publicationMeta.label)}</span>
          ${seriesTitle ? `<span class="series-chip">${escapeHTML(seriesTitle)}</span>` : ""}
        </div>
        <div class="event-card-foot">
          <span class="muted">${escapeHTML(`${counts} · ${publicationMeta.detail}`)}</span>
          <span class="muted">${escapeHTML(item.slug || "")}</span>
        </div>
        <div class="event-card-actions">
          ${isRegistrationContext
            ? `<button class="btn tiny light" type="button" data-event-action="focus-registrations" data-event-id="${escapeAttr(item.id)}">Teilnehmer</button>
               <button class="btn tiny light" type="button" data-event-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>`
            : `<button class="btn tiny light" type="button" data-event-action="edit" data-event-id="${escapeAttr(item.id)}">Bearbeiten</button>
               <button class="btn tiny light" type="button" data-event-action="embed" data-event-id="${escapeAttr(item.id)}">Embed</button>
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

        if (action === "embed") {
          await loadEventRegistrationEmbedCode(id);
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
            setFlash(`Freigabe fuer '${currentItem ? currentItem.title : "Event"}' wurde aktualisiert.`);
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
    const scheduleConfig = getScheduleConfig();
    let slug = providedSlug;
    if (scheduleConfig.event_slug_mode === "required" && !slug) {
      throw new Error("Bitte einen Event-Slug vergeben oder den Slug-Modus in den Settings anpassen.");
    }
    if (scheduleConfig.event_slug_mode === "auto" || !slug) {
      slug = slugify(title) || fallbackSlug;
    }
    const maxParticipantsRaw = String(formData.get("max_participants") || "").trim();

    let maxParticipants = null;
    if (maxParticipantsRaw) {
      const parsed = Number(maxParticipantsRaw);
      if (!Number.isInteger(parsed) || parsed <= 0) {
        throw new Error("Maximale Teilnehmerzahl muss eine ganze Zahl > 0 sein.");
      }
      maxParticipants = parsed;
    }

    const publicVisibleFrom = readOptionalSteppedDateTime(formData, "public_visible_from", "Oeffentlich sichtbar ab");
    const registrationOpensAt = readOptionalSteppedDateTime(formData, "registration_opens_at", "Registrierung moeglich ab");
    const registrationClosesAt = readOptionalSteppedDateTime(formData, "registration_closes_at", "Registrierung moeglich bis");
    if (registrationOpensAt && registrationClosesAt && new Date(registrationClosesAt).getTime() < new Date(registrationOpensAt).getTime()) {
      throw new Error("Registrierung moeglich bis muss nach Registrierung moeglich ab liegen.");
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
      public_visible_from: publicVisibleFrom,
      registration_opens_at: registrationOpensAt,
      registration_closes_at: registrationClosesAt,
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
    updateEventPublicationHint(null);
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
    setFieldValue(ui.eventForm, "public_visible_from", toLocalDateTimeInputValue(item.public_visible_from));
    setFieldValue(ui.eventForm, "registration_opens_at", toLocalDateTimeInputValue(item.registration_opens_at));
    setFieldValue(ui.eventForm, "registration_closes_at", toLocalDateTimeInputValue(item.registration_closes_at));
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
    applyEventSlugMode();
    syncRecurrenceFields();
    updateEventPublicationHint(item);
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
    setFieldValue(ui.eventForm, "public_visible_from", "");
    setFieldValue(ui.eventForm, "registration_opens_at", "");
    setFieldValue(ui.eventForm, "registration_closes_at", "");
    setFieldValue(ui.eventForm, "change_note", "");
    setFieldValue(ui.eventForm, "starts_at_time", getDefaultStartTime());
    setFieldValue(ui.eventForm, "ends_at_time", "");
    applyEventSlugMode();
    updateEventPublicationHint(currentEditingEvent());
  }

  function applyRegistrationActivationShortcut() {
    if (!ui.eventForm) {
      return;
    }
    const registrationEnabledField = ui.eventForm.querySelector("input[name='registration_enabled']");
    if (!registrationEnabledField || !registrationEnabledField.checked) {
      return;
    }
    if (!confirmRegistrationActivationShortcut()) {
      registrationEnabledField.checked = false;
      return;
    }
    const nowValue = currentSteppedLocalDateTimeValue();
    if (!nowValue) {
      return;
    }
    setFieldValue(ui.eventForm, "public_visible_from", nowValue);
    setFieldValue(ui.eventForm, "registration_opens_at", nowValue);
    setFieldValue(ui.eventForm, "registration_closes_at", "");
    updateEventPublicationHint(currentEditingEvent());
  }

  function confirmRegistrationActivationShortcut() {
    const message = [
      "Achtung: 'Anmeldung aktiv' setzt sofort neue Werte.",
      "",
      "Folgende Felder werden ueberschrieben:",
      "- Oeffentlich sichtbar ab = jetzt",
      "- Registrierung moeglich ab = jetzt",
      "- Registrierung moeglich bis = leer",
      "",
      "Moechtest du fortfahren?"
    ].join("\n");
    if (typeof window !== "undefined" && typeof window.confirm === "function") {
      return window.confirm(message);
    }
    return true;
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
    if (mode === "daily") {
      nextDate.setDate(nextDate.getDate() + (interval * iteration));
    } else if (mode === "weekly") {
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

  function resetSnippetForm() {
    state.editingSnippetId = "";
    if (ui.snippetForm) {
      ui.snippetForm.reset();
    }
    setFieldValue(ui.snippetForm, "theme", "light");
    setCheckboxValue(ui.snippetForm, "register", true);
    setCheckboxValue(ui.snippetForm, "load_css", true);
    setCheckboxValue(ui.snippetForm, "is_active", true);
    if (ui.snippetFormHeading) {
      ui.snippetFormHeading.textContent = "Snippet anlegen";
    }
    if (ui.snippetSubmitBtn) {
      ui.snippetSubmitBtn.textContent = "Snippet speichern";
      ui.snippetSubmitBtn.dataset.idleLabel = "Snippet speichern";
    }
    if (ui.snippetCancelEditBtn) {
      ui.snippetCancelEditBtn.hidden = true;
    }
  }

  function populateSnippetFormForEdit(item) {
    state.editingSnippetId = String(item && item.id || "").trim();
    if (ui.snippetForm) {
      ui.snippetForm.reset();
    }
    const eventFilter = item && item.event_filter ? item.event_filter : {};
    const displayOptions = item && item.display_options ? item.display_options : {};
    setFieldValue(ui.snippetForm, "name", item && item.name || "");
    setFieldValue(ui.snippetForm, "slug", item && item.slug || "");
    setFieldValue(ui.snippetForm, "view_type", item && item.view_type || "cards");
    setFieldValue(ui.snippetForm, "series", eventFilter.series || "");
    setFieldValue(ui.snippetForm, "event", eventFilter.event || "");
    setFieldValue(ui.snippetForm, "limit", eventFilter.limit === null || eventFilter.limit === undefined ? "" : eventFilter.limit);
    setFieldValue(ui.snippetForm, "theme", displayOptions.theme || "light");
    setCheckboxValue(ui.snippetForm, "register", displayOptions.register !== false);
    setCheckboxValue(ui.snippetForm, "load_css", displayOptions.load_css !== false);
    setCheckboxValue(ui.snippetForm, "include_past", !!eventFilter.include_past);
    setCheckboxValue(ui.snippetForm, "is_active", item && item.is_active !== false);
    if (ui.snippetFormHeading) {
      ui.snippetFormHeading.textContent = "Snippet bearbeiten";
    }
    if (ui.snippetSubmitBtn) {
      ui.snippetSubmitBtn.textContent = "Aenderungen speichern";
      ui.snippetSubmitBtn.dataset.idleLabel = "Aenderungen speichern";
    }
    if (ui.snippetCancelEditBtn) {
      ui.snippetCancelEditBtn.hidden = false;
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
                <button class="btn tiny light" type="button" data-snippet-action="edit" data-snippet-id="${escapeAttr(item.id)}">Bearbeiten</button>
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
          if (action === "edit") {
            const item = state.snippets.find((entry) => entry.id === snippetID);
            if (!item) {
              throw new Error("Snippet nicht gefunden.");
            }
            populateSnippetFormForEdit(item);
            activateTab("snippets");
            setFlash(`Snippet '${item.name || "Snippet"}' zur Bearbeitung geladen.`);
            return;
          }
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
    const eventSlug = String(formData.get("event") || "").trim();
    const limitRaw = String(formData.get("limit") || "").trim();
    const includePast = isChecked(ui.snippetForm, "include_past");
    const isActive = isChecked(ui.snippetForm, "is_active");
    const theme = String(formData.get("theme") || "light").trim() || "light";
    const registerCTA = isChecked(ui.snippetForm, "register");
    const loadCSS = isChecked(ui.snippetForm, "load_css");
    const isEdit = !!state.editingSnippetId;

    if (!name) {
      setFlash("Snippet-Name ist ein Pflichtfeld.", "error");
      return;
    }

    const slug = providedSlug || slugify(name) || `snippet-${Math.floor(Date.now() / 1000)}`;
    const eventFilter = {};
    if (series) {
      eventFilter.series = series;
    }
    if (eventSlug) {
      eventFilter.event = eventSlug;
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
        load_css: loadCSS,
      },
      is_active: isActive,
    };

    const targetPath = isEdit
      ? `/api/v1/admin/snippets/${encodeURIComponent(state.editingSnippetId)}`
      : "/api/v1/admin/snippets";
    const method = isEdit ? "PATCH" : "POST";

    setButtonBusy(ui.snippetSubmitBtn, true, "Speichere...");
    try {
      await apiRequest(targetPath, {
        method,
        body: JSON.stringify(body),
      });
      resetSnippetForm();
      await loadSnippets(false);
      setFlash(isEdit ? "Snippet wurde aktualisiert." : "Snippet wurde angelegt.");
      activateTab("snippets");
    } catch (err) {
      setFlash(`Snippet konnte nicht gespeichert werden: ${errorMessage(err)}`, "error");
    } finally {
      setButtonBusy(ui.snippetSubmitBtn, false, isEdit ? "Aenderungen speichern" : "Snippet speichern");
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

  async function loadEventRegistrationEmbedCode(eventID) {
    const payload = await apiRequest(`/api/v1/admin/events/${encodeURIComponent(eventID)}/embed-code`);
    const embedCode = String(payload && payload.embed_code ? payload.embed_code : "").trim();
    const scriptSrc = String(payload && payload.script_src ? payload.script_src : "").trim();
    const detailURL = String(payload && payload.event_detail_api_url ? payload.event_detail_api_url : "").trim();
    const item = payload && payload.item ? payload.item : null;
    const warnings = Array.isArray(payload && payload.warnings) ? payload.warnings.filter(Boolean) : [];
    ui.snippetEmbedOutput.value = embedCode;
    if (ui.snippetScriptSrcOutput) {
      ui.snippetScriptSrcOutput.value = scriptSrc;
    }
    if (ui.snippetDataUrlOutput) {
      ui.snippetDataUrlOutput.value = detailURL;
    }
    if (embedCode) {
      ui.snippetEmbedOutput.focus();
      ui.snippetEmbedOutput.select();
    }
    renderSnippetPreview(item, scriptSrc, true);
    activateTab("snippets");
    if (warnings.length) {
      setFlash(`Embed geladen. Hinweis: ${warnings.join(" ")}`);
      return payload;
    }
    setFlash(`Anmeldeformular fuer '${item && item.title ? item.title : "Event"}' geladen.`);
    return payload;
  }

  function statusPill(value) {
    const status = String(value || "-").trim() || "-";
    return `<span class="status-pill" data-status="${escapeAttr(status)}">${escapeHTML(status)}</span>`;
  }

  function publicationStateOf(item) {
    const explicit = String(item && item.publication_state ? item.publication_state : "").trim();
    if (explicit) {
      return explicit;
    }
    if (item && item.is_published) {
      return "published";
    }
    if (item && item.is_public) {
      return "prepared";
    }
    return "internal";
  }

  function getPublicationMeta(item) {
    const stateName = publicationStateOf(item);
    switch (stateName) {
      case "scheduled_publication":
        return {
          label: "Freigegeben ab",
          detail: item && item.public_visible_from ? formatDateTime(item.public_visible_from) : "Noch nicht sichtbar",
        };
      case "published":
        return {
          label: "Oeffentlich live",
          detail: item && item.published_at ? `Live seit ${formatDateTime(item.published_at)}` : "Live geschaltet",
        };
      case "prepared":
        return {
          label: "Fuer Freigabe vorbereitet",
          detail: "Wird erst ueber 'Freigeben' sichtbar",
        };
      case "archived":
        return {
          label: "Archiviert",
          detail: "Nicht mehr oeffentlich sichtbar",
        };
      default:
        return {
          label: "Nur intern",
          detail: "Nicht fuer die oeffentliche Uebersicht vorgesehen",
        };
    }
  }

  function updateEventPublicationHint(item) {
    if (!ui.eventPublicationHint || !ui.eventForm) {
      return;
    }
    const isPublicChecked = isChecked(ui.eventForm, "is_public");
    if (item && item.publication_state === "scheduled_publication") {
      ui.eventPublicationHint.textContent = item.public_visible_from
        ? `Dieses Event ist bereits freigegeben und wird ab ${formatDateTime(item.public_visible_from)} automatisch sichtbar. Mit "Verbergen" stoppst du die geplante Freigabe.`
        : "Dieses Event ist bereits freigegeben und wird automatisch sichtbar, sobald der Sichtbarkeitszeitpunkt erreicht ist.";
      return;
    }
    if (item && item.is_published) {
      ui.eventPublicationHint.textContent = item.published_at
        ? `Dieses Event ist aktuell live und seit ${formatDateTime(item.published_at)} freigegeben. Mit "Verbergen" blendest du es aus, ohne die oeffentliche Vorbereitung zu verlieren.`
        : "Dieses Event ist aktuell live. Mit 'Verbergen' blendest du es aus, ohne die oeffentliche Vorbereitung zu verlieren.";
      return;
    }
    if (isPublicChecked) {
      ui.eventPublicationHint.textContent = "Mit Haken ist das Event fuer die oeffentliche Uebersicht vorbereitet. Live geht es erst ueber den Button 'Freigeben'.";
      return;
    }
    ui.eventPublicationHint.textContent = "Ohne Haken bleibt das Event intern und erscheint auch nach einer Freigabe nicht in der oeffentlichen Uebersicht.";
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
          <span class="series-chip">${item.is_public ? "Oeffentlich" : "Intern"}</span>
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

  function setLoginHint(message, type) {
    if (!ui.loginHint) {
      return;
    }
    const tone = String(type || "").trim();
    ui.loginHint.textContent = message || "";
    ui.loginHint.classList.toggle("is-success", tone === "success");
    ui.loginHint.classList.toggle("is-error", tone === "error");
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
          const visibilityAction = item.is_published ? "unpublish" : "publish";
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
      return "Freigeben";
    }
    return item.is_published ? "Verbergen" : "Freigeben";
  }

  function dashboardVisibilityButtonClass(eventID) {
    const item = findEventByID(eventID);
    if (!item) {
      return "light";
    }
    return item.is_published ? "ok" : "light";
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
      time: normalizeScheduleTime(parts[1] || "", false),
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
    return `${date}T${normalizeScheduleTime(time, false)}`;
  }

  function setDateTimeFieldValue(form, fieldBaseName, value) {
    const parts = splitLocalDateTimeValue(value);
    setFieldValue(form, `${fieldBaseName}_date`, parts.date);
    setFieldValue(form, `${fieldBaseName}_time`, parts.time);
  }

  function fillTimeSelectOptions() {
    const startField = ui.eventForm ? ui.eventForm.querySelector("[name='starts_at_time']") : null;
    const endField = ui.eventForm ? ui.eventForm.querySelector("[name='ends_at_time']") : null;
    fillTimeSelect("starts_at_time", startField ? startField.value : getDefaultStartTime());
    fillTimeSelect("ends_at_time", endField ? endField.value : "", true);
    applySteppedDateTimeInputConfig();
  }

  function fillTimeSelect(fieldName, selectedValue, allowEmpty) {
    const field = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}']`) : null;
    if (!field) {
      return;
    }
    const scheduleConfig = getScheduleConfig();
    const options = allowEmpty ? ["<option value=''>Keine Endzeit</option>"] : [];
    const startMinutes = timeStringToMinutes(scheduleConfig.event_time_start);
    const endMinutes = timeStringToMinutes(scheduleConfig.event_time_end);
    for (let minutes = startMinutes; minutes <= endMinutes; minutes += scheduleConfig.event_time_step_minutes) {
      const value = minutesToTimeString(minutes);
      options.push(`<option value="${value}">${value} Uhr</option>`);
    }
    field.innerHTML = options.join("");
    const fallbackValue = allowEmpty ? "" : getDefaultStartTime();
    field.value = selectedValue === "" ? "" : normalizeScheduleTime(selectedValue || field.value || fallbackValue, allowEmpty);
  }

  function normalizeScheduleTime(value, allowEmpty) {
    const scheduleConfig = getScheduleConfig();
    const raw = String(value || "").trim();
    if (!raw) {
      return allowEmpty ? "" : getDefaultStartTime();
    }
    const match = raw.match(/^(\d{2}):(\d{2})/);
    if (!match) {
      return allowEmpty ? "" : getDefaultStartTime();
    }
    const rawMinutes = (Number(match[1]) * 60) + Number(match[2]);
    const startMinutes = timeStringToMinutes(scheduleConfig.event_time_start);
    const endMinutes = timeStringToMinutes(scheduleConfig.event_time_end);
    const clamped = Math.max(startMinutes, Math.min(endMinutes, rawMinutes));
    const aligned = startMinutes + (Math.round((clamped - startMinutes) / scheduleConfig.event_time_step_minutes) * scheduleConfig.event_time_step_minutes);
    return minutesToTimeString(Math.max(startMinutes, Math.min(endMinutes, aligned)));
  }

  function normalizeSteppedTime(value) {
    const scheduleConfig = getScheduleConfig();
    const raw = String(value || "").trim();
    const match = raw.match(/^(\d{2}):(\d{2})/);
    if (!match) {
      return "";
    }
    const totalMinutes = (Number(match[1]) * 60) + Number(match[2]);
    const aligned = Math.round(totalMinutes / scheduleConfig.event_time_step_minutes) * scheduleConfig.event_time_step_minutes;
    const clamped = Math.max(0, Math.min((24 * 60) - scheduleConfig.event_time_step_minutes, aligned));
    return minutesToTimeString(clamped);
  }

  function currentSteppedLocalDateTimeValue() {
    const now = new Date();
    if (Number.isNaN(now.getTime())) {
      return "";
    }
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, "0");
    const day = String(now.getDate()).padStart(2, "0");
    const time = normalizeSteppedTime(`${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`);
    if (!time) {
      return "";
    }
    return `${year}-${month}-${day}T${time}`;
  }

  function readOptionalSteppedDateTime(formData, fieldName, label) {
    const raw = String(formData.get(fieldName) || "").trim();
    if (!raw) {
      return "";
    }
    const normalized = raw.includes("T")
      ? `${raw.split("T")[0]}T${normalizeSteppedTime(raw.split("T")[1] || "")}`
      : raw;
    const iso = toISO(normalized);
    if (!iso) {
      throw new Error(`${label} ist ungueltig.`);
    }
    return iso;
  }

  function applySteppedDateTimeInputConfig() {
    if (!ui.eventForm) {
      return;
    }
    const scheduleConfig = getScheduleConfig();
    const stepSeconds = scheduleConfig.event_time_step_minutes * 60;
    ["public_visible_from", "registration_opens_at", "registration_closes_at"].forEach((fieldName) => {
      const field = ui.eventForm.querySelector(`[name='${fieldName}']`);
      if (field) {
        field.step = String(stepSeconds);
      }
    });
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
      applyDateTimeFieldValidation(fieldName, label, !!allowEmpty);
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
    const scheduleConfig = getScheduleConfig();
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
    const totalMinutes = (hours * 60) + minutes;
    const minMinutes = timeStringToMinutes(scheduleConfig.event_time_start);
    const maxMinutes = timeStringToMinutes(scheduleConfig.event_time_end);
    if ((totalMinutes - minMinutes) % scheduleConfig.event_time_step_minutes !== 0) {
      return `${label} muss auf die konfigurierte Schrittweite von ${scheduleConfig.event_time_step_minutes} Minuten passen.`;
    }
    if (totalMinutes < minMinutes || totalMinutes > maxMinutes) {
      return `${label} muss zwischen ${scheduleConfig.event_time_start} und ${scheduleConfig.event_time_end} Uhr liegen.`;
    }
    return "";
  }

  function getAppSettings() {
    const source = state.tenantSettings && state.tenantSettings.app_settings
      ? state.tenantSettings.app_settings
      : {};
    return {
      event_time_start: String(source.event_time_start || "08:00").trim() || "08:00",
      event_time_end: String(source.event_time_end || "22:00").trim() || "22:00",
      event_time_step_minutes: Number(source.event_time_step_minutes || 15) || 15,
      event_slug_mode: String(source.event_slug_mode || "optional").trim() || "optional",
      allowed_embed_origins: Array.isArray(source.allowed_embed_origins) ? source.allowed_embed_origins : [],
      event_detail_base_url: String(source.event_detail_base_url || "").trim(),
      participant_cancel_deadline_hours: Number.isInteger(Number(source.participant_cancel_deadline_hours)) && Number(source.participant_cancel_deadline_hours) >= 0
        ? Number(source.participant_cancel_deadline_hours)
        : 24,
    };
  }

  function updateEventDetailBaseURLHint() {
    if (!ui.eventDetailBaseUrlHint) {
      return;
    }

    const exampleSlug = "beispiel-event";
    const publicBaseField = ui.settingsProfileForm ? ui.settingsProfileForm.querySelector("[name='public_base_url']") : null;
    const detailBaseField = ui.settingsRulesForm ? ui.settingsRulesForm.querySelector("[name='event_detail_base_url']") : null;
    const publicBaseURL = String(publicBaseField && publicBaseField.value ? publicBaseField.value : state.tenantProfile && state.tenantProfile.public_base_url || "").trim();
    const detailBaseURL = String(detailBaseField && detailBaseField.value ? detailBaseField.value : "").trim();

    const defaultURL = buildDetailPreviewURL(publicBaseURL, "", exampleSlug) || `.../events/${exampleSlug}`;
    const configuredURL = buildDetailPreviewURL(publicBaseURL, detailBaseURL, exampleSlug);
    const usesCustomBase = detailBaseURL !== "";
    const sameAsPublicBase = normalizePreviewURL(detailBaseURL) !== "" && normalizePreviewURL(detailBaseURL) === normalizePreviewURL(publicBaseURL);

    if (!usesCustomBase) {
      ui.eventDetailBaseUrlHint.innerHTML = `Leer lassen, wenn EEP die Standard-Detailseiten selbst ausliefern soll. Beispiel draussen: <code>${escapeHTML(defaultURL)}</code>`;
      return;
    }

    let message = `Aktuelle Detailseiten-Vorschau: <code>${escapeHTML(configuredURL || detailBaseURL)}</code>`;
    if (sameAsPublicBase) {
      message += ` Wenn dieselbe URL wie bei <code>public_base_url</code> gesetzt wird, entsteht bewusst kein automatisches <code>/events</code>. Fuer die bisherige EEP-Standardroute nutze besser <code>${escapeHTML(trimTrailingSlash(publicBaseURL) + "/events")}</code> oder lasse das Feld leer.`;
    } else {
      message += ` Erlaubte Platzhalter sind <code>{event_slug}</code>, <code>{slug}</code>, <code>{tenant_slug}</code> und <code>{series_slug}</code>.`;
    }
    ui.eventDetailBaseUrlHint.innerHTML = message;
  }

  function buildDetailPreviewURL(publicBaseURL, detailBaseURL, eventSlug) {
    const base = trimTrailingSlash(String(detailBaseURL || "").trim());
    const publicBase = trimTrailingSlash(String(publicBaseURL || "").trim());
    const slug = encodeURIComponent(String(eventSlug || "").trim());
    if (!slug) {
      return "";
    }
    if (!base) {
      return publicBase ? `${publicBase}/events/${slug}` : "";
    }
    let resolved = base;
    resolved = resolved.replaceAll("{event_slug}", slug);
    resolved = resolved.replaceAll("{slug}", slug);
    resolved = resolved.replaceAll("{tenant_slug}", "tenant-demo");
    resolved = resolved.replaceAll("{series_slug}", "serie-demo");
    if (resolved !== base) {
      return resolved;
    }
    return `${base}/${slug}`;
  }

  function normalizePreviewURL(value) {
    return trimTrailingSlash(String(value || "").trim().toLowerCase());
  }

  function trimTrailingSlash(value) {
    return String(value || "").replace(/\/+$/, "");
  }

  function getScheduleConfig() {
    const settings = getAppSettings();
    const step = Number(settings.event_time_step_minutes || 15);
    return {
      event_time_start: normalizeSimpleTime(settings.event_time_start, "08:00"),
      event_time_end: normalizeSimpleTime(settings.event_time_end, "22:00"),
      event_time_step_minutes: step > 0 ? step : 15,
      event_slug_mode: settings.event_slug_mode || "optional",
    };
  }

  function getDefaultStartTime() {
    return getScheduleConfig().event_time_start;
  }

  function normalizeSimpleTime(value, fallback) {
    const raw = String(value || "").trim();
    const match = raw.match(/^(\d{2}):(\d{2})$/);
    if (!match) {
      return fallback;
    }
    return `${match[1]}:${match[2]}`;
  }

  function timeStringToMinutes(value) {
    const match = String(value || "").trim().match(/^(\d{2}):(\d{2})$/);
    if (!match) {
      return 0;
    }
    return (Number(match[1]) * 60) + Number(match[2]);
  }

  function minutesToTimeString(totalMinutes) {
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
  }

  function applyDateTimeFieldValidation(fieldName, label, allowEmpty) {
    const dateField = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}_date']`) : null;
    const timeField = ui.eventForm ? ui.eventForm.querySelector(`[name='${fieldName}_time']`) : null;
    if (!dateField || !timeField) {
      return;
    }
    const formData = new FormData(ui.eventForm);
    const composedValue = composeLocalDateTimeValue(formData, fieldName);
    const message = getScheduleValidationMessage(composedValue, label, !!allowEmpty);
    dateField.setCustomValidity(message);
    timeField.setCustomValidity(message);
  }

  function validateEventScheduleFields() {
    applyDateTimeFieldValidation("starts_at", "Startzeit", false);
    applyDateTimeFieldValidation("ends_at", "Endzeit", true);
  }

  function applyEventSlugMode() {
    const slugField = ui.eventForm ? ui.eventForm.querySelector("[name='slug']") : null;
    if (!slugField) {
      return;
    }
    const mode = getScheduleConfig().event_slug_mode;
    slugField.required = mode === "required";
    if (mode === "auto") {
      slugField.placeholder = "wird automatisch erzeugt";
    } else if (mode === "required") {
      slugField.placeholder = "event-slug";
    } else {
      slugField.placeholder = "smoke-event";
    }
  }

  function parseOriginsTextarea(value) {
    return String(value || "")
      .split(/\r?\n|,/)
      .map((entry) => String(entry || "").trim())
      .filter(Boolean);
  }

  function validateSettingsSchedule(appSettings) {
    const step = Number(appSettings && appSettings.event_time_step_minutes || 15);
    if (!Number.isInteger(step) || step <= 0 || step > 60 || (60 % step) !== 0) {
      throw new Error("Die Schrittweite muss ein Teiler von 60 sein.");
    }
    const start = normalizeSimpleTime(appSettings && appSettings.event_time_start, "");
    const end = normalizeSimpleTime(appSettings && appSettings.event_time_end, "");
    if (!start || !end) {
      throw new Error("Bitte Start- und Endzeit vollstaendig setzen.");
    }
    if ((timeStringToMinutes(start) % step) !== 0 || (timeStringToMinutes(end) % step) !== 0) {
      throw new Error(`Start- und Endzeit muessen zur Schrittweite von ${step} Minuten passen.`);
    }
    if (timeStringToMinutes(end) <= timeStringToMinutes(start)) {
      throw new Error("Die Endzeit muss nach der Startzeit liegen.");
    }
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
    if (eventFilter.event) {
      chips.push(`Event ${eventFilter.event}`);
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
    if (displayOptions.load_css === false) {
      chips.push("Ohne EEP-CSS");
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
      ui.snippetPreviewHint.textContent = "Lade ein Snippet oder ein Event-Formular, um die Einbindung wie in Ghost direkt zu pruefen.";
      ui.snippetPreviewFrame.srcdoc = "";
      return;
    }
    const label = item && (item.name || item.title) ? (item.name || item.title) : "Embed";
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
      setFlash("Bitte zuerst ein Snippet oder Formular laden.", "error");
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
      setFlash("Bitte zuerst ein Snippet oder Formular laden.", "error");
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
