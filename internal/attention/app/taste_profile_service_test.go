package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 示例来自当前公开 Git 固定原档；规则与期望值是明确提交的人工配置参数。
func publicTasteCommand(material MaterialDetail, key string, expected int) UpdateTasteProfileCommand {
	return UpdateTasteProfileCommand{
		CommandMeta: meta(key, expected), Enabled: true,
		Rubrics:     []TasteRubric{{ID: "readme-scope", QuestionType: "truth", Criteria: "依据固定公开 README 核查所述能力的证据范围，缺少验证结果时保留 unknown。", Options: []TasteOption{}, Levels: []string{}}},
		Examples:    []TasteExample{{ID: "readme-example", RubricID: "readme-scope", MaterialID: material.Material.ID, Revision: material.Revisions[0].Revision, ContentDigest: material.Revisions[0].ContentDigest, Expected: "0"}},
		Delegations: []TasteDelegation{},
	}
}

func tasteReceiptCount(t *testing.T, db *sqlite.DB) int {
	t.Helper()
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM attention_receipts WHERE command='UpdateTasteProfile'").Scan(&count); err != nil {
		t.Fatal("count actual taste update receipts")
	}
	return count
}

func TestActualTasteProfileVersionsFixedSourcesAndRevocation(t *testing.T) {
	ctx := context.Background()
	core, db, objects, material, _ := readyOpportunityFromPublicGit(t)
	empty, err := core.GetTasteProfile(ctx, human)
	if err != nil || empty.Configured || empty.Profile != nil {
		t.Fatal("unconfigured Taste profile acquired invented values")
	}
	snapshot, err := core.ReadTasteProfile(ctx, human, "attention", 1)
	if err != nil || snapshot.Configured || snapshot.Current != nil || snapshot.Fixed != nil {
		t.Fatal("unconfigured Taste source invented a current or fixed version")
	}
	firstCommand := publicTasteCommand(material, "taste-first", 1)
	first, err := core.UpdateTasteProfile(ctx, human, firstCommand)
	if err != nil || first.Profile == nil || first.Profile.Version != 1 || first.Profile.ConfiguredBy != human.ID || len(first.Profile.Delegations) != 0 {
		t.Fatal("persist explicit human Taste configuration against actual fixed source")
	}
	ranking, err := core.GetRankingProfile(ctx, human)
	if err != nil || ranking.Configured || ranking.Profile != nil {
		t.Fatal("complete Taste configuration invented independent ranking weights")
	}
	secondCommand := firstCommand
	secondCommand.CommandMeta = meta("taste-rubric-next", first.Profile.Version)
	secondCommand.Rubrics = slices.Clone(firstCommand.Rubrics)
	secondCommand.Rubrics[0].Criteria = "依据固定公开 README 分开记录声明与实际验证，对未核实能力保留 unknown。"
	second, err := core.UpdateTasteProfile(ctx, human, secondCommand)
	if err != nil || second.Profile == nil || second.Profile.Version != 2 {
		t.Fatal("actual rubric update did not produce a new Taste version")
	}
	thirdCommand := secondCommand
	thirdCommand.CommandMeta = meta("taste-example-next", second.Profile.Version)
	thirdCommand.Examples = slices.Clone(secondCommand.Examples)
	thirdCommand.Examples[0].Expected = "1"
	third, err := core.UpdateTasteProfile(ctx, human, thirdCommand)
	if err != nil || third.Profile == nil || third.Profile.Version != 3 {
		t.Fatal("actual example update did not produce a new Taste version")
	}
	fourthCommand := thirdCommand
	fourthCommand.CommandMeta = meta("taste-disable-next", third.Profile.Version)
	fourthCommand.Enabled = false
	fourth, err := core.UpdateTasteProfile(ctx, human, fourthCommand)
	if err != nil || fourth.Profile == nil || fourth.Profile.Version != 4 || fourth.Profile.Enabled {
		t.Fatal("explicit human disable did not produce a new Taste version")
	}
	for _, original := range []TasteProfileDetail{first, second, third, fourth} {
		read, err := core.ReadTasteProfile(ctx, human, "attention", original.Profile.Version)
		if err != nil || !read.Configured || !sameAdmissionJSON(t, read.Fixed, original.Profile) || !sameAdmissionJSON(t, read.Current, fourth.Profile) {
			t.Fatal("actual Taste update overwrote immutable rubric or example history")
		}
	}
	_, err = core.UpdateTasteProfile(ctx, agent, publicTasteCommand(material, "taste-agent-denied", fourth.Profile.Version))
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = core.UpdateTasteProfile(ctx, human, publicTasteCommand(material, "taste-stale-denied", first.Profile.Version))
	errorCode(t, err, apierrors.VersionConflict)
	if countRows(t, db, "attention_taste_profile_revisions") != 4 || tasteReceiptCount(t, db) != 4 {
		t.Fatal("denied Taste mutation retained an extra revision or receipt")
	}
	core, db, _, _ = reopenAdmissionServices(t, db, objects)
	read, err := core.ReadTasteProfile(ctx, human, "attention", first.Profile.Version)
	if err != nil || !sameAdmissionJSON(t, read.Fixed, first.Profile) || !sameAdmissionJSON(t, read.Current, fourth.Profile) {
		t.Fatal("Taste fixed/current source changed after database reopen")
	}
	replay, err := core.UpdateTasteProfile(ctx, human, firstCommand)
	if err != nil || !sameAdmissionJSON(t, replay, first) || tasteReceiptCount(t, db) != 4 {
		t.Fatal("original complete Taste receipt changed or was repeated after reopen")
	}
	currentMaterial, err := core.GetMaterial(ctx, human, material.Material.ID)
	if err != nil {
		t.Fatal("load actual Taste example source before withdrawal")
	}
	if _, err = core.UpdateMaterial(ctx, human, material.Material.ID, UpdateMaterialCommand{CommandMeta: meta("taste-withdraw-source", currentMaterial.Material.Version), Lifecycle: "withdrawn"}); err != nil {
		t.Fatal("withdraw actual fixed Taste example source")
	}
	_, err = core.ReadTasteProfile(ctx, human, "attention", first.Profile.Version)
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = core.UpdateTasteProfile(ctx, human, publicTasteCommand(material, "taste-withdrawn-denied", fourth.Profile.Version))
	errorCode(t, err, apierrors.ScopeDenied)
	core, db, _, _ = reopenAdmissionServices(t, db, objects)
	_, err = core.ReadTasteProfile(ctx, human, "attention", fourth.Profile.Version)
	errorCode(t, err, apierrors.ScopeDenied)
	if countRows(t, db, "attention_taste_profile_revisions") != 4 || tasteReceiptCount(t, db) != 4 {
		t.Fatal("source revocation or reopen rewrote Taste history or produced another receipt")
	}
	assertNoAdmissionExecution(t, db)
}

func TestActualTasteProfileRollbackAndFixedDigestRejection(t *testing.T) {
	ctx := context.Background()
	core, db, _, material, _ := readyOpportunityFromPublicGit(t)
	command := publicTasteCommand(material, "taste-outbox-rollback", 1)
	events, receipts := countRows(t, db, "attention_outbox"), countRows(t, db, "attention_receipts")
	restore := refuseOutbox(t, db)
	_, err := core.UpdateTasteProfile(ctx, human, command)
	errorCode(t, err, apierrors.InternalError)
	restore()
	if countRows(t, db, "attention_taste_profiles") != 0 || countRows(t, db, "attention_taste_profile_revisions") != 0 || countRows(t, db, "attention_outbox") != events || countRows(t, db, "attention_receipts") != receipts {
		t.Fatal("actual Taste outbox failure retained state, immutable history or receipt")
	}
	first, err := core.UpdateTasteProfile(ctx, human, command)
	if err != nil || first.Profile == nil || first.Profile.Version != 1 {
		t.Fatal("original rolled-back Taste command could not subsequently succeed")
	}
	wrong := publicTasteCommand(material, "taste-digest-rejected", first.Profile.Version)
	wrong.Examples[0].ContentDigest = "explicit-invalid-reference"
	_, err = core.UpdateTasteProfile(ctx, human, wrong)
	errorCode(t, err, apierrors.VersionConflict)
	missing := publicTasteCommand(material, "taste-revision-rejected", first.Profile.Version)
	missing.Examples[0].Revision = material.Material.CurrentRevision + 1
	_, err = core.UpdateTasteProfile(ctx, human, missing)
	errorCode(t, err, apierrors.NotFound)
	if countRows(t, db, "attention_taste_profile_revisions") != 1 || tasteReceiptCount(t, db) != 1 || countRows(t, db, "attention_outbox") != events+1 || countRows(t, db, "attention_receipts") != receipts+1 {
		t.Fatal("rejected fixed Taste reference changed history, receipts or events")
	}
}

func TestActualTasteProfileRejectsUnknownQuestionType(t *testing.T) {
	core, db, _, material, _ := readyOpportunityFromPublicGit(t)
	command := publicTasteCommand(material, "taste-unknown-type", 1)
	command.Rubrics[0].QuestionType = "source-check"
	events, receipts := countRows(t, db, "attention_outbox"), countRows(t, db, "attention_receipts")
	_, err := core.UpdateTasteProfile(context.Background(), human, command)
	errorCode(t, err, apierrors.ValidationFailed)
	if countRows(t, db, "attention_taste_profiles") != 0 || countRows(t, db, "attention_taste_profile_revisions") != 0 || countRows(t, db, "attention_outbox") != events || countRows(t, db, "attention_receipts") != receipts {
		t.Fatal("unknown Taste question type retained a configuration or success receipt")
	}
}

func TestActualTasteProfileChoiceAndScoreKeepHumanConfiguration(t *testing.T) {
	for _, kind := range []string{"choice", "score"} {
		t.Run(kind, func(t *testing.T) {
			core, db, _, material, _ := readyOpportunityFromPublicGit(t)
			command := publicTasteCommand(material, "taste-typed-"+kind, 1)
			command.Rubrics[0].QuestionType = kind
			if kind == "choice" {
				command.Rubrics[0].Options = []TasteOption{{ID: "needs-evidence", Description: "继续核查证据"}, {ID: "review-scope", Description: "审阅已列出的来源范围"}}
				command.Examples[0].Expected = "needs-evidence"
			} else {
				command.Rubrics[0].Levels = []string{"补充来源", "审阅固定来源", "比较附加验证"}
				command.Examples[0].Expected = "1"
			}
			configured, err := core.UpdateTasteProfile(context.Background(), human, command)
			if err != nil || configured.Profile == nil || !sameAdmissionJSON(t, configured.Profile.Rubrics, command.Rubrics) || !sameAdmissionJSON(t, configured.Profile.Examples, command.Examples) || len(configured.Profile.Delegations) != 0 {
				t.Fatal("actual typed Taste storage changed explicit human configuration")
			}
			read, err := core.ReadTasteProfile(context.Background(), human, "attention", configured.Profile.Version)
			if err != nil || !sameAdmissionJSON(t, read.Fixed, configured.Profile) || !sameAdmissionJSON(t, read.Current, configured.Profile) || tasteReceiptCount(t, db) != 1 {
				t.Fatal("actual typed Taste reader changed the configured example or rubric")
			}
			assertNoAdmissionExecution(t, db)
		})
	}
}

func TestActualTasteProfileRejectsInvalidTypedExpectations(t *testing.T) {
	for _, expected := range []string{"unknown", "null", "false", "2", "{}"} {
		t.Run(expected, func(t *testing.T) {
			core, db, _, material, _ := readyOpportunityFromPublicGit(t)
			command := publicTasteCommand(material, "taste-invalid-expectation", 1)
			command.Examples[0].Expected = expected
			events, receipts := countRows(t, db, "attention_outbox"), countRows(t, db, "attention_receipts")
			_, err := core.UpdateTasteProfile(context.Background(), human, command)
			var reported *apierrors.ServiceError
			if errors.As(err, &reported) {
				t.Logf("typed expectation returned service error code=%s", reported.Code)
			}
			errorCode(t, err, apierrors.ValidationFailed)
			if countRows(t, db, "attention_taste_profiles") != 0 || countRows(t, db, "attention_taste_profile_revisions") != 0 || countRows(t, db, "attention_outbox") != events || countRows(t, db, "attention_receipts") != receipts {
				t.Fatal("invalid typed expectation retained a Taste configuration or success receipt")
			}
		})
	}
}
