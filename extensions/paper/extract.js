// Astrocyte Paper extraction core (browser build).
//
// Adapted from steipete/summarize's DOM clone + Readability approach
// (MIT, Peter Steinberger) and the Astrocyte scholarly metadata parser used by
// the Node build. This file runs in the extension's isolated world, reads only
// the current page the human clicked, performs no navigation or network call,
// and never persists credentials or cookies. Full text requires scholarly body
// evidence; metadata/restricted states stay explicit.
(function (global) {
  "use strict";

  function normalizeDOI(raw) {
    if (!raw) return "";
    const value = String(raw).trim()
      .replace(/^doi:\s*/i, "")
      .replace(/^https?:\/\/(?:dx\.)?doi\.org\//i, "");
    return /^10\.\d{4,9}\/[^\s<>"?#]+$/i.test(value) ? value.toLowerCase() : "";
  }

  function paperHostFamily(host) {
    host = (host || "").toLowerCase();
    const families = {
      arxiv: ["arxiv.org"], pmc: ["pmc.ncbi.nlm.nih.gov"],
      preprints: ["biorxiv.org", "medrxiv.org"], plos: ["journals.plos.org"],
      nature: ["nature.com"], springer: ["link.springer.com"],
      elsevier: ["sciencedirect.com", "cell.com"], ieee: ["ieeexplore.ieee.org"],
      acm: ["dl.acm.org"], wiley: ["onlinelibrary.wiley.com"], science: ["science.org"],
      mdpi: ["mdpi.com"], frontiers: ["frontiersin.org"], openreview: ["openreview.net"],
      acl: ["aclanthology.org"], cvf: ["openaccess.thecvf.com"], pmlr: ["proceedings.mlr.press"],
      doi: ["doi.org"]
    };
    return Object.entries(families).find(([, domains]) =>
      domains.some(d => host === d || host.endsWith("." + d))
    )?.[0] ?? "generic";
  }

  function extractPaper(locator) {
    const url = new URL(locator);
    if (url.protocol !== "https:" || url.username || url.password || url.port) {
      throw new Error("Public HTTPS paper URL required");
    }
    const document = global.document;
    const Readability = global.Readability;
    if (!Readability || typeof Readability !== "function") {
      throw new Error("Readability not loaded");
    }
    const metas = new Map();
    for (const node of document.querySelectorAll("meta[content]")) {
      const key = (node.getAttribute("name") || node.getAttribute("property") || "").toLowerCase();
      const value = (node.getAttribute("content") || "").trim();
      if (key && value) metas.set(key, [...(metas.get(key) || []), value]);
    }
    const first = (...keys) => keys.map(k => metas.get(k)?.[0]).find(Boolean) || "";
    let scholarly = {};
    for (const script of document.querySelectorAll('script[type="application/ld+json"]')) {
      try {
        const root = JSON.parse(script.textContent);
        const entries = Array.isArray(root) ? root : [...(root["@graph"] || []), root];
        scholarly = entries.find(v => ["ScholarlyArticle", "MedicalScholarlyArticle"].includes(v["@type"])) || scholarly;
      } catch { /* malformed metadata does not invent identity */ }
    }
    const doi = normalizeDOI(first("citation_doi", "dc.identifier", "dc.identifier.doi", "prism.doi")) ||
      normalizeDOI(typeof scholarly.identifier === "string" ? scholarly.identifier : scholarly.identifier?.value || "");
    const arxivMatch = url.hostname.match(/(^|\.)arxiv\.org$/) && url.pathname.match(/^\/(?:abs|html|pdf)\/((?:\d{4}\.\d{4,5}|[a-z][a-z0-9.-]*(?:\.[A-Z]{2})?\/\d{7})(?:v[1-9]\d*)?)(?:\.pdf)?\/?$/);
    const arxiv = arxivMatch?.[1] || "";
    const title = first("citation_title", "dc.title", "og:title") || scholarly.headline || document.title || "";
    const abstractNode = document.querySelector("#abstract, .abstract, .acl-abstract, #Abs1, section.abstract, .c-article-section__content[id^=\"Abs\"]");
    const abstractCopy = abstractNode?.cloneNode(true);
    for (const heading of abstractCopy?.querySelectorAll("h1,h2,h3,h4") || []) heading.remove();
    const abstract = first("citation_abstract") || abstractCopy?.textContent?.trim() || scholarly.abstract || first("dc.description", "description", "og:description") || "";
    const authors = metas.get("citation_author") || metas.get("dc.creator") || [];
    const pdfs = [];
    const addPDF = raw => {
      try {
        const parsed = new URL(raw, locator);
        if (parsed.protocol === "https:" && !parsed.username && !parsed.password && !parsed.port && !pdfs.includes(parsed.href)) {
          pdfs.push(parsed.href);
        }
      } catch { /* bad metadata remains absent */ }
    };
    for (const raw of metas.get("citation_pdf_url") || []) addPDF(raw);
    for (const node of document.querySelectorAll('a[href$=".pdf"], link[type="application/pdf"]')) addPDF(node.getAttribute("href"));
    if (arxiv) addPDF("https://arxiv.org/pdf/" + arxiv);
    const sourceKey = arxiv ? "arxiv:" + arxiv.replace(/v[1-9]\d*$/, "") : doi ? "doi:" + doi : locator;
    const clone = document.cloneNode(true);
    for (const node of clone.querySelectorAll("script, style, nav, footer, form, aside, [hidden], [aria-hidden=\"true\"], .related-articles, #related-articles, [aria-label=\"Related articles\"]")) node.remove();
    const visible = [...clone.querySelectorAll("article,main,section,p,div,h1,h2,h3,h4")].map(n => n.textContent).join("\n");
    const parsed = new Readability(clone, { keepClasses: false }).parse();
    const readable = parsed?.textContent?.trim() || "";
    const retained = document.createElement("div");
    retained.innerHTML = parsed?.content || "";
    const sections = [...retained.querySelectorAll("h1,h2,h3,h4")]
      .filter(n => !n.closest('[id*="abstract"], [id^="Abs"], .abstract, .acl-abstract'))
      .map(n => n.textContent.trim());
    const hasIntro = sections.some(s => /^(?:\d+[.\s]*)?introduction\b/i.test(s));
    const hasBody = sections.some(s => /\b(methods?|results?|experiments?|discussion|conclusions?)\b/i.test(s));
    const fullBody = hasIntro && hasBody && readable.length > 1000;
    const paywall = !fullBody && /\b(?:purchase this article|subscribe to access)\b/i.test(visible || readable);
    const restricted = !fullBody && !paywall && /\b(?:access through your institution|sign in to access|checking your browser|verify you are human)\b/i.test(visible || readable);
    const state = fullBody ? "readable_fulltext" : paywall ? "paywall" : restricted ? "restricted" : "abstract_only";
    const text = state === "readable_fulltext" ? readable : "";
    const warning = state === "readable_fulltext" ? "" : (paywall || restricted) ? "Access restriction detected; no entitlement bypass" : "Readable full-paper structure not established; title/abstract remain metadata";
    return {
      schema_version: 1, source_url: locator, host_family: paperHostFamily(url.hostname), source_key: sourceKey,
      title: String(title).trim(), abstract: String(abstract).trim(), authors, doi, arxiv_id: arxiv,
      observed_version: arxiv.match(/v[1-9]\d*$/)?.[0] || first("citation_publication_date", "dc.date") || "",
      pdf_urls: pdfs, content_state: state, text, warning, truncated: false,
      provenance: { processor: "summarize-readability-approach", version: "browser " + (readable ? "readability" : "none"), mode: state, source: locator }
    };
  }

  global.astrocytePaperExtract = extractPaper;
})(typeof self !== "undefined" ? self : this);
