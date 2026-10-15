package judgment

func applyRule(q Question, evidence []Evidence, result *JudgmentResult) bool {
	if q.Rule == "" {
		return false
	}
	present := map[EvidenceRef]bool{}
	for _, entry := range evidence {
		present[entry.Reference] = entry.Text != ""
	}
	switch q.Rule {
	case "all_evidence_present":
		for _, found := range present {
			if !found {
				return false
			}
		}
		value := 1.0
		result.Value = &value
	case "evidence_coverage":
		count := 0
		for _, found := range present {
			if found {
				count++
			}
		}
		value := float64(count) * float64(len(q.Levels)-1) / float64(len(evidence))
		result.Value = &value
	case "single_supported_option":
		var selected *string
		for _, option := range q.Options {
			supported := true
			for _, reference := range option.RequiredEvidence {
				if !present[reference] {
					supported = false
				}
			}
			if supported {
				if selected != nil {
					return false
				}
				id := option.ID
				selected = &id
			}
		}
		if selected == nil {
			return false
		}
		result.Selected = selected
	default:
		return false
	}
	result.Decision, result.ReasonCode = Recommend, "deterministic_"+q.Rule
	return true
}
