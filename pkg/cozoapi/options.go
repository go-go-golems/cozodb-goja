package cozoapi

type QueryOptionPolicy struct {
	DefaultTimeoutSec *float64
	MaxTimeoutSec     *float64
	MaxLimit          *int64
	MaxOffset         *int64
}

func NormalizeQueryOptions(opts *QueryOptions, policy QueryOptionPolicy) QueryOptions {
	if opts == nil {
		opts = &QueryOptions{}
	}
	normalized := opts.Clone()

	if normalized.Limit != nil {
		v := *normalized.Limit
		if v < 0 {
			v = 0
		}
		if policy.MaxLimit != nil && v > *policy.MaxLimit {
			v = *policy.MaxLimit
		}
		normalized.Limit = PtrInt64(v)
	}

	if normalized.Offset != nil {
		v := *normalized.Offset
		if v < 0 {
			v = 0
		}
		if policy.MaxOffset != nil && v > *policy.MaxOffset {
			v = *policy.MaxOffset
		}
		normalized.Offset = PtrInt64(v)
	}

	if normalized.TimeoutSec == nil && policy.DefaultTimeoutSec != nil {
		v := *policy.DefaultTimeoutSec
		normalized.TimeoutSec = PtrFloat64(v)
	}
	if normalized.TimeoutSec != nil {
		v := *normalized.TimeoutSec
		if v < 0 {
			v = 0
		}
		if policy.MaxTimeoutSec != nil && v > *policy.MaxTimeoutSec {
			v = *policy.MaxTimeoutSec
		}
		normalized.TimeoutSec = PtrFloat64(v)
	}

	return normalized
}
