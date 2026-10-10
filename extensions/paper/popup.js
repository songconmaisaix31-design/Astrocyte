// Astrocyte Paper popup. On an explicit human click it injects the vendored
// Readability and the extraction core into the current tab only (activeTab),
// then shows title/abstract/identity and the observed content state. Import is
// a separate explicit action; a snapshot is never auto-imported.
"use strict";

const result = document.getElementById("result");
const hint = document.getElementById("hint");
const extractBtn = document.getElementById("extract");

function stateLabel(state) {
  return {
    readable_fulltext: "可读全文",
    abstract_only: "仅摘要/元数据",
    paywall: "付费墙",
    restricted: "访问受限"
  }[state] || state;
}

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

async function getActiveTab() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  return tab;
}

async function extractCurrentPage() {
  const tab = await getActiveTab();
  if (!tab || !tab.id) throw new Error("没有找到当前标签页");
  const url = new URL(tab.url);
  if (url.protocol !== "https:") {
    throw new Error("仅支持公开 HTTPS 页面");
  }
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    files: ["Readability.js", "extract.js"]
  });
  const [injection] = await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    func: locator => globalThis.astrocytePaperExtract(locator),
    args: [tab.url]
  });
  return injection.result;
}

function render(snapshot) {
  const authors = (snapshot.authors || []).join(", ");
  const pdfs = (snapshot.pdf_urls || [])
    .map(u => `<div class="meta">PDF 候选：<a href="${esc(u)}" target="_blank" rel="noreferrer">${esc(u)}</a></div>`)
    .join("");
  result.innerHTML = `
    <div class="card">
      <span class="state ${esc(snapshot.content_state)}">${esc(stateLabel(snapshot.content_state))}</span>
      <div class="title">${esc(snapshot.title || "（无标题）")}</div>
      ${authors ? `<div class="meta">作者：${esc(authors)}</div>` : ""}
      ${snapshot.doi ? `<div class="meta">DOI：${esc(snapshot.doi)}</div>` : ""}
      ${snapshot.arxiv_id ? `<div class="meta">arXiv：${esc(snapshot.arxiv_id)}</div>` : ""}
      ${snapshot.abstract ? `<div class="label">摘要</div><div class="abstract">${esc(snapshot.abstract)}</div>` : ""}
      ${pdfs}
      ${snapshot.warning ? `<div class="warn">${esc(snapshot.warning)}</div>` : ""}
    </div>
    <button id="copy" class="secondary" type="button">复制快照（供 Astrocyte 应用人工审阅）</button>
  `;
  const copy = document.getElementById("copy");
  copy.addEventListener("click", async () => {
    await copySnapshot(snapshot);
  });
}

async function copySnapshot(snapshot) {
  try {
    await navigator.clipboard.writeText(JSON.stringify(snapshot, null, 2));
    hint.textContent = "已复制快照到剪贴板；请在 Astrocyte 应用中粘贴并人工确认入库。";
  } catch {
    hint.textContent = "复制失败；请手动记录标题与来源。";
  }
}

extractBtn.addEventListener("click", async () => {
  result.innerHTML = "";
  hint.textContent = "";
  extractBtn.disabled = true;
  extractBtn.textContent = "提取中…";
  try {
    const snapshot = await extractCurrentPage();
    render(snapshot);
  } catch (err) {
    result.innerHTML = `<div class="error">${esc(err.message || err)}</div>`;
  } finally {
    extractBtn.disabled = false;
    extractBtn.textContent = "读取当前页并提取";
  }
});
