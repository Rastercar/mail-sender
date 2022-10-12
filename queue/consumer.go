package queue

func (s *Server) startConsumer() {
	for d := range s.deliveries {
		s.ConsumerFn(&d)
	}
}
