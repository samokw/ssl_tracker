package ssl

import "log/slog"

type CertService struct {
	pool    *WorkerPool
	results func(Result)
}

func NewCertService() *CertService {
	return &CertService{
		pool: NewWorkerPool(20),
	}
}

func (cs *CertService) processResults() {
	for result := range cs.pool.GetResults() {
		if cs.results != nil {
			cs.results(result)
		} else {
			cs.defaultHandler(result)
		}
	}
}

func (cs *CertService) Start() {
	cs.pool.Start()
	go cs.processResults()
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
