package routedinformer

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers/internalinterfaces"
)

// LabelSelectorTweak returns a TweakListOptionsFunc that sets ListOptions.LabelSelector
// so the informer only lists/watches resources matching the given labels (server-side).
// Use at setup with WithTweakListOptions(LabelSelectorTweak(map[string]interface{}{"managed-by": managerTag})).
// Values are converted to strings; multiple key-value pairs are ANDed (comma-separated selector).
func LabelSelectorTweak(labels map[string]interface{}) internalinterfaces.TweakListOptionsFunc {
	if len(labels) == 0 {
		return func(*metav1.ListOptions) {}
	}
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, k+"="+fmt.Sprint(v))
	}
	selector := strings.Join(parts, ",")
	return func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector
	}
}
