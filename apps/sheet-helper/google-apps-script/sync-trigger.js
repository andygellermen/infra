/**
 * Google Apps Script trigger helper for the Sheet Helper app.
 *
 * Recommended usage:
 * 1. Open the target Google Sheet.
 * 2. Open Extensions -> Apps Script.
 * 3. Paste this file into the project.
 * 4. Set the script properties:
 *    - SHEET_HELPER_SYNC_URL
 *    - SHEET_HELPER_SYNC_TOKEN
 * 5. Run installRecommendedTriggers() once, or manually create an
 *    installable "On change" trigger for sheetHelperOnChange().
 *
 * This script is intentionally simple and aimed at unsensitive content only.
 */

function sheetHelperOnEdit(e) {
  if (!e || !e.range || typeof e.range.getSheet !== 'function') {
    return;
  }

  var sheet = e.range.getSheet();
  var sheetName = sheet.getName();

  if (shouldIgnoreSheet_(sheetName)) {
    return;
  }

  triggerSheetHelperSync_('onEdit', {
    sheetName: sheetName,
    a1Notation: e.range.getA1Notation()
  });
}

function sheetHelperOnChange(e) {
  var spreadsheet = e && e.source ? e.source : SpreadsheetApp.getActiveSpreadsheet();
  var activeSheet = spreadsheet && typeof spreadsheet.getActiveSheet === 'function'
    ? spreadsheet.getActiveSheet()
    : null;

  var details = {
    changeType: e && e.changeType ? e.changeType : 'UNKNOWN'
  };

  if (activeSheet && typeof activeSheet.getName === 'function') {
    var sheetName = activeSheet.getName();
    if (shouldIgnoreSheet_(sheetName)) {
      return;
    }
    details.sheetName = sheetName;
  }

  triggerSheetHelperSync_('onChange', details);
}

function manualSync() {
  triggerSheetHelperSync_('manual', {});
}

function installRecommendedTriggers() {
  removeSheetHelperTriggers();

  ScriptApp.newTrigger('sheetHelperOnChange')
    .forSpreadsheet(SpreadsheetApp.getActiveSpreadsheet())
    .onChange()
    .create();
}

function installEditOnlyTrigger() {
  removeSheetHelperTriggers();

  ScriptApp.newTrigger('sheetHelperOnEdit')
    .forSpreadsheet(SpreadsheetApp.getActiveSpreadsheet())
    .onEdit()
    .create();
}

function installEditAndChangeTriggers() {
  removeSheetHelperTriggers();

  ScriptApp.newTrigger('sheetHelperOnEdit')
    .forSpreadsheet(SpreadsheetApp.getActiveSpreadsheet())
    .onEdit()
    .create();

  ScriptApp.newTrigger('sheetHelperOnChange')
    .forSpreadsheet(SpreadsheetApp.getActiveSpreadsheet())
    .onChange()
    .create();
}

function removeSheetHelperTriggers() {
  var triggers = ScriptApp.getProjectTriggers();

  for (var i = 0; i < triggers.length; i++) {
    var trigger = triggers[i];
    var handler = trigger.getHandlerFunction();

    if (!isSheetHelperTriggerHandler_(handler)) {
      continue;
    }

    ScriptApp.deleteTrigger(trigger);
  }
}

function triggerSheetHelperSync_(reason, details) {
  var props = PropertiesService.getScriptProperties();
  var baseUrl = props.getProperty('SHEET_HELPER_SYNC_URL');
  var token = props.getProperty('SHEET_HELPER_SYNC_TOKEN');

  if (!baseUrl) {
    throw new Error('Missing script property: SHEET_HELPER_SYNC_URL');
  }
  if (!token) {
    throw new Error('Missing script property: SHEET_HELPER_SYNC_TOKEN');
  }

  var payload = {
    reason: reason,
    spreadsheetId: SpreadsheetApp.getActiveSpreadsheet().getId(),
    spreadsheetName: SpreadsheetApp.getActiveSpreadsheet().getName(),
    editedAt: new Date().toISOString(),
    details: details || {}
  };

  var url = buildSyncUrl_(baseUrl, token);
  var response = UrlFetchApp.fetch(url, {
    method: 'post',
    contentType: 'application/json',
    muteHttpExceptions: true,
    payload: JSON.stringify(payload)
  });

  var code = response.getResponseCode();
  if (code < 200 || code >= 300) {
    throw new Error('Sync failed with HTTP ' + code + ': ' + response.getContentText());
  }
}

function buildSyncUrl_(baseUrl, token) {
  var normalized = baseUrl.replace(/\/+$/, '');
  return normalized + '/' + encodeURIComponent(token);
}

function shouldIgnoreSheet_(sheetName) {
  return sheetName === '_meta' || sheetName === '_stats';
}

function isSheetHelperTriggerHandler_(handler) {
  return handler === 'sheetHelperOnEdit' ||
    handler === 'sheetHelperOnChange' ||
    handler === 'onEdit' ||
    handler === 'onChange';
}
