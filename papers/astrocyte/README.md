# Astrocyte method and protocol draft 0.4

The complete [PDF](preprint/manuscript.pdf) and its complete [LaTeX source](preprint/manuscript.tex) are the versioned review deliverables. [manuscript.md](manuscript.md) remains the single editable manuscript source; [references.bib](references.bib) contains the fourteen cited records, including four original statistical articles and DecentMem. Pandoc regenerates the complete LaTeX, including equations, tables, and inline resolved references, from those two inputs. There is no separately maintained abbreviated manuscript.

[sources.md](sources.md) records versions, reading limits, and metadata discrepancies. [claims-evidence.md](claims-evidence.md) distinguishes proposed methods, inspected implementation, engineering checks, and uncollected results. The [Chinese research feedback](../../docs/research/astrocyte-research-feedback.md) records research requirements for coordinator decisions.

Authors, affiliations, and ORCID identifiers remain unfilled. Formal models, tasks, budgets, thresholds, replication counts, and registration remain undecided. The PDF contains the current method/protocol draft, without formal experimental results. This package has not been submitted to arXiv or a registration platform.

The proposed release rule fixes actual produced artifacts and source-supported dispositions for planned positions. An observed algorithmic failure may have no artifact and receive its predeclared failure outcome. Unknown evaluation, damaged data and incomplete blocks retain the existing analysis-stop rules; missing measurements are not converted into zeros.

Pilot settings and the intended variance target are versioned before pilot outcomes. A materially revised formal target needs an explicit applicability argument for the earlier variance; otherwise independently preplanned pilot blocks for the revision precede sample-size decisions. Pilot blocks and their costs remain development records, and no formal outcomes are used to revise the plan.

The accepted round-06 semantics and five conditional proofs remain unchanged: versioned pre-decision conditions, obligation-specific verdicts, actual authorized record views, delayed visibility, scheduled-position decay, initialization, disabled-channel projections and their stated bounds. Round 08 adds an explicit four-role access matrix, evaluator blinding and a proposed seal through all planned main-study blocks, leakage-route checks, source-linked cost accounting without inclusive-total duplication, a proposed one-time preparation allocation over the planned eight-condition sequences, and the readiness order from lawful sources and implementation through independent pilot blocks to formal parameter decisions. These are complete proposals with no claim of actual access isolation, audited experimental costs, measured benefit, or approved interfaces. The round-04 BCa/inversion/Holm/TOST procedure is preserved; its planning subsection now specifies those prerequisites.

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

## TeX Live technical preparation

`build-texlive.ps1` is a PowerShell 7.2+ entry for the separately authorized Windows portable TeX Live 2025 final check. Download or explicitly authorized reuse through VerifyInstaller, followed by Install, Build and Validate, each require the coordinator's assigned window and run serially under the same resource guard. The full `scheme-full` distribution retains documentation and sources. An absolute-path profile must first be written in the worktree's ignored `.verification/texlive2025/` directory; portable mode and explicit profile settings disable host integration. Downloads, public keyring, installation, caches, clean extraction and new PDF/logs stay there.

The entry checks the official installer checksum and uses a dearmored public keyring plus GnuPG's native signer-fingerprint assertion. It then installs from the fixed 2025 final repository, packages the unchanged full review TeX, extracts that ZIP into a fresh directory, and invokes XeLaTeX twice with shell escape disabled and immediate error termination. After both passes and diagnostic checks succeed, Build stages `preprint/arxiv-source.zip`; Validate then checks the actual PDF and complete documents together. The explicit PDF option retains all original manuscript/reference/PDF assertions and saves extracted text beside the selected PDF. A staged ZIP alone does not establish validation, and the original TeX/PDF remains intact. Invoke stages individually only as authorized:

```powershell
& .\papers\astrocyte\build-texlive.ps1 -Phase Download
& .\papers\astrocyte\build-texlive.ps1 -Phase Install
& .\papers\astrocyte\build-texlive.ps1 -Phase Build
& .\papers\astrocyte\build-texlive.ps1 -Phase Validate
```

Completed original downloads are reused with `-Phase VerifyInstaller` only under explicit coordinator authorization. It preserves the earlier failure and uses new public-keyring and verification logs before any extraction.

Install invokes the verified distribution's bundled Perl and original `install-tl` directly, with the supported `-lang en` option. It clears the same inherited Perl options as the official Windows wrapper, provides the bundled library and executable directory in the child environment, and checks the real Perl exit code. The selected GPG executable is supplied by its filename and explicit process PATH directory, as required by this release's Windows discovery routine; its keyring and signature checks remain enabled. Installer and console logs use the resource sample timestamp and preserve earlier diagnostics.

The latest actual invocation verified the official 2025 final repository and completed the six-package initial stage. The coordinator then interrupted the full installation during its subsequent 4973-package stage to restore the user application. All owned installation processes exited. The partial installation and logs remain in the ignored directory; Build, Validate and source-ZIP generation have not run. Install rejects an existing target, so this state requires an explicitly reviewed recovery plan before another invocation. The original review PDF and TeX remain unchanged.

The 2025 final repository differs from arXiv's 2025-08-03 snapshot. A local check cannot establish the site's exact processing or submission readiness. Accurate commands, failure diagnostics, interruption and resource samples are in [validation.md](validation.md). Full TeX Live installation and compilation remain incomplete.

## Submission status

The local PDF is compiled with Tectonic. The complete generated source has not been compiled in arXiv's supported TeX Live environment or processed by arXiv. A source upload package must contain only files needed to compile, excluding the review PDF and build logs, under the current [TeX submission requirements](https://info.arxiv.org/help/submit_tex.html). The `preprint/` directory provides review deliverables and is not an approved arXiv upload package.

Substantive contribution, method approval and controlled validation, actual research evidence, author metadata, paper license, and author consent require further work or decisions under arXiv's [moderation policy](https://info.arxiv.org/help/moderation/index.html) and [submission agreement](https://info.arxiv.org/help/policies/submission_agreement.html). Actual Codex-assisted writing is disclosed in the manuscript; human scientific review and responsibility remain to be confirmed. A successful local build does not establish submission readiness. Independent reviews accepted the limited scopes of rounds 02, 04, 06 and 08. The present technical entry fixes await independent review; complete TeX Live compilation remains outstanding, and the same research Worker handles revisions.
