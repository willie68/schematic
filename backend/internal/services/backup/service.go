package backup

import (
	"log/slog"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/schematics2/backend/internal/logging"
)

type Option struct {
}

type service struct {
	timeout time.Duration
	log     *slog.Logger
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func New(inj do.Injector, option ...func(*service)) (*service, error) {
	s := &service{
		log: logging.New("backup"),
	}
	for _, opt := range option {
		opt(s)
	}
	return s, nil
}

func WithTimeout(timeout time.Duration) func(*service) {
	return func(s *service) {
		s.log.Info("set backup timeout", "timeout", timeout)
		s.timeout = timeout
	}
}

func (s *service) Start() error {
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go func() {
		defer close(s.doneCh)
		ticker := time.NewTicker(s.timeout)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				_ = s.Backup()
			}
		}
	}()

	return nil
}

func (s *service) Stop() error {
	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
	}
	return nil
}

func (s *service) Backup() error {
	return nil
}
