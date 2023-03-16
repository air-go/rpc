package http

type Options struct {
	labels  []Label
	filters []Filter
}

type OptionFunc func(*Options)

func defaultOptions() *Options {
	return &Options{
		labels: DefaultLabels,
	}
}

func WithLabels(labels []Label) func(*Options) {
	return func(o *Options) { o.labels = append(o.labels, labels...) }
}

func WithFilters(filters []Filter) func(*Options) {
	return func(o *Options) { o.filters = filters }
}
