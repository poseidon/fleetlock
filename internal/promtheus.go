package fleetlock

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	papi "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PromClient is a wrapper around the Prometheus Client interface and implements the api
// This way, the PromClient can be instantiated with the configuration the Client needs, and
// the ability to use the methods the api has, like Query and so on.
type PromClient struct {
	papi   papi.Client
	api    v1.API
	Query  string
	Filter regexp.Regexp
}

// NewPromClient creates a new client to the Prometheus API.
// It returns an error on any problem.
func NewPromClient(conf papi.Config) (*PromClient, error) {
	promClient, err := papi.NewClient(conf)
	if err != nil {
		return nil, err
	}
	client := PromClient{papi: promClient, api: v1.NewAPI(promClient)}
	return &client, nil
}

// ActiveAlerts is a method of type PromClient, it returns a list of names of active alerts
// (e.g. pending or firing), filtered by the supplied regexp.
// filter by regexp means when the regex finds the alert-name; the alert is exluded from the
// block-list and will NOT block rebooting.
func (p *PromClient) ActiveAlerts() ([]string, error) {

	// get all alerts from prometheus
	value, _, err := p.api.Query(context.Background(), p.Query, time.Now())
	if err != nil {
		return nil, err
	}

	if value.Type() == model.ValVector {
		if vector, ok := value.(model.Vector); ok {
			activeAlertSet := make(map[string]bool)
			for _, sample := range vector {
				if alertName, isAlert := sample.Metric[model.AlertNameLabel]; isAlert && sample.Value != 0 {
					if p.Filter.String() == "" || !p.Filter.MatchString(string(alertName)) {
						activeAlertSet[string(alertName)] = true
					}
				}
			}

			var activeAlerts []string
			for activeAlert := range activeAlertSet {
				activeAlerts = append(activeAlerts, activeAlert)
			}
			sort.Strings(activeAlerts)

			return activeAlerts, nil
		}
	}

	return nil, fmt.Errorf("unexpected value type from promtheus: %v", value)
}
