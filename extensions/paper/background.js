// Astrocyte Paper service worker.
//
// The extension performs only the human-invoked, activeTab-scoped extraction.
// It never enumerates tabs, reads cookies or storage, stores provider
// credentials, or imports anything without an explicit human selection.
//
// The application-side same-origin review/import transport is owned by W0 and
// is intentionally not wired here; the extension only exposes an in-memory
// snapshot for the popup and clipboard, so a plugin token can never substitute
// for human approval.
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (!message || message.type !== "astrocyte-paper-snapshot") {
    return false;
  }
  // The sender must be the extension popup itself; no external sender is
  // trusted. A future W0-owned app handoff will add an explicit
  // externally_connectable origin check and human review here.
  if (!sender.id || sender.id !== chrome.runtime.id) {
    sendResponse({ ok: false, reason: "untrusted_sender" });
    return false;
  }
  sendResponse({ ok: true });
  return false;
});
