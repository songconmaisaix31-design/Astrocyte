// Astrocyte Paper service worker 只响应用户选择的 activeTab 提取。
// 快照保存在 popup 内存并供用户复制；应用侧负责人工审阅和确认导入。
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (!message || message.type !== "astrocyte-paper-snapshot") {
    return false;
  }
  // 只接受当前 extension 的发送者身份。
  if (!sender.id || sender.id !== chrome.runtime.id) {
    sendResponse({ ok: false, reason: "untrusted_sender" });
    return false;
  }
  sendResponse({ ok: true });
  return false;
});
