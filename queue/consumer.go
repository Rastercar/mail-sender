package queue

func (s *Server) startConsumer() {
	for d := range s.deliveries {
		go s.ConsumerFn(&d)
	}
}
