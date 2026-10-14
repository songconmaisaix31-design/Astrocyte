import importlib.util
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[3]
SPEC = importlib.util.spec_from_file_location("experiment_tests", ROOT / "tests/validation/experiment/test_experiment.py")
module = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(module)

# 原有纯函数断言不依赖 Fixtures 的外部来源替换或私人输出目录。
PURE = (
    "test_four_same_questions_cover_dimensions_natural_facts",
    "test_empty_day_not_floor_activity",
    "test_fact_aliases_hallucinations_omissions_not_ids",
    "test_fact_duplicates_invalid_not_silently_normalized",
    "test_b_whole_oversize_skip_continue_chronology",
    "test_b_full_fit_all_loaded",
    "test_user_judge_independence_and_anonymous_ab",
    "test_user_numeric_rubric_required_independent_of_preference",
    "test_local_route_rejects_cloud_hostname_credentials",
    "test_complete_capture_rejects_partial_windows_late_unknown",
    "test_summary_missing_not_shrunken_and_judge_not_authority",
    "test_summary_success_line_five_of_seven_user_wins_fact_noninferiority",
    "test_judge_requires_independent_numeric_ok_receipt",
)
suite = unittest.TestSuite()
for name in PURE:
    case = module.Fixtures(name)
    suite.addTest(unittest.FunctionTestCase(getattr(case, name), description=name))
suite.addTests(unittest.defaultTestLoader.loadTestsFromTestCase(module.LosslessCandidateTests))
for name in (
    "test_only_exact_pre_turn_official_warnings_are_classified",
    "test_reason_template_accepts_structured_explanation_without_rewriting",
    "test_native_protocol_binds_complete_split_frames",
    "test_native_lock_rejects_stale_transport_and_unreviewed_draft",
    "test_split_complete_unicode_packet_empty_second_and_overflow_rejection",
    "test_context_file_complete_roundtrip_and_shared_total_bytes",
    "test_native_structured_input_rejection_is_numeric_not_unknown_network",
    "test_layered_full_state_day_boundary_refs_and_no_weak_weight_rounding",
    "test_b_shared_byte_ceiling_preserves_whole_message",
    "test_native_private_egress_and_exact_tokenizer_fail_closed",
):
    suite.addTest(module.V4Contracts(name))

if __name__ == "__main__":
    result = unittest.TextTestRunner(verbosity=2).run(suite)
    raise SystemExit(0 if result.wasSuccessful() else 1)
