package observability

import "github.com/prometheus/client_golang/prometheus"

type BusinessMetrics struct {
	usersRegistered prometheus.Counter
	logins          prometheus.Counter
	articlesCreated prometheus.Counter
	commentsCreated prometheus.Counter
}

func NewBusinessMetrics(registry *prometheus.Registry) *BusinessMetrics {
	metrics := &BusinessMetrics{
		usersRegistered: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "users_registered_total",
			Help:      "Total number of user accounts successfully created.",
		}),
		logins: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "logins_total",
			Help:      "Total number of successful logins with a generated access token.",
		}),
		articlesCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "articles_created_total",
			Help:      "Total number of articles successfully created.",
		}),
		commentsCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "comments_created_total",
			Help:      "Total number of comments successfully created.",
		}),
	}

	registry.MustRegister(metrics.usersRegistered, metrics.logins, metrics.articlesCreated, metrics.commentsCreated)
	return metrics
}

func (m *BusinessMetrics) UserRegistered() {
	m.usersRegistered.Inc()
}

func (m *BusinessMetrics) UserLoggedIn() {
	m.logins.Inc()
}

func (m *BusinessMetrics) ArticleCreated() {
	m.articlesCreated.Inc()
}

func (m *BusinessMetrics) CommentCreated() {
	m.commentsCreated.Inc()
}
