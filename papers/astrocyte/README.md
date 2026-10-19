# Astrocyte method and protocol draft 0.2

The complete [PDF](preprint/manuscript.pdf) and its complete [LaTeX source](preprint/manuscript.tex) are the versioned review deliverables. [manuscript.md](manuscript.md) remains the single editable manuscript source; [references.bib](references.bib) contains the fourteen cited records, including four original statistical articles and DecentMem. Pandoc regenerates the complete LaTeX, including equations, tables, and inline resolved references, from those two inputs. There is no separately maintained abbreviated manuscript.

[sources.md](sources.md) records versions, reading limits, and metadata discrepancies. [claims-evidence.md](claims-evidence.md) distinguishes proposed methods, inspected implementation, engineering checks, and uncollected results. The [Chinese research feedback](../../docs/research/astrocyte-research-feedback.md) records research requirements for coordinator decisions.

Authors, affiliations, and ORCID identifiers remain unfilled. Formal models, tasks, budgets, thresholds, replication counts, and registration remain undecided. The PDF contains the current method/protocol draft, without formal experimental results. This package has not been submitted to arXiv or a registration platform.

## Build

The tested tools are Pandoc 3.9 and the official portable Windows Tectonic 0.17.0. With both executables available:

```powershell
& .\papers\astrocyte\build.ps1 -Pandoc 'pandoc' -Tectonic 'tectonic'
```

This worktree used executables in its ignored verification directory:

```powershell
& .\papers\astrocyte\build.ps1 `
  -Pandoc (Resolve-Path .\.verification\python\pypandoc\files\pandoc.exe).Path `
  -Tectonic (Resolve-Path .\.verification\tectonic\tectonic.exe).Path
```

The script generates HTML with MathML, a Pandoc JSON AST, and complete LaTeX, then invokes `tectonic -X compile --untrusted --keep-logs --keep-intermediates`. Pandoc citation/conversion warnings and unsuccessful process exits terminate the build. TeX diagnostics remain in `build/manuscript.log` and `build/tectonic-console.log`; successful PDF generation still requires reviewing those diagnostics. The PDF and complete LaTeX are copied to `preprint/` for versioned delivery. Intermediate outputs, downloaded tools, and caches remain ignored.

[fontconfig.conf](fontconfig.conf) supplies a project configuration for the portable Windows engine. Text and mathematics use TeX-provided Latin Modern font files, without configured system font directories. Process-level `FONTCONFIG_FILE`, `TECTONIC_CACHE_DIR`, `XDG_CACHE_HOME`, `TEMP`, and `TMP` direct configuration, caches, and temporary files to this package. No global configuration or PATH change is required. The official [Tectonic installation guide](https://tectonic-typesetting.github.io/book/latest/getting-started/install.html) and [Fontconfig configuration reference](https://www.freedesktop.org/software/fontconfig/fontconfig-user.html) describe the respective facilities. The first Tectonic build downloads the required TeX support files and needs network access.

Use only the coordinator's assigned heavy-operation window. The script runs at BelowNormal priority with Node's heap setting at 1024 MiB and performs conversions serially. It samples whole-machine CPU and available memory twice, checks the measured interval is 2–5 seconds, and stops if either CPU sample is at least 85% or available memory is below 4 GiB. These are startup checks under the coordinator's updated allowance of 19 October 2026; they do not assert a peak-utilization bound.

## Verification

`verify.py` uses Pandoc's JSON AST, Pybtex, BeautifulSoup, Markdown-it, and PyMuPDF. It checks equality of the resolved citation keys and bibliography entries, the eight configuration rows, generated references/MathML, local links, privacy patterns in parsed prose/code/link targets and rendered PDF text, and the delivered PDF. Reference counts follow the actual bibliography. PDF checks cover readable pages without repair, all manuscript section headings, all reference titles, configuration identifiers, metadata, embedded fonts without Type 3, and text spans inside page boundaries. The entire extracted PDF text is saved to `build/manuscript-pdf.txt` for reading alongside the source and TeX diagnostics. Text extraction and bounding boxes do not establish visual quality.

Python dependencies are `pybtex`, `beautifulsoup4`, `markdown-it-py`, and `pymupdf`. The current worktree used existing BeautifulSoup/Markdown-it/PyMuPDF and a local Pybtex installation. Run verification under the same resource and window conditions:

```powershell
$env:PYTHONPATH = (Resolve-Path .\.verification\python).Path
python -X utf8 .\papers\astrocyte\verify.py
git diff --check
```

The actual commands, output, retained diagnostic, and limits are in [validation.md](validation.md). No product services, model experiments, sub-agents, or visual checks were run in this round.

## Submission status

The local PDF is compiled with Tectonic. The complete generated source has not been compiled in arXiv's supported TeX Live environment or processed by arXiv. A source upload package must contain only files needed to compile, excluding the review PDF and build logs, under the current [TeX submission requirements](https://info.arxiv.org/help/submit_tex.html). The `preprint/` directory provides review deliverables and is not an approved arXiv upload package.

Substantive contribution, complete method definitions, actual research evidence, author metadata, paper license, author consent, and disclosure of significant generative-AI use require further work or decisions under arXiv's [moderation policy](https://info.arxiv.org/help/moderation/index.html) and [submission agreement](https://info.arxiv.org/help/policies/submission_agreement.html). A successful local build does not establish those conditions. Independent review of this round remains pending; the same research Worker handles revisions.
