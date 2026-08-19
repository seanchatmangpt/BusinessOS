# BusinessOS LinkedIn RevOps Strategy

## Standing and claim ceiling

BusinessOS already exposes a client CRM with deals pipeline, contact management, interaction history, projects, documents, calendar, agent orchestration, and an extensible module model. That makes it a strong **REVOPS SYSTEM-OF-ENGAGEMENT / PARTIAL_ALIVE** candidate for LinkedIn-originated demand. The repository does **not** currently prove direct LinkedIn API publication or lead-ingestion behavior, so this strategy does not claim it.

## Role in the Chatman revenue system

BusinessOS should own the account/customer operating context after a LinkedIn signal becomes commercially relevant:

```text
LinkedIn content / signal
  -> readiness assessment
  -> admitted Person + Account
  -> CRM interaction / deal
  -> Problem + Capability gap
  -> POV / Experiment
  -> Outcome + revenue
```

`linkedin-public-canon` remains the publication source. The RevOps acquisition surface produces the assessment/handoff envelope. BusinessOS becomes the longitudinal workspace where agents and humans manage the account, deal, evidence, meetings, documents, tasks, and proof program.

## Canonical objects

Map LinkedIn-originated demand into explicit business objects rather than a generic contact note:

- `Person` and `Account`
- `Campaign` and `ContentAsset`
- `Signal` and `Interaction`
- `Assessment`
- `Problem` and affected `Capability`
- `Opportunity`
- `Experiment` / POV
- `Outcome`
- `Receipt`

Reuse existing client/deal/contact primitives where they are semantically adequate. Add campaign or evidence fields as extensions rather than duplicating CRM entities.

## August 31 campaign contract

For `10k_august_2026`, BusinessOS should be able to ingest a qualified handoff carrying source, campaign, asset, persona, assessment result, identified synchronization constraint, and consent. It should then create or reconcile the Person/Account, attach the interaction to the correct deal/account history, and open a bounded follow-up workflow.

The desired follow-up is not a generic demo. It is a proposed **Enterprise Manufacturing Proof** with one exact workflow, present synchronization cost, authority path, and falsifiable success condition.

## Challenger Sale logic

BusinessOS should preserve the teaching diagnosis that created the opportunity:

```text
Teach: coding is not necessarily the bottleneck
Tailor: identify the account-specific synchronization tax
Take Control: propose one measurable proof
```

The CRM must retain the prospect's identified problem and consequence rather than reducing the opportunity to stage + amount. This enables agents to prepare account-specific follow-up without losing the commercial thesis.

## Qualification states

Recommended machine-readable stages:

```text
LEAD
MQL
SQL
POV_PROPOSED
POV_ACTIVE
POV_PROVEN
CUSTOMER
EXPANSION
DISQUALIFIED
```

Transitions require evidence. In particular:

```text
MQL = ICP_fit AND identifiable_problem
SQL = exact_subject AND pain AND consequence AND authority_path AND falsifiable_outcome
```

A LinkedIn reaction, profile view, or form submit alone cannot produce `SQL`.

## SELECT / CONSTRUCT / DO

BusinessOS agents may SELECT an admitted next action and CONSTRUCT outreach, meeting, deal, or POV intents. Sending external communications, creating external calendar events, mutating external CRM/vendor records, or initiating consequential customer work is DO and must cross an authorized integration boundary with receipts.

No LinkedIn automation authority is implied by CRM capability.

## Revenue attribution

Persist first-touch and last-touch campaign identity while also recording subsequent interactions. Minimum fields:

```yaml
source: linkedin
campaign: 10k_august_2026
content_asset_id: <id>
first_touch: <timestamp>
last_touch: <timestamp>
persona: <persona>
assessment_id: <id>
problem_id: <id>
opportunity_id: <id>
pov_id: <id>
revenue: <amount-if-realized>
```

Revenue reporting should work backward from realized outcomes to the originating evidence graph, not treat impressions as revenue standing.

## Next admitted increments

1. Define a versioned LinkedIn/RevOps handoff schema for Person, Account, campaign, asset, assessment, and consent.
2. Implement idempotent Person/Account reconciliation and duplicate refusal rules.
3. Extend interaction/deal history with campaign and evidence provenance.
4. Add explicit MQL/SQL/POV transition predicates and refusal reasons.
5. Add an Enterprise Manufacturing Proof workflow with tasks, meetings, documents, evidence, and outcome fields.
6. Add revenue-attribution projections across content asset -> account -> opportunity -> POV -> customer.

## Falsifiers

The BusinessOS LinkedIn role is narrowed if its existing CRM cannot preserve campaign provenance, if account identity cannot be reconciled deterministically, or if external mutations occur without an auditable authority boundary. CRM presence alone is not proof of a functioning LinkedIn acquisition pipeline.
