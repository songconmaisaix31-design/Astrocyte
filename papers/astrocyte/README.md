# Astrocyte research draft 0.1

[manuscript.md](manuscript.md) is the complete English draft and the single editable manuscript source. [references.bib](references.bib) contains the nine cited records: six papers, two official communications/documentation sources, and one pinned engineering report. [sources.md](sources.md) records full-text access, versions, reading limits, and metadata discrepancies. [claims-evidence.md](claims-evidence.md) separates proposed methods, inspected implementation, reported engineering checks, and uncollected formal results. The [Chinese research feedback](../../docs/research/astrocyte-research-feedback.md) gives implementation gaps and acceptance methods for coordinator decisions.

Authors, affiliations, and ORCID identifiers are unfilled. Eight-configuration experiments, formal budgets, thresholds, sample sizes, and registration remain unfinished. This package was not submitted to arXiv or a registration platform.

## Build

The build uses mature Pandoc conversion and citeproc. On Windows, with Pandoc available:

```powershell
& .\papers\astrocyte\build.ps1 -Pandoc 'pandoc'
```

The script samples whole-machine CPU and available memory twice, three seconds apart, and refuses to start if either CPU sample is at least 70% or memory is below 4 GiB. It runs at BelowNormal priority, keeps Node's heap setting at 512 MiB, and performs conversions serially. It produces `build/manuscript.html`, `build/manuscript.tex`, and `build/manuscript.json`, plus startup resource samples. The HTML embeds MathML and needs no downloaded script for its formulas. Reading that file requires a MathML-capable viewer; no browser or visual inspection was performed here.

In this worktree Pandoc was installed only in the ignored verification directory with `pypandoc_binary`, and the actual build command was:

```powershell
& .\papers\astrocyte\build.ps1 -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path
```

It completed with Pandoc 3.9, including `--fail-if-warnings` HTML and LaTeX source generation. The generated LaTeX contains the full manuscript, equations, tables, and resolved references. `manuscript.md` and `references.bib` regenerate that source; generated files and downloaded papers are ignored. There is no separately maintained abbreviated LaTeX manuscript.

## Verification

`verify.py` uses Pandoc's JSON AST, Pybtex, BeautifulSoup, and Markdown-it. It checks actual citation resolution, the eight factorial rows, generated references/MathML, local document links, document metadata, and private-path absence in the manuscript. It checks document integrity and does not estimate scientific effects or validate product execution.

Dependencies are `pybtex`, `beautifulsoup4`, and `markdown-it-py`. This worktree used the existing BeautifulSoup/Markdown-it and a local Pybtex installation:

```powershell
$env:PYTHONPATH = (Resolve-Path .\.verification\python).Path
python -X utf8 .\papers\astrocyte\verify.py
git diff --check
```

Use the same resource guard before dependency installation and verification. The exact local commands, results, and remaining limits are in [validation.md](validation.md). There were no model experiments, native product requests, product builds/services, GPU jobs, sub-agents, or visual checks during this writing task.

## PDF and submission limits

No TeX engine (`latexmk`, `pdflatex`, `xelatex`, `lualatex`, or `tectonic`) was found on PATH. PDF compilation was not performed. Generated LaTeX is available locally, but successful conversion does not prove TeX compilation, pagination, or arXiv compatibility. A TeX-equipped reviewer can compile the generated source; that compilation and any typography/template corrections remain required before submission. This package does not claim a built or visually checked PDF.

Formal research execution and public submission remain separate, unperformed operations. The independent Reviewer will review the exact candidate Commit; this Worker remains responsible for revisions on the same task.
