package telemetry

type Config interface {
	GetOtelServerURL() string
	GetServiceName() string
	GetCommitTag() string
	GetEnvironment() string
	GetSentryDsn() string
}

type Params struct {
	OtelServerURL string `json:"otelServerUrl" yaml:"otelServerUrl"`
	ServiceName   string `json:"serviceName" yaml:"serviceName"`
	SentryDsn     string `json:"sentryDsn" yaml:"sentryDsn"`
	Environment   string `json:"environment" yaml:"environment"`
	CommitTag     string `json:"commitTag" yaml:"commitTag"`
}

func (p *Params) GetOtelServerURL() string {
	return p.OtelServerURL
}

func (p *Params) GetServiceName() string {
	return p.ServiceName
}

func (p *Params) GetCommitTag() string {
	return p.CommitTag
}

func (p *Params) GetEnvironment() string {
	return p.Environment
}

func (p *Params) GetSentryDsn() string {
	return p.SentryDsn
}
