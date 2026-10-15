package domain

import "math"

// Usage 与预留分开；没有实际证据的维度保持 nil。
type Usage struct {
	Calls    *int64 `json:"calls"`
	Tokens   *int64 `json:"tokens"`
	Money    *int64 `json:"money"`
	Currency string `json:"currency"`
}

func (u Usage) Validate() error {
	for _, v := range []*int64{u.Calls, u.Tokens, u.Money} {
		if v != nil && *v < 0 {
			return Failure("validation_failed", "actual usage cannot be negative")
		}
	}
	if u.Money != nil && u.Currency == "" || u.Money == nil && u.Currency != "" {
		return Failure("validation_failed", "actual monetary usage requires its currency")
	}
	return nil
}

func usageAdd(a, b *int64) (*int64, error) {
	if a == nil || b == nil {
		return nil, nil
	}
	if *a > math.MaxInt64-*b {
		return nil, Failure("budget_exhausted", "actual usage overflow")
	}
	n := *a + *b
	return &n, nil
}

func UsageFromReservation(r Resources) Usage {
	calls, tokens := r.Calls, r.Tokens
	return Usage{Calls: &calls, Tokens: &tokens, Money: r.Money, Currency: r.Currency}
}

func (u Usage) Add(v Usage) (Usage, error) {
	var out Usage
	if err := u.Validate(); err != nil {
		return out, err
	}
	if err := v.Validate(); err != nil {
		return out, err
	}
	var err error
	out.Calls, err = usageAdd(u.Calls, v.Calls)
	if err != nil {
		return out, err
	}
	out.Tokens, err = usageAdd(u.Tokens, v.Tokens)
	if err != nil {
		return out, err
	}
	if u.Money != nil && v.Money != nil {
		if u.Currency != v.Currency {
			return out, Failure("validation_failed", "actual monetary usage currencies differ")
		}
		out.Money, err = usageAdd(u.Money, v.Money)
		if err != nil {
			return out, err
		}
		out.Currency = u.Currency
	}
	return out, nil
}

func (u Usage) AddReserved(r Resources) (Usage, error) {
	if err := r.Validate(); err != nil {
		return Usage{}, err
	}
	return u.Add(UsageFromReservation(r))
}

func (u Usage) Within(b Budget) bool {
	return (b.Calls == nil || u.Calls != nil && *u.Calls <= *b.Calls) && (b.Tokens == nil || u.Tokens != nil && *u.Tokens <= *b.Tokens) && (b.Money == nil || u.Money != nil && u.Currency == b.Currency && *u.Money <= *b.Money)
}

func (u Usage) WithinReservation(r Resources, b Budget) bool {
	return (b.Calls == nil || u.Calls != nil && *u.Calls <= r.Calls) && (b.Tokens == nil || u.Tokens != nil && *u.Tokens <= r.Tokens) && (b.Money == nil || u.Money != nil && r.Money != nil && u.Currency == r.Currency && *u.Money <= *r.Money)
}

func (m *Mission) RecordUsage(u Usage) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if !u.Within(m.Grant.Budget) {
		return Failure("budget_exhausted", "bounded actual usage evidence is missing or exceeds approval")
	}
	if m.SettledOperations == 0 {
		m.Used = u
	} else {
		sum, err := m.Used.Add(u)
		if err != nil {
			return err
		}
		if !sum.Within(m.Grant.Budget) {
			return Failure("budget_exhausted", "cumulative actual usage exceeds approval")
		}
		m.Used = sum
	}
	m.SettledOperations++
	return nil
}
