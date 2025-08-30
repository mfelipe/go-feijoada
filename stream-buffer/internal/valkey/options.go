package valkey

type Option func(*stream)

func WithClient(c client) Option {
	return func(s *stream) {
		s.cli = c
	}
}
