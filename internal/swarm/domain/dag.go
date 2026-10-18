package domain

import "time"

// 依赖是永久任务事实，协作事件不进入此图。
func ValidateDAG(missionID string, items []WorkItem) error {
	degree := make(map[string]int, len(items))
	children := make(map[string][]string, len(items))
	for _, item := range items {
		if item.ID == "" || item.MissionID != missionID {
			return Failure("validation_failed", "DAG nodes require their actual mission identity")
		}
		if _, exists := degree[item.ID]; exists {
			return Failure("validation_failed", "DAG contains a duplicate node")
		}
		degree[item.ID] = 0
	}
	for _, item := range items {
		seen := map[string]bool{}
		for _, dependency := range item.Dependencies {
			if _, exists := degree[dependency]; !exists || dependency == item.ID || seen[dependency] {
				return Failure("validation_failed", "DAG dependency is missing, duplicated or self-referencing")
			}
			seen[dependency] = true
			degree[item.ID]++
			children[dependency] = append(children[dependency], item.ID)
		}
	}
	queue := []string{}
	for _, item := range items {
		if degree[item.ID] == 0 {
			queue = append(queue, item.ID)
		}
	}
	visited := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		visited++
		for _, child := range children[id] {
			degree[child]--
			if degree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}
	if visited != len(items) {
		return Failure("validation_failed", "DAG contains a dependency cycle")
	}
	return nil
}

func (m Mission) StopReason(now time.Time) string {
	l := m.Grant.Limits
	switch {
	case !now.Before(l.Deadline):
		return "deadline_reached"
	case l.MaxNoProgress > 0 && m.NoProgress >= l.MaxNoProgress:
		return "no_progress_round_limit"
	case l.NoProgressSeconds > 0 && !now.Before(m.LastProgressAt.Add(time.Duration(l.NoProgressSeconds)*time.Second)):
		return "no_progress_time_limit"
	case m.Grant.Budget.Calls != nil && m.Used.Calls != nil && *m.Used.Calls >= *m.Grant.Budget.Calls:
		return "call_budget_exhausted"
	case m.Grant.Budget.Tokens != nil && m.Used.Tokens != nil && *m.Used.Tokens >= *m.Grant.Budget.Tokens:
		return "token_budget_exhausted"
	case m.Grant.Budget.Money != nil && m.Used.Money != nil && m.Used.Currency == m.Grant.Budget.Currency && *m.Used.Money >= *m.Grant.Budget.Money:
		return "money_budget_exhausted"
	}
	return ""
}
