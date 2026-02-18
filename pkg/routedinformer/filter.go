package routedinformer

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Filter returns true if the object should be passed to the handler.
// Stacked filters are applied in order; all must pass for the event to be routed.
type Filter func(obj interface{}) bool

// LabelFiltering returns a Filter that passes only objects whose labels match
// all given key-value pairs. Values in required are converted to strings for comparison.
// Supports multiple values per key as comma-separated string for OR semantics (e.g. "env": "dev,staging").
func LabelFiltering(required map[string]interface{}) Filter {
	if len(required) == 0 {
		return func(interface{}) bool { return true }
	}
	return func(obj interface{}) bool {
		meta, ok := obj.(metav1.Object)
		if !ok {
			return false
		}
		labels := meta.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		for k, v := range required {
			objVal := fmt.Sprint(v)
			labelVal, has := labels[k]
			if !has {
				return false
			}
			// Optional: allow "val1,val2" in required to mean label must be one of these
			if strings.Contains(objVal, ",") {
				allowed := strings.Split(objVal, ",")
				for i := range allowed {
					allowed[i] = strings.TrimSpace(allowed[i])
				}
				found := false
				for _, a := range allowed {
					if labelVal == a {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			} else if labelVal != objVal {
				return false
			}
		}
		return true
	}
}
