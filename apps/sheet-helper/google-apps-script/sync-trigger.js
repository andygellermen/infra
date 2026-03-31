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
 * 5. Create an installable trigger for onEdit or call manualSync() manually.
 *
 * This script is intentionally simple and aimed at unsensitive content only.
 */

function onEdit(e) {
  if (!e) {
    return;
  }

  var sheet = e.range.getSheet();
  var sheetName = sheet.getName();

  // Keep noisy edits away from hidden/admin sheets if needed.
  if (sheetName === '_meta' || sheetName === '_stats') {
    return;
  }

  triggerSheetHelperSync_('onEdit', {
    sheetName: sheetName,
    a1Notation: e.range.getA1Notation()
  });
}

function manualSync() {
  triggerSheetHelperSync_('manual', {});
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
