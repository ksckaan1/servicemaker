package servicemaker

import "github.com/ksckaan1/logger"

func (s *ServiceMaker) closeAll() {
	s.closerMu.Lock()
	if s.closed {
		s.closerMu.Unlock()
		return
	}
	s.closed = true
	s.closerMu.Unlock()

	for _, closer := range s.closers {
		err := closer(s.gCtx)
		if err != nil {
			logger.Default.Error(s.gCtx, "error when closing component", "error", err)
		}
		s.closerWg.Done()
	}
}
