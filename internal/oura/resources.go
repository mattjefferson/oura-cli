package oura

import "strings"

type QueryKind int

const (
	QueryNone QueryKind = iota
	QueryDate
	QueryDateTime
	QueryNextTokenOnly
)

type Resource struct {
	Key          string
	PathSegment  string
	SupportsList bool
	SupportsGet  bool
	Query        QueryKind
}

var resources = []Resource{
	{Key: "personal_info", PathSegment: "personal_info", SupportsGet: true, Query: QueryNone},
	{Key: "daily_activity", PathSegment: "daily_activity", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_cardiovascular_age", PathSegment: "daily_cardiovascular_age", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_readiness", PathSegment: "daily_readiness", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_resilience", PathSegment: "daily_resilience", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_sleep", PathSegment: "daily_sleep", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_spo2", PathSegment: "daily_spo2", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "daily_stress", PathSegment: "daily_stress", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "sleep", PathSegment: "sleep", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "sleep_time", PathSegment: "sleep_time", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "session", PathSegment: "session", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "workout", PathSegment: "workout", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "tag", PathSegment: "tag", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "enhanced_tag", PathSegment: "enhanced_tag", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "rest_mode_period", PathSegment: "rest_mode_period", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "vo2_max", PathSegment: "vO2_max", SupportsList: true, SupportsGet: true, Query: QueryDate},
	{Key: "ring_configuration", PathSegment: "ring_configuration", SupportsList: true, SupportsGet: true, Query: QueryNextTokenOnly},
	{Key: "heartrate", PathSegment: "heartrate", SupportsList: true, SupportsGet: false, Query: QueryDateTime},
}

var resourceIndex = func() map[string]Resource {
	idx := map[string]Resource{}
	for _, r := range resources {
		idx[strings.ToLower(r.Key)] = r
	}
	return idx
}()

func LookupResource(name string) (Resource, bool) {
	key := strings.ToLower(name)
	r, ok := resourceIndex[key]
	return r, ok
}

func Resources() []Resource {
	out := make([]Resource, 0, len(resources))
	for _, r := range resources {
		if _, ok := resourceIndex[strings.ToLower(r.Key)]; ok {
			out = append(out, r)
		}
	}
	return out
}
