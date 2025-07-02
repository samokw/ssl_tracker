package ssl

import (
	"log/slog"
	"sync"
)

type CertService struct {
	pool    *WorkerPool
	results func(Result)
	started bool
	mu      sync.Mutex
}

func NewCertService() *CertService {
	return &CertService{
		pool: NewWorkerPool(20),
	}
}

func (cs *CertService) processResults() {
	for result := range cs.pool.GetResults() {
		cs.mu.Lock()
		handler := cs.results
		cs.mu.Unlock()

		if handler != nil {
			handler(result)
		} else {
			cs.defaultHandler(result)
		}
	}
}

func (cs *CertService) Start() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.started {
		return // Already started
	}

	cs.pool.Start()
	go cs.processResults()
	cs.started = true
}

func (cs *CertService) Stop() {
	cs.pool.Stop()
}

func (cs *CertService) CheckDomain(domain string, domainID, userID int) {
	task := Task{
		Domain:   domain,
		DomainID: domainID,
		UserID:   userID,
	}
	cs.pool.AddTask(task)
}

func (cs *CertService) SetResultHandler(handler func(Result)) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.results = handler
}

func (cs *CertService) defaultHandler(result Result) {
	if result.Error != nil {
		slog.Error("SSL check failed",
			"domain", result.Task.Domain,
			"error", result.Error,
		)
	} else {
		slog.Info("SSL check succeeded",
			"domain", result.Task.Domain,
			"expires_in_days", result.Certificate.TimeLeft,
		)
	}
}
