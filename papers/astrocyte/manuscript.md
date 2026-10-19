---
title: 'Astrocyte: Experience Reuse and Collective Adaptation in Decentralized Agent Swarms'
author: 'Authors: TO BE PROVIDED'
date: '19 October 2026 — method and protocol draft 0.3'
lang: en
abstract: |
  Reusable experience can influence how agents solve tasks and choose collaborators, while collaboration generates evidence that can revise both experience and future relationships. We propose Astrocyte as a framework for studying this coupled adaptation with fixed foundation-model parameters and constrained execution budgets. The proposed method combines versioned task and collaboration experience, condition-specific directed relationship scores, evidence-gated updates, and a shared decision rule that can be executed centrally or by individual members. A three-factor, eight-configuration protocol separates experience updates, relationship updates, and decision location. Its primary estimand is the interaction between the two updates under local decisions; secondary estimands measure how this interaction and the total update gain vary with decision location. Complete task sequences form paired replication units, and execution feedback is separated from sealed evaluation. We specify cost accounting, member perturbations, and a prospective system-iteration case. This manuscript presents a method and an unregistered experimental protocol. Current implementation evidence covers selected product components; the research policy, complete experimental controls, and formal measurements remain incomplete. No formal experimental results are reported, and no positive adaptation or decentralization effect is assumed.
---

**Affiliations: TO BE PROVIDED. ORCID identifiers: TO BE PROVIDED.**

**Status.** This is a complete first manuscript for review, based on the project positioning document v3 and experimental preregistration draft v0.1, both dated 19 October 2026. The protocol has not been formally registered. Model versions, tasks, budgets, sample sizes, and several analysis details require confirmation before registration. Nothing in this draft constitutes an empirical performance claim for the proposed method.

# Introduction

An agent can preserve a useful procedure after completing a task. Reusing that procedure may change which task it attempts, which collaborator it requests, or which check it performs. Those choices change the evidence available for subsequent learning. A collaboration network can similarly favor partners whose recent contributions have worked under relevant conditions. These processes form a coupled feedback system: experience affects collaboration, and collaboration affects the production and revision of experience.

The scientific question is whether updating both processes yields a practically useful interaction over continuous tasks. A system might benefit from each update independently and still show little additional value from their combination. It might also exhibit a negative interaction when an erroneous experience repeatedly directs work toward an unsuitable partner. Measuring only the full system against a static baseline cannot distinguish these possibilities.

We therefore propose a factorial study of experience updates, relationship updates, and the location of collaboration decisions. Foundation-model parameters remain fixed within the main study. The external experience state and condition-specific relationship records may change according to the assigned configuration. The word *decentralized* refers to members submitting their own task and partner decisions. Authorization, task claims, durable records, and budget enforcement may use shared services. Each Mission retains one authoritative execution backend.

The proposed research focuses on connecting exact experience versions to checkable collaboration evidence, defining condition-specific relationship updates, and separating update interaction from deployment-dependent decision-location effects. This manuscript provides a method specification, a proposed analysis procedure, and an implementation audit. Their substantive originality still requires assessment against closely related methods, and the incomplete research implementation requires controlled validation. Factorial contrast algebra and general statistical procedures are established tools; they are not presented as independent theoretical breakthroughs. Benefits remain unmeasured.

The main study is distinct from the proposed system-iteration case. In the main study, the code and decision rule remain fixed across conditions. In the case, selected code or configuration changes are themselves prospective research artifacts. Mixing these settings would confound adaptation of external state with changes to the underlying agent system.

# Related work and the scope of the research question

Strategy Genes studies compact, structured experience representations and selective incorporation of failure information with fixed model parameters [@wang2026genes, Sections 3–4 and Appendix A]. This motivates controlling experience representation and its cost. Astrocyte's proposed factorial question concerns the joint effect of updating that experience and persistent collaboration relationships.

ShareMem separates transferable experience from user-bound preferences and studies two-stage consolidation and scoped retrieval [@hu2026sharemem, Section 3 and Appendices A–B]. It motivates recording coverage, provenance, and construction costs. Its results do not determine how continuous relationship updates interact with experience updates in the software-task setting proposed here.

EDV separates heterogeneous execution, third-party distillation, and executor consensus verification [@zhu2026edv, Section 4 and Appendix G]. Consensus remains a model judgment with possible shared bias. Astrocyte therefore proposes separate records for model agreement and independently checkable execution evidence.

MAAGL already uses historical trajectories for confidence estimation, collaborator selection, and experience learning in graph tasks [@qu2026maagl, Sections IV-C–IV-D]. Historical experience guiding partner selection is consequently an existing mechanism. The proposed distinction here is a controlled interaction study with condition-specific directed relationship updates over sequential software tasks. This is a research target, with novelty and effectiveness still requiring assessment. The cited author list follows the paper's HTML and PDF title pages; the abstract-page indexing discrepancy remains documented in the source audit.

DecentMem maintains agent-private exploitation and exploration memory pools, reuses collaboration trajectories, and updates online routing from stage-wise LLM judgments [@hao2026decentmem, Sections 3–4]. Decentralized memory combined with collaboration experience is therefore also established prior work. Our proposed decision-location comparison gives central and local decision makers the same designated information view; it does not independently manipulate private versus shared memory architecture. The theoretical claims in DecentMem have their own search-space and reward assumptions, which cannot be transferred to the changing authorized candidate sets and discounted evidence proposed here. Neither a generic memory combination nor the word decentralization establishes this draft's novelty.

Market-based post-training jointly trains action and bidding policies and evaluates population changes [@xu2026mpt, Sections 2 and 3.4]. Astrocyte proposes external-state adaptation with fixed model parameters and separately manipulates decision location. Member perturbations are a relevant evaluation dimension, with their effects to be measured under the present protocol.

Darwin Gödel Machine maintains an archive of code-modified agents and evaluates candidate changes [@zhang2025dgm, Sections 3–4 and Appendix F]. It motivates the prospective iteration case and the need to protect evaluation from candidate-controlled instrumentation. EvoMap's official Swarm documentation describes shared memory and history-sensitive collaboration [@evomapSwarm, “Shared Memory” and “Intelligent Orchestration”]. That documentation provides design context; it supplies no causal estimate for Astrocyte.

The reviewed sources establish substantial prior work on experience reuse and adaptive coordination. This draft makes no claim to an exhaustive literature review or priority for combining memory with collaborator selection. The intended evidence would concern the specified update interaction and its limits.

# Setting and state

Let a task sequence contain $T$ scheduled tasks in a fixed order. A member has an identity, a known capability version, a current availability state, and an authorization scope. A decision $d$ is associated with scheduled position $t$ and an authoritative within-position order. Its recorded input contains only information available before submission. Foundation-model versions and inference settings are recorded; no parameter training occurs in the main experiment.

The adaptive state consists of an experience collection $E_t$ and relationship records $R_t$. Immediate state $Z_t$ includes task progress, outstanding claims, current load, availability, and permissions. Separating these states is essential: disabling learning must preserve operational permission checks and changes in availability.

## Conditions observable before a decision

Propose a fixed, versioned function $x_d=F_v(I_d)$ of the member's authorized pre-decision information $I_d$. The record retains the source task/graph/environment versions, extraction position, observation order, and component-wise missingness. Final patches, later failures, and sealed evaluation never fill in an earlier $x_d$. An approved task-metadata revision affects subsequent decisions and keeps its own source version; an earlier decision and its relationship key remain immutable.

The following is one concrete condition catalogue for software-task development. It is a method proposal, not an approved product schema or registered set of categories. The same approved catalogue and extraction code would be fixed across all eight configurations.

| Component | Proposed extraction from pre-decision sources | Proposed grouping value |
| --- | --- | --- |
| Dependency structure | Count direct predecessors in the approved task-graph version; retain the full declared predecessor set | Zero, one, or multiple predecessors |
| Potential write overlap | Apply the project's existing path normalization and overlap rule to declared write scopes in that graph, before inspecting actual patches | Declared overlap present or absent |
| Required tools | Read the explicitly declared tool/capability families from the task version; retain their version requirements | Exact finite set of families in the approved catalogue |
| Modification kind | Read a declared task label, rather than classify the completed change | Documentation, configuration, source, tests, mixed, or other declared kind |
| Environment | Read the declared language/runtime family and required environment version | Exact approved environment class |
| Verification obligation | Read the permitted feedback-check or human-rubric kind and its version, without sealed answers | Exact approved obligation class |

Every component also has a distinguished unavailable value $\bot$. An explicitly declared empty tool set differs from an unavailable declaration. Missing paths do not mean disjoint paths, and unknown compatibility does not mean compatibility. A malformed value or a known value outside the approved catalogue stops extraction for the affected decision. A permitted missing value remains $\bot$; it is not imputed from outcomes or a selected collaborator's performance. Missing information required for authorization, capability, or safe execution prevents assignment under the operational gate.

Define the condition scope $c_d=g_v(x_d)$ as the exact tuple of these grouping values together with the approved project/sharing scope and extractor version. Distinct missingness patterns remain distinct groups. This rule uses no learned outcome-dependent boundaries and no automatic pooling into a neighboring group. Scope identifiers denote this semantic tuple; this proposal does not require a new hashing or storage protocol. A different catalogue is a new method version, not an unrecorded within-study adjustment.

The versioned $x_d$ supports both experience-applicability checks and relationship grouping. Any additional structural or semantic retrieval score uses its declared frozen representation and is recorded separately. Its model, normalization, and utility coefficients remain registration fields. Current load, active write conflicts, availability, and live permissions stay in $Z_d$ and are checked for each decision; the historical condition group cannot replace them. The proposed catalogue and grouping have not been implemented or empirically validated in the inspected fixed policy.

# Proposed method

## Versioned experience and use records

An experience unit records the task conditions, suggested actions, collaboration arrangements, participant capability versions, source events, verification evidence, failure conditions, cost, use history, and version lineage. Task experience describes problem-solving or tool-use procedures. Collaboration experience describes role assignments, handoff requirements, and checks on contributions.

A retrieval hit establishes availability of an experience. Actual adoption requires a use record linking its exact version to a submitted decision and subsequent execution. This distinction permits measurement of exposure, adoption, and downstream outcomes without counting all retrieved entries as used.

The research draft distinguishes unverified, single-task-verified, and multi-condition-verified experience. Unverified entries enter an isolated verification queue. Ordinary decisions retrieve only entries meeting the applicable evidence requirement **and** current authorization. In the development defaults, one independent successful task supports single-task verification; at least three independent tasks covering two predefined conditions support multi-condition verification. These are draft operating rules, with no claim that three tests establish general reliability.

An applicable counterexample suspends ordinary recommendation and prompts revision and revalidation. Repeated checks on one task do not create independent task coverage. Evidence records distinguish automated tests, checkable tool results, blinded human evaluation, and model judgments.

Evidence strength and permission to use an entry are separate concepts. Existing human approval must remain effective. A future evidence-level implementation cannot automatically grant the current product's approval or cross-project sharing permission. This separation requires an explicit product decision before execution.

## Directed relationships and specific obligations

Let a persistent relationship key be $\rho=(i,v_i,j,v_j,h,o,c)$: requester and collaborator identities, their known capability versions, collaboration type $h$, obligation class $o$, and condition scope $c$ captured at the originating decision. Obligation classes permit accumulation over comparable tasks. A specific obligation instance additionally binds the actual task, handoff, target artifact/source version, required checks, and resolution rule. It is not interchangeable with its class.

Use one outcome unit $u$ per task and relationship key. Fix its aggregation and closure rule at first use; each constituent handoff declares its target obligation before execution. A terminal pass requires the declared target obligations to pass; a terminal failure requires an independently checkable breach under that resolution rule. An unresolved check, unclosed retry, or uncertain attribution leaves the unit unadjudicated. Several tests or retries inside the same task/unit do not create several complete relationship counts. Their executions and costs remain recorded. Different obligation classes require separate predeclared keys, not classes invented after observing which check failed.

While task work for a unit remains open, each real retry or new target retains its own source and artifact version under the same declared resolution rule. At the task-resolution cutoff, that rule fixes which final target versions the single unit verdict assesses; an unfinished independent check can leave adjudication pending after this cutoff. Subsequent corrections reassess those same fixed targets and sources. New work or a new artifact after closure is not a correction that overwrites the earlier target, and cannot manufacture an additional count for the same task/key. Its treatment requires the predeclared task/obligation rules, with every execution retained.

A verdict links the exact obligation instance, its captured $\rho$, target version, actual source event, check/rubric version, result $r_u\in\{0,1\}$, and evidence weight $q_u\in[0,1]$. The development primary rule uses $q_u=1$ for an admissible independent target check. Model opinion, executor self-report, ordinary process exit, and unresolved attribution supply no successful or failed count. Human checks identify the target and use a fixed configuration-blinded rubric. Task-level success does not credit every relationship in a chain, and an infrastructure interruption is not a breach by every collaborator. Attribution to a particular obligation identifies what was checked; it does not identify that collaborator's causal contribution to total task success.

These are semantic research definitions to be mapped to existing task, operation, artifact, and event identities. This manuscript creates no events or replacement execution authority. Evidence strength, access to the source, and current human permission to reuse experience remain separate requirements.

## Scheduled position, visibility, and corrections

Use the predeclared task position $t=1,\ldots,T$ as the decay clock. It advances with the common scheduled sequence, including failed or unattempted scheduled positions, rather than the number of successful completions, model calls, messages, or relationship updates. Wall-clock time and waiting are measured separately. The index uses existing task order and does not introduce a scheduler or timer service.

For an assessed outcome, $b_u$ is the logical position of the underlying obligation outcome/target version. For verdict record $e$, $a_e$ is the first logical position at which that record and its required provenance are validated and available at the existing authoritative record source for the authorized update process. This is not the time at which every member has read it. Require $b_u\leq a_e$. Missing or contradictory timing/provenance stops the affected update; an arrival time is not substituted for an unknown $b_u$. A correction of the same target retains $b_u$. A new target is not silently renamed as a correction of an older one.

A decision at position $t$ uses only verdicts with $a_e<t$. Newly available verdicts at $t$ wait until a later position, including when execution within $t$ has several decisions. This is a necessary visibility gate, not a guarantee that the decision has received all eligible records. Record the existing authoritative within-position order for arrivals and decisions. For records first available at the same position, the authoritative order determines the correction chain; transport arrival order is not a replacement for it. Live availability, permissions, and claim state may change immediately within the position through $Z_d$.

For each unit, an initial adjudication and explicitly linked authorized corrections form one validated chain. Unverified opinions or pending check reports do not enter that chain and cannot replace an existing verified verdict. A withdrawal identifies the verdict, source and authorized reason it invalidates. Duplicate deliveries retain the same evidence identity and first-availability position. Repeated test reports for a unit remain evidence for its single adjudication.

Let $V_d$ describe the version and actual record collection read and authorized for decision $d$ before submission. It preserves existing source identities/versions, snapshot production position or query scope, the authoritative order and parent links represented, and the actual observation/read time. A version label alone is insufficient: the retained source material must identify the collection and permit checking its contents. This notation describes an input view, not a new version service. Let $H_d$ be its admissible verdict records: source/provenance and current access are valid, $a_e<t$, and each included correction's required parents are present in that view or its provenance-preserving initial snapshot. A missing parent, known omitted eligible successor with $a_e<t$, inconsistent target, contradictory time, or competing incomparable corrections stops the affected score/assignment. The process cannot silently choose an older convenient chain.

For an enabled relationship channel, select each unit's unique latest verdict within this actual parent-closed collection and authoritative order. A latest withdrawal contributes nothing; a unit with no available adjudication remains pending. A known eligible withdrawal/correction cannot be omitted to retain an older score: incorporate its validated closed chain or stop the affected assignment. A later correction not actually read in $V_d$ is unknown to this decision; record the observed source version and its real age without claiming a globally latest state. Later read-only audit may establish additional lag, but it cannot supply future knowledge to the original decision. Current permission revocation is checked independently even when a learned channel is frozen.

Let $\mathcal U_\rho(t;V_d)$ denote active units for $\rho$ selected from $H_d$. Counts below use this actual set and the current decision position $t$, not the clock position at which a cached snapshot was produced. A stale but complete permitted snapshot can omit unknown later evidence; it still ages its represented units from their original $b_u$ to the present $t$ when relationship updates are enabled. An aggregate without the required unit decomposition cannot establish the specified correction/age semantics. The disabled channel instead uses its position-zero projection, with no such re-aging. Late verdicts and corrections alter later views; they never rewrite a submitted decision, its captured conditions, or its reported propensity. If an old condition record is invalidated, withdraw its contribution prospectively under the enabled projection; do not relabel it into a group chosen using later outcomes.

## Discounted state and initialization

For $0<\lambda\leq1$, define nonnegative counts from the effective visible verdicts:

$$
s_\rho(t;V_d)=\sum_{u\in\mathcal U_\rho(t;V_d)}
q_u r_u\lambda^{t-b_u},\qquad
f_\rho(t;V_d)=\sum_{u\in\mathcal U_\rho(t;V_d)}
q_u(1-r_u)\lambda^{t-b_u}.
$$

With fixed $\alpha_0,\beta_0>0$, set

$$
w_\rho(t)=\frac{\alpha_0+s_\rho(t)}
{\alpha_0+\beta_0+s_\rho(t)+f_\rho(t)}.
$$

Here and below, $s_\rho(t),f_\rho(t),w_\rho(t)$ suppress $V_d$ when its identity is fixed explicitly by the decision. Equal $t$ alone does not imply equal views or counts.

Each unit appears once with its effective verdict. Replacing or withdrawing a verdict recomputes its contribution to this finite sum; it does not append an extra success/failure. Delayed evidence is discounted from $b_u$, not rejuvenated at $a_e$. With the same represented effective units and verdicts and no new evidence, both counts multiply by $\lambda^{t'-t}$ between $t$ and $t'$. The original one-event recursion is recovered by adding a newly visible unit's already-aged term to those decayed counts. A zero-weight record adds no evidence while elapsed-position decay can still occur.

At position zero, the common nonempty historical snapshot contains approved experience versions and historical units with source, target, effective verdict, and documented occurrence positions $b_u\leq a_e<0$. Pre-study history uses the same registered ordinal clock with zero as the study boundary; its verdicts and provenance are available before that boundary. Snapshot counts are the same sums evaluated at zero. Restoring a snapshot retains unit identities and correction lineage, so historical events are not counted again as new evidence. Initial nonzero aggregates require per-unit provenance/decomposition for later withdrawal or correction; aggregate-only history is insufficient for this proposal. Historical positions cannot be invented to retrofit an existing store. An enabled view retaining the same initial units and unchanged verdicts gives their contribution at $t$ as exactly $\lambda^t s_\rho(0)$ and $\lambda^t f_\rho(0)$.

The development illustrations retain $\alpha_0=\beta_0=1$ and $\lambda=2^{-1/20}$, interpreted as a half-life of 20 scheduled positions. These are unregistered values. A newly eligible, fully identified capability version with no evidence has zero counts and the defined prior-like score $\alpha_0/(\alpha_0+\beta_0)$. Propose no automatic transfer across capability versions or condition-schema versions. An unannounced change remains unobserved until evidence reveals it; earlier decisions keep the capability information they actually had. No version or evidence update changes current permissions.

When relationship updates are off, decisions read the position-zero counts and weight without decay, new verdicts, or corrections entering the learned state. Logs remain writable. A live withdrawal of permission still blocks unauthorized reads/use through the operational gate; freezing a learned representation cannot authorize it. When experience updates are off, content, evidence level, historical use/ranking inputs, and associated adaptive caches remain frozen. Enabled experience revisions similarly become eligible only after their declared availability position and current approval checks.

These scores summarize selected, discounted evidence. Unique task accounting does not make observations independent, and the prior-like ratio is not a calibrated Bayesian posterior. The complete-block analysis and its sampling limitations remain those in Section 6.

## Common decision rule

For eligible candidates $C$, the proposed `policy_score` combines normalized task fit, experience support, relationship score, estimated cost, and current load. Development defaults add the first three and subtract the last two, with coefficients one. Accurate feature definitions, aggregation, and cost prediction remain registration fields.

For candidate $j$ with score $U_j$, choose according to

$$
p(j)=(1-\epsilon)
\frac{\exp(U_j/\tau)}{\sum_{k\in C}\exp(U_k/\tau)}
+\frac{\epsilon}{|C|}.
$$

The draft defaults are $\epsilon=0.10$ and $\tau=1$. Candidate ordering, score components, probabilities, random draws, selected experience versions, observation time, state version, submission time, and claim outcome must be observable. An empty candidate set produces an unassigned decision under the common waiting rule. Exploration operates inside capability and authorization constraints.

Required but undefined condition, permission or utility inputs prevent the affected assignment with an explicit error. They are not replaced with guessed zero scores or an unrecorded alternative policy. The defined zero-evidence state for a known, eligible relationship key is separate from missing identity or unreadable mandatory state. Current authorization is checked again by the common claim/execution service at submission.

Central and local decisions call the same implementation. For a decision assigned to member $i$, the central service receives that member's actual information/record view $I_d,V_d$, candidate set, retrieval allowance, and decision budget; access to an authoritative source does not add records to that designated input. Both use the same claim and conflict-handling service and the same total decision-concurrency allowance. The central service submits choices for members; local members submit their own choices.

## Conditional properties and proofs

The following properties describe the proposed definitions under stated assumptions. They are elementary algebra and coupling arguments, not a claim of original general theory, empirical benefit, calibrated prediction, regret control, or causal identification. The inspected fixed policy has not acquired these properties merely because the specification states them.

### Canonical evidence and duplicate delivery

**Claim.** Fix a decision position, relation key, the actual admissible collection $H_d$ from a specified $V_d$, its authenticated parent-closed correction chains, and its occurrence positions. Assume every unit has one unambiguous effective verdict or withdrawal. The counts are invariant to the order in which those records are enumerated and to repeated delivery of a record with the same evidence identity.

**Proof.** Each unit's effective verdict is a function of the fixed visible chain and authoritative order, not the enumeration order. Repeating an identity does not change that chain. Each active unit therefore supplies the same single two-component contribution. Finite addition is commutative and associative, so the two sums and their ratio agree. This compares the same visible set. Different first-availability positions can change an online decision and are not covered by the claim.

### Sensitivity of a relationship score

**Claim.** Compare two nonnegative states for the same relation key and the same positive priors. Put $m_0=\alpha_0+\beta_0$. Then

$$
|w-w'|\leq\min\left\{1,
\frac{|s-s'|+|f-f'|}{m_0}\right\}.
$$

The following evidence-mass refinement additionally assumes that both states are generated by the enabled relationship-update projection at the same position $t$, with the same ordinal clock and $\lambda$. For each unit in the union of their effective sets, let $v_u=(q_u r_u\lambda^{t-b_u},q_u(1-r_u)\lambda^{t-b_u})$ be its actual contribution under that state's effective verdict, using the zero vector when absent; define $v'_u$ from the other state's effective verdict. Thus $(s,f)=\sum_u v_u$ and $(s',f')=\sum_u v'_u$. If $M=\sum_u\|v_u-v'_u\|_1$, the numerator above is at most $M$. This refinement does not compare an enabled, decaying state with a relationship-off frozen snapshot. That comparison can use the original bound with each projection's actual $s,f$; identical historical identities need not then have identical mass.

**Proof.** Write $D=m_0+s+f$. The derivatives of $w=(\alpha_0+s)/D$ are $(\beta_0+f)/D^2$ and $-(\alpha_0+s)/D^2$, whose absolute values are at most $1/m_0$. The segment joining the two nonnegative states stays in that domain. Integrating the derivative along the segment gives the stated Lipschitz bound; $0<w,w'<1$ gives the other bound. The triangle inequality applied to the contribution vectors gives the $M$ bound. Missing evidence has its own discounted mass; a success-to-failure correction changes both components and can contribute twice that mass to $M$. The bound does not treat such a correction as a single missing count.

### Sensitivity of the fixed-candidate choice distribution

**Claim.** Let the same nonempty finite candidate set, ordering, $\tau>0$, and $0\leq\epsilon\leq1$ be used in two decisions. If all utilities are finite and $\|U-U'\|_\infty\leq\eta$, their distributions satisfy

$$
\operatorname{TV}(p,p')\leq
(1-\epsilon)\min\{1,\eta/\tau\}.
$$

**Proof.** Along $U(z)=U+z(U'-U)$, let $\sigma_j(z)$ denote the softmax term and $\Delta=U'-U$. Differentiation gives

$$
\frac{d\sigma_j}{dz}=
\frac{\sigma_j}{\tau}\left(\Delta_j-
\sum_k\sigma_k\Delta_k\right).
$$

Because $|\Delta_j|\leq\eta$, the sum of absolute derivatives is at most $2\eta/\tau$. Integrating and dividing by two bounds total variation by $\eta/\tau$; total variation is also at most one. The identical uniform exploration component cancels, leaving the factor $1-\epsilon$.

If only an affine relationship term $c_R w_\rho$ changes and all other utility inputs/scales are fixed, the preceding evidence-mass refinement supplies $\eta\leq |c_R|\max_\rho\min\{1,M_\rho/m_0\}$ only when both states meet its same-position, same-clock, enabled-projection assumptions. For an enabled/frozen comparison, use each projection's actual counts in the original score bound to obtain a utility bound instead. A changing normalization must be included in the utility error. Candidate entry/removal, changed authorization, different retrievals, or unequal budgets violate the fixed-input premises; the bound then supplies no task-quality or deployment conclusion.

### Equal information and equal random draws

**Claim.** With the same deterministic policy implementation, eligible ordered candidates, identical actual information/state views $I_d,V_d$ and their admissible collections/projections, finite utilities, and the same uniform draw, central and local decision makers select the same candidate.

**Proof.** Their score vectors and probability vectors are identical. Their cumulative probabilities in the common ordering are identical, so inverse-CDF selection with a fixed boundary convention yields the same candidate for the same draw. Induction extends the statement over decisions only if initial states and all intervening exogenous inputs, evidence availability/corrections, authorization and claim outcomes are also coupled identically. Different provider outputs or request/claim timing are not made equal by a requested seed. Synchronous replay checks this conditional equivalence; it cannot establish a positive effect of location or equal realized performance under a different deployment.

### History excluded by disabled updates

**Claim.** Fix task inputs, operational state, initial snapshots, policy/random draws, and the enabled channel's allowed state. Changing only the post-initialization history of a disabled channel does not change decision probabilities under the specified state projection.

**Proof.** For a disabled relationship channel, the policy receives the initial counts/weight with no decay or historical corrections. For a disabled experience channel, it receives initial content, verification/use/ranking values and adaptive cache inputs. The changed later history enters neither projection. All remaining policy arguments are fixed, so the probability vector is fixed. This is conditional on enforcing the projection at every read, including caches and summaries. If another enabled channel actually changes, or current permissions/availability change, the premises do not hold. This specification is not evidence that the product already enforces those exclusions.

## Feedback cycle and failure handling

After an executed task, the allowed feedback suite produces obligation-specific evidence. The experience process proposes and verifies revisions within its assigned budget. Validated verdicts, withdrawals and corrections enter their recorded availability positions. Later decisions use the effective earlier-visible set and the projection allowed by their configuration. The intended cycle is specified below; it is not an executed algorithm trace.

```text
For each scheduled task in an isolated configuration:
    obtain current operational state and permitted earlier-visible adaptive state
    capture versioned pre-decision conditions, obligations, and eligible candidates
    retrieve eligible experience and compute the common decision probabilities
    submit the selected task or partner decision through the common claim service
    execute within the registered task and sequence budgets
    record actual feedback sources, occurrence/availability, usage, and unresolved units
    verify experience and adjudications under current permissions and declared budgets
    if updates are enabled: expose valid revisions at their allowed future positions
    submit the final artifact to the isolated evaluator
```

Unavailable shared services prevent affected new decisions and produce an explicit failure or waiting record. Recovery must follow the registered policy, including any explicitly selected snapshot mode and its version. No unobserved retry, substituted state, or unknown-delivery resend can be treated as successful execution. Duplicate evidence handling reuses event identity; it does not create a separate scheduling authority.

# Experimental protocol draft

## Eight paired configurations

Every replication contains the following eight configurations in randomized execution order. All begin from the same nonempty experience and relationship snapshots, copied into isolated condition state, and process the same scheduled task sequence.

| Configuration | Decision location | Experience updates | Relationship updates |
|---|---|---|---|
| L00 | Local | Off | Off |
| L10 | Local | On | Off |
| L01 | Local | Off | On |
| L11 | Local | On | On |
| C00 | Central | Off | Off |
| C10 | Central | On | Off |
| C01 | Central | Off | On |
| C11 | Central | On | On |

All conditions can retrieve initial experience. Experience-off freezes content, verification state, historical use statistics, and ranking terms derived from those statistics. Relevance to a current task may be recomputed from the frozen representation. Relationship-off freezes counts, weights, and time decay. Immediate task state, permissions, and availability continue to change in every condition. Logs remain writable, but a disabled channel cannot update its decision inputs from those logs through caches, indexes, summaries, or shared services. An enabled experience channel can still learn from permitted feedback while relationship counts are frozen; an enabled relationship channel can still update while the experience representation is frozen.

Each isolated copy includes historical unit identities, effective verdicts, occurrence positions, capability/extractor versions, and correction lineage, not only aggregate counts. New visible verdicts/revisions are gated separately by the two switches. The disabled-channel property above is conditional on holding the other policy inputs fixed; the eight configurations do not assume that an enabled channel or operational trajectory remains unchanged. Current permission withdrawal blocks use without silently rewriting frozen historical weights.

Model lists, prompts, source collections, feature extraction, task order within each paired block, concurrency quotas, and task/sequence budget rules are shared. New literature or other knowledge cannot enter only selected conditions. Initialization and execution order are recorded together with service versions and calendar time to expose drift and shared-service interference.

## Decision location and identification

A synchronous development replay supplies equal information snapshots, event order, candidate identities, and random draws to both implementations. Matching selections test implementation consistency. With equal state and random draws, the common rule implies equal decisions; replay supplies no evidence for a positive location effect.

The deployed comparison records each actual $V_d$, its represented source collection, snapshot production position, read/submission times, request waiting, claim ordering/conflicts, and communication. Observed snapshot age and authoritative first availability are distinct; unknown later corrections cannot be labeled known at the decision. Different actual views can yield different scores even at the same $t$, while an enabled projection always ages its represented evidence to that decision's $t$. A nonzero location effect would describe those specified deployment conditions. It cannot be attributed to abstract decentralization while central caching, extra information, different scoring, or unequal compute also vary. Shared storage, process placement, cache policy, communication path, and scheduling must therefore be declared. Whether the deployment introduces a meaningful location intervention remains a readiness question.

## Tasks, feedback, and sealed evaluation

Each task has a `feedback_suite` and an `evaluation_suite`. Feedback is available to execution and the permitted updates. Evaluation answers, item-level outcomes, and scores stay inaccessible to those processes until the entire task sequence ends. An independent evaluation process receives final artifacts and retains its results separately. Ordinary completion messages and retrieval indexes must not return sealed scores indirectly.

Task and suite selection are unfinished. The registration record must identify exact versions, permissions, overlap between suites, and the public-project scope. Isolating results does not guarantee independence when the feedback suite reveals the same assertions or target answers; overlap and adaptation to public tests require explicit review.

The primary task outcome is binary satisfaction of the fixed evaluation suite. Errors, timeouts, budget exhaustion, and unfinished tasks count as failures. After a sequence exhausts its budget, all remaining scheduled tasks remain in the denominator. Damaged task data or missing evaluation programs follow predeclared, configuration-symmetric exclusion rules. Algorithmic failures, unhelpful experience, invalid tools, waiting, and retries remain outcomes and costs. A retry never erases its original execution record.

## Budgets and accounting

Accounting covers model calls, tools, experience extraction, retrieval, verification, coordination, retries, and initialization. The unverified-experience queue has a draft ceiling of 5% of its configuration's total budget; its work is charged to that configuration. This ceiling and all other budgets require confirmation.

Model usage distinguishes input, cached input, output, and reasoning tokens when observable, by exact model version and billing period. Payment and resource usage are separate: a subscription-covered request can have no incremental payment while consuming resources. Unknown usage remains unknown and cannot be recorded as zero. Missing budget-relevant evidence blocks a formal budget-matched claim.

Communication records message count, bytes, model-context tokens, waiting time, and elapsed time. Snapshot production is reported as an actual shared preparation expense and as an explicitly amortized per-run amount. The allocation rule must be registered; generating a snapshot once does not create eight independent preparation events. Conversely, reusing a public snapshot does not justify hiding its production cost in end-to-end comparisons. Paired executions remain isolated even when they start from the same snapshot.

Let the main study use $r$ replications of $T$ tasks, perturbations use $r_p$ replications of $T_p$ tasks in three configurations and four scenarios, and the case use $r_s$ replications of three processes over $G$ rounds and $T_s$ evaluations per round. The draft task-accounting identity is

$$
N_{\mathrm{tasks}}=8rT+12r_pT_p+3r_sGT_s+N_{\mathrm{pilot}}+N_{\mathrm{initial}}.
$$

Case candidate generation and development verification are charged additionally. The input's $r=5,T=10$ example gives 40 sequences and 400 task executions; its assumed 30 calls per task gives approximately 12,000 calls before preparation and development work. These are budget illustrations, with no observed call rate or approved formal sample size.

# Estimands and analysis

## Primary and secondary contrasts

For replication $k$, let

$$
Q_{ab,k}^{P}=\frac{1}{T_k}\sum_{u=1}^{T_k}Y_{ab,k,u}^{P},
\qquad P\in\{\mathrm{local},\mathrm{central}\},\quad a,b\in\{0,1\}.
$$

$T_k$ is the common post-exclusion scheduled denominator across the eight configurations. We compute each replication's contrasts first, then average contrasts equally across replications. The planned position-specific interaction is

$$
I_{\mathrm{feedback}}^{P}=Q_{11}^{P}-Q_{10}^{P}-Q_{01}^{P}+Q_{00}^{P}.
$$

The primary target is $I_{\mathrm{feedback}}^{\mathrm{local}}$. The draft practical target $\delta_{\mathrm{feedback}}=0.03$ means three additional successes per hundred scheduled tasks beyond the additive expectation from the individual updates. It does not mean that L11's total gain over L00 is three points.

Decision-location modification of the interaction is

$$
I_{\mathrm{three}}=I_{\mathrm{feedback}}^{\mathrm{local}}
-I_{\mathrm{feedback}}^{\mathrm{central}}.
$$

Decision-location modification of the total joint-update gain is

$$
I_{\mathrm{position}}=(Q_{11}^{\mathrm{local}}-Q_{00}^{\mathrm{local}})
-(Q_{11}^{\mathrm{central}}-Q_{00}^{\mathrm{central}}).
$$

These are distinct estimands. In each position,

$$
Q_{11}^{P}-Q_{00}^{P}=I_{\mathrm{feedback}}^{P}
+(Q_{10}^{P}-Q_{00}^{P})+(Q_{01}^{P}-Q_{00}^{P}).
$$

Consequently, the location difference in total gain also contains location differences in the two individual update gains. A large L11–L00 improvement alone does not establish interaction. We additionally report L11–L10, L11–L01, and L11–C11 together with costs. Positive interaction and acceptable operational value are separate requirements.

## Replication and uncertainty

The independent unit is a complete paired block of eight task sequences, with separately executed requests and reinitialized condition state. For block $k$, store the full vector $V_k=(Q^{L}_{00,k},Q^{L}_{10,k},Q^{L}_{01,k},Q^{L}_{11,k},Q^{C}_{00,k},Q^{C}_{10,k},Q^{C}_{01,k},Q^{C}_{11,k})$. Each estimand above is a fixed linear contrast $D_k=h(V_k)$, estimated by $\hat\theta=r^{-1}\sum_{k=1}^{r}D_k$. All blocks receive equal weight. A pooled task mean would change this estimand when post-exclusion denominators differ.

Tasks within a sequence are dependent through adaptation. Tests, seeds, eight configuration labels, and repeated analysis do not add independent blocks. The nonparametric proposal treats the $V_k$ as independent, identically distributed observations from the declared replication population. Repeated use of a fixed task list and one development snapshot supports conditional inference for those inputs. Restoring separate state copies does not independently sample historical experience. Generalization across tasks or histories requires corresponding sampling. Shared-provider drift, overlapping requests, or cross-block state leakage can invalidate the independent-block assumption; a seed label does not repair it.

Only predeclared, configuration-symmetric damaged-task exclusions alter a block's common denominator. Budget exhaustion and algorithmic failures remain scheduled outcomes. Missing measurements or an incomplete block stop the specified paired analysis; they are not silently deleted, completed from other conditions, or imputed as observed failures. All original records remain available for accounting and reporting.

## Proposed complete-block confidence procedure

The following BCa procedure is a research proposal, with its computational settings and suitability to be assessed before formal collection. It adapts the bias correction, acceleration, and confidence-point construction of DiCiccio and Efron [@diciccio1996, Sections 2–3 and 6] to the block contrasts defined here. It has not been executed on Astrocyte pilot or formal outcomes.

Draw $B$ index vectors of length $r$ independently with replacement from $\{1,\ldots,r\}$. Apply each index vector to entire $V_k$ rows and reuse those indices for every contrast. For each contrast retain every resampled mean $\hat\theta_b^*$. Let $\hat G$ be their right-continuous empirical distribution, with $\hat G^{-1}(u)=\inf\{x:\hat G(x)\geq u\}$. A resample with constant $D_k$ values remains a legitimate mean. No task-level or configuration-wise resampling is performed.

Let $\Phi$ denote the standard normal distribution function and $z_u=\Phi^{-1}(u)$. Compute

$$
\hat z_0=\Phi^{-1}\!\left(B^{-1}\sum_{b=1}^{B}
\mathbf 1\{\hat\theta_b^*<\hat\theta\}\right).
$$

For $r>1$, let $\hat\theta_{(-k)}$ be the mean omitting complete block $k$, let $\bar\theta_{(-\cdot)}=r^{-1}\sum_k\hat\theta_{(-k)}$, and set $d_k=\bar\theta_{(-\cdot)}-\hat\theta_{(-k)}$. The jackknife acceleration and adjusted endpoint are

$$
\hat a=\frac{\sum_k d_k^3}{6(\sum_k d_k^2)^{3/2}},
\qquad
A(u)=\Phi\!\left(\hat z_0+
\frac{\hat z_0+z_u}{1-\hat a(\hat z_0+z_u)}\right),
$$

$$
q(u)=\hat G^{-1}(A(u)),\qquad
C_{1-2\gamma}=[q(\gamma),q(1-\gamma)].
$$

The nominal 95% interval is $C_{0.95}$ with $\gamma=0.025$. Its lower endpoint above the draft $\delta_{\mathrm{feedback}}=0.03$ supports the proposed practical primary target; an upper endpoint below that target indicates it is not reached under the interval rule. An interval crossing or touching the target is inconclusive. A 95% two-sided lower endpoint corresponds to a 97.5% one-sided bound, which is the boundary used in the planning example.

The procedure requires a nondegenerate contrast distribution and applicable independent-block asymptotics. Bounded task scores give finite moments, but discrete small-block means can violate the smooth approximations used for higher-order bootstrap claims. No exact finite-sample coverage or automatic second-order accuracy is asserted for this design.

Before formal data, specify $B$, the randomization record, a finite one-sided significance grid $\mathcal A\subset(0,1/2]$, and minimum usable tail ranks. Include the primary level and every one-sided level needed by the multiplicity decisions below. Require finite $\hat z_0$ and $\hat a$, positive $\sum_k d_k^2$, positive denominators in $A(u)$ for all required $u$ and $1-u$, nested finite endpoints on that branch, and the registered minimum ranks in both tails of every adjusted quantile. Record distinct block contrasts, ties, extreme adjusted probabilities, and adjusted quantile ranks. $B$ and these numerical criteria remain unconfirmed, not values selected from formal outcomes.

For computational sensitivity, propose one designated primary Monte Carlo batch and a fixed number $J$ of additional independent batches, each containing $B$ complete-block resamples. Each batch computes its own correction and endpoints; within a batch the indices remain common to all contrasts. Report the endpoint ranges and the grid-inversion bounds across batches. The primary batch supplies the proposed inferential quantities; the others diagnose computational instability. Register $J$ and endpoint-spread tolerances before formal data. A diagnostic failure or disagreement about an intended decision makes the affected inference unavailable; no batch is selected for a favorable answer and no extra batches are added after inspection. These ranges describe Monte Carlo sensitivity, not confidence coverage over new experimental blocks.

Failure of these conditions ends inferential calculation for the affected contrast. Report the observed blocks and point estimate with the reason that inference is unavailable. Do not replace a zero jackknife denominator, clip an infinite correction, remove constant bootstrap samples, change interval methods, or add formal runs after seeing the result. A family containing an unavailable marginal calculation receives no confirmatory family conclusion under this proposal.

This choice also addresses a specific studentization problem. Bootstrap-t divides each resampled centered mean by that resample's standard error [@diciccio1996, Section 5]. If the observed distinct contrast values have multiplicities $n_j$, the conditional probability of a constant resample is $\sum_j(n_j/r)^r$. Such samples can have zero standard error even when the original variance is positive. Deleting, redrawing, or adding a small denominator would define another procedure. BCa uses their means without that division; its own degeneracy and tail checks still apply.

## Null hypotheses and confidence-curve inversion

Bootstrap uncertainty around $\hat\theta$ is not itself a null distribution. Hall and Wilson [@hall1991, Sections 1–3] distinguish a resampling comparison centered at the estimate from an uncentered comparison with the tested value. DiCiccio and Efron [@diciccio1996, Section 3] likewise distinguish an ordinary bootstrap tail area from a hypothesis-testing confidence level. The proposed tests invert the same adjusted confidence endpoints above. They do not label a raw fraction of positive bootstrap contrasts as a calibrated $p$ value, and do not assume a sharp-null permutation or exchangeable configuration labels.

For $H^+_0:\theta\leq\theta_0$, define the grid upper bound $\bar p_+(\theta_0)$ as the smallest $a\in\mathcal A$ such that $q(a)>\theta_0$; if none exists, set it to 1. For $H^-_0:\theta\geq\theta_0$, define $\bar p_-(\theta_0)$ using $q(1-a)<\theta_0$. The value 1 here records non-rejection over a valid grid; a failed diagnostic remains unavailable. Strict endpoint comparisons handle ties without jitter. A value at the smallest grid level is reported as an upper bound at that resolution, never as an exact zero $p$ value.

This is a conservative grid coarsening of inversion of a nested confidence curve: at a grid level $a$, $\bar p_+(\theta_0)\leq a$ holds exactly when the lower endpoint excludes $\theta_0$ in the required direction. The analogous statement holds for the upper endpoint. Its statistical calibration inherits the approximation of the BCa confidence curve; grid coarsening is not an independent proof that the underlying curve is calibrated. Computational tail resolution and sampling error are distinct.

A two-sided difference calculation uses

$$
\bar p_{\mathrm{diff}}(\theta_0)=
\min\{1,2\min(\bar p_+(\theta_0),\bar p_-(\theta_0))\}.
$$

For a test at level $a$, include $a/2$ in $\mathcal A$. The primary practical test uses $H_0:\theta_{\mathrm{primary}}\leq\delta_{\mathrm{feedback}}$ at the proposed one-sided level 0.025, equivalently the strict lower-95%-interval rule. Numerical and sampling-calibration requirements apply to it as to the secondary calculations. No calibrated test or rejection is currently available because no independent pilot blocks or formal outcomes exist.

## Difference, equivalence, and multiplicity

Preserve the three-member secondary difference family: $\theta_1=I_{\mathrm{three}}$, $\theta_2=I_{\mathrm{position}}$, and $\theta_3=Q_{11}^{\mathrm{local}}-Q_{11}^{\mathrm{central}}$. Apply the two-sided calculation at zero to each. For sorted marginal upper bounds $\bar p_{(1)}\leq\bar p_{(2)}\leq\bar p_{(3)}$, Holm's step-down adjusted values are

$$
\bar p^{\mathrm{adj}}_{(j)}=
\min\{1,\max_{i\leq j}(3-i+1)\bar p_{(i)}\}.
$$

Reject sorted hypotheses sequentially while $\bar p_{(j)}\leq\alpha_D/(3-j+1)$, stopping at the first failure. Holm's theorem controls familywise error under arbitrary dependence when the marginal $p$ values are valid [@holm1979, Sections 1–3]. Here its guarantee remains conditional on adequate calibration of the proposed marginal confidence curves. Applying Holm cannot repair invalid small-sample inputs. The nominal BCa intervals are marginal intervals, not Holm-adjusted simultaneous intervals.

For each of these three location contrasts, the draft equivalence bounds are $\ell=-0.02$ and $u=0.02$. TOST tests the union null $\theta\leq\ell$ or $\theta\geq u$, using

$$
\bar p_{\mathrm{eq}}=\max\{\bar p_+(\ell),\bar p_-(u)\}.
$$

Both one-sided nulls must be rejected. At a common marginal level $a$, inversion corresponds to $C_{1-2a}$ lying strictly inside $(\ell,u)$; touching a bound does not establish equivalence. Thus each-side 0.05 corresponds to a nominal 90% interval, whereas a complete 95% interval inside the bounds is the stricter each-side 0.025 rule. This test/interval relationship follows the TOST construction [@schuirmann1987, pages 657–662 and 675]. Its original normal-theory setting supplies no finite-sample guarantee for our block-bootstrap application.

An additional three-member equivalence family is proposed, not yet approved as confirmatory. If approved, apply Holm to its three $\bar p_{\mathrm{eq}}$ values separately, with budget $\alpha_E$. The recommended joint secondary policy allocates a predeclared secondary error budget $\alpha_S$ equally: $\alpha_D=\alpha_E=\alpha_S/2$. For an illustrative $\alpha_S=0.05$, each family receives 0.025. These are proposed allocation values, not registered thresholds. With valid marginal tests, the union bound then limits the error probability across the six secondary claims by $\alpha_D+\alpha_E$. Running both families at 0.05 would only establish two separate family controls. This proposal does not give joint control over the primary claim plus all secondary claims.

Register the exact families, error budgets, grid levels for every Holm threshold, and decision rules before formal data. A nominal 90% interval inside the equivalence range does not bypass family adjustment. Non-rejection of a difference is not equivalence; either decision can remain inconclusive. L11–L10, L11–L01, perturbations, and process analyses are descriptive unless separately designated and budgeted before collection. The ASA press release motivates transparent effect reporting [@asa2016release]; technical support for these procedures comes from the original articles cited above.

## Sample-size planning

Formal replication counts depend on pilot estimates of the sequence-level interaction variance, task dependence, the minimum useful effect, and budget. The supplied planning example considers a true interaction of 0.06 against a target of 0.03 at 80% power. Under an approximate independent-block normal model, a planning calculation has the form

$$
r\ \approx\left(\frac{(z_{0.975}+z_{0.80})\sigma_I}{0.06-0.03}\right)^2,
$$

where $\sigma_I$ must come from independent pilot block contrasts. The normal approximation concerns independent blocks and the proposed primary 97.5% one-sided bound; it does not prove power for the BCa procedure. Real pilot sequences are needed to assess variance, dependence, discreteness, degeneracy, and usable precision. Coverage or null calibration additionally needs justified sampling assumptions and an independently specified calibration assessment; repeatedly resampling one empirical pilot distribution cannot establish population coverage by itself. No such assessment, pilot variance, or numerical replication count exists here. More Monte Carlo resamples reduce computational error without adding independent experimental blocks. Replication counts, $B$, maximum runs, stopping rules, and numerical diagnostics remain open. Formal outcomes must not determine them retrospectively.

# Adaptation, mechanisms, and prospective iteration case

## Member perturbations

The protocol proposes member exit, entry, announced capability change, and unannounced capability change, comparing L11, L00, and C11. Perturbation times, proportions, identities, and capabilities remain scenario-registration fields. Matched configurations begin from corresponding states and schedules. An exit updates availability immediately; its effect on ongoing work requires real lifecycle evidence.

Measures include task transfer, incorrect delegation, detection delay, relationship adjustment, recovery time, quality loss, and cost. Selection records reveal whether a new member receives opportunities and whether reused experience supports subsequent work. A stopped Mission with no Worker provides no evidence about adaptation to member exit.

## Structural distance and negative interactions

Task distance is computed before execution relative to the initial historical tasks. The draft uses normalized clipped numerical differences, categorical disagreement, and set differences based on intersection and union, averaged over observed features. Insufficiently observed tasks form an unknown group. The exact scale, nearest-history rule, and near/middle/far boundaries must be fixed using development data.

Distance groups share the same tasks across conditions. Group-level L11–L00 gains are reported with coverage and difficulty indicators. Distance is observed rather than randomized, so a gain-distance pattern does not independently establish a causal transfer law.

Process records support examination of harmful experience reuse, concentration of assignments, stale capability versions, delayed updates, and learning overhead. These are candidate explanations. Attribution requires separately labeled interventions, such as disabling a particular experience or changing update visibility in a paired rerun. Replaying decisions cannot reproduce unobserved model or tool outcomes in a different environment. Observational logs alone cannot establish the cause of a negative interaction.

## Swarm A/B/C application case

The proposed case assigns hypothesis formation to A, candidate modification to B, and development verification and independent evaluation to C. Each candidate records a prediction, applicability conditions, code/configuration version, development checks, adoption decision, cost, and lineage. These are functional responsibilities; this manuscript does not claim that autonomous A/B/C iteration has occurred.

Three processes compare persistent cross-round experience, reset cross-round experience, and an unchanged system under matched budgets. Development feedback supports candidate selection. Sealed evaluation is controlled by the evaluator; its results must not influence later candidate generation during the evaluated sequence of rounds. Repeated access to the same final holdout would compromise that separation.

Round-level quality, transfer to new tasks, adopted-improvement cost, and cumulative performance will be reported when measured. A hypothesis-acceptance rate describes the research process and cannot substitute for independent performance. All candidates, failures, and costs must remain in the case record. This first draft contains no invented round table.

# Current implementation and available evidence

The inspected product candidate is `7ce3a5efe312f4cdac3491293e002f5559f076b0`; its business-code source is `982333b871520e606c74d553e448c5049f81abde`. The research writing branch began from a separate repository default and does not imply an updated product checkout. The following evidence levels are kept distinct.

| Level | Current scope | Implication for this manuscript |
|---|---|---|
| Proposed method | Conditional experience, relationship updates, common probabilistic policy, eight configurations | Specified here; complete execution remains unverified |
| Inspected implementation | Mission/WorkItem and lease logic; fixed candidate ranking; human-governed experience; accepted-source linkage | Reusable product components with narrower semantics |
| Independent engineering report | Selected manually created DAG, permission, reconnect, lifecycle-without-Worker, and loading checks | Component evidence reported by a separate agent session |
| Formal research results | Eight-configuration study, member adaptation, autonomous iteration | Not collected |

In `internal/swarm/domain/policy.go`, `swarm.fixed.v1` filters by approval, budget, capabilities, availability of dependencies, and write-path conflicts, then orders by matching inputs and identifier. It does not establish the proposed experience-and-relationship softmax policy. The inspected `internal/attention/domain/experience.go` and application service distinguish quarantine and approved entries with human governance. Accepted-artifact linkage exists in `internal/swarm/app/experience.go`. These findings support an implementation plan; they do not demonstrate realized experience-driven collective adaptation.

The independent report is pinned separately [@astrocyteEngineering2026]. Its accepted scenarios use manually created work items and do not validate model-generated task planning or Worker exit. This research worker read that report and inspected the stated source; it did not rerun the product tests or independently query their event database.

Restricted native execution remains blocked by the requirement to disable native file and command tools. The pending S3 decisions concern call-budget accounting and the evidence required for tools and external destinations; S3 here names the project's Swarm execution stage. They must be resolved for the affected research execution path. Unknown delivery cannot be retried as though unsent. OpenViking's positive remote path, actual work completion/adoption, and subsequent experience reuse remain unverified.

Full CI has not passed. The reported automated runs were ultimately cancelled, and their Ubuntu full-repository Go test jobs had already failed before cancellation. Some earlier fixture and migration inconsistencies were identified; failure attribution was incomplete. Limited component acceptance cannot establish whole-product correctness.

The separately governed V4 study is outside this protocol. Its original external evidence root is unavailable according to the current project handoff. Historical aggregates without their inspectable raw evidence are excluded from the present research results. No V4 configuration, lock, database, score, or automation was changed for this draft.

# Results status and interpretation limits

No main-study, perturbation, or iteration-case results have been collected for this manuscript. There is no estimated $Q$, update interaction, interval, significance value, equivalence conclusion, cost improvement, or measured adaptation curve to report. Implementation inspection and literature reading establish readiness information only.

When data become available, reporting must include all eight condition outcomes, paired contrasts, uncertainty, failure denominators, cost components, and the actual sampling population. A beneficial total update gain with negligible interaction would support useful additive effects. A negative interaction would motivate controlled mechanism investigation. A near-zero location contrast with sufficiently precise intervals could describe similar behavior under the tested deployment; a wide interval would remain inconclusive. None of these outcomes is selected in advance.

# Limitations, reproducibility, and research governance

The proposed semantics have not been implemented as a complete executable research system. The concrete condition catalogue, unique-verdict rule and clock are proposals for review, without formal parameter or interface approval. Conditions may omit relevant task differences, and partner selection changes which relationships are observed. Calibration among selected partners cannot automatically generalize to unselected pairs. Fixed software tasks and one initial history limit external validity. Provider-side randomness and model-version changes can weaken paired comparisons even when requested seeds match.

The evidence-quality defaults do not guarantee broad applicability. Feedback tests may be exploitable or overlap with final evaluation; human ratings may vary; model agreement can share errors. Authorization and privacy remain constraints on available evidence, with no permission inferred from the existence of an experience store. Initial snapshots and the formal task corpus require lawful, explicit provenance.

Before registration, the reproducibility package must identify the exact code and policy versions, all model/service versions, prompts and decoding settings, task sequences and suite access, initial-state sources, per-condition isolation, hardware and communication placement, resource limits, costs, randomization, exclusions, replication counts, and analysis programs. Existing project services and records should be reused. This requirement does not authorize a new scheduler or proof-of-completion system.

Formal execution depends on legitimate native-controller configuration, real persistent state and HTTP services, and the actual session files produced by the controller. Each claimed event must be locatable in the real event database. The writing work reported here sends no new product-native model input and runs no model experiment. Development writing tools are separate from permissions for the proposed experimental agents.

The supplied protocol remains a draft. Authors, affiliations, ORCID identifiers, final thresholds, budgets, tasks, replication counts, and a registration identifier are unfilled. Public submission and formal registration require subsequent decisions. The manuscript is intended for scientific and implementation review in its present evidence state.

## Use of AI tools

Codex CLI 0.162.0 using the Orca-configured `gpt-6.1-sol` model assisted source retrieval, synthesis, manuscript editing, and build/verification script preparation. No experimental model outcomes were generated for this manuscript. Human authorship, scientific verification, approval, and responsibility for a public submission remain to be established. This disclosure records actual tool use; it does not attest that currently unidentified authors have completed those checks.

# Conclusion

Astrocyte specifies a study of coupled external experience and collaboration-relationship updates under fixed model parameters. The eight-configuration design distinguishes update interaction, decision-location modification of interaction, and decision-location modification of total gain. Its evidential requirements include paired complete sequences, isolated adaptive states, protected evaluation, and comprehensive accounting. Current product evidence supports selected components; the proposed research effects remain to be measured.

# References

::: {#refs}
:::
