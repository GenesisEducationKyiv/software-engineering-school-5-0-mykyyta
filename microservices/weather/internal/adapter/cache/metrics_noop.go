package cache

type NoopMetrics struct{}

func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}

func (n NoopMetrics) Register()                          {}
func (n NoopMetrics) RecordProviderHit(provider string)  {}
func (n NoopMetrics) RecordProviderMiss(provider string) {}
func (n NoopMetrics) RecordTotalHit()                    {}
func (n NoopMetrics) RecordTotalMiss()                   {}
