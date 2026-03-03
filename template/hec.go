package template

// TransformHECFields renames Splunk internal fields to HEC format:
// _raw -> event, _time -> time.
func TransformHECFields(event map[string]string) {
	if v, ok := event["_raw"]; ok {
		event["event"] = v
		delete(event, "_raw")
	}
	if v, ok := event["_time"]; ok {
		event["time"] = v
		delete(event, "_time")
	}
}
