package events

import (
	"context"
	"sync"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
)

// Bus is the application port. LocalMemoryBus exercises async semantics
// without EventBridge/SQS. AWS adapters will implement the same port later.
type Bus interface {
	Publish(ctx context.Context, e Envelope) error
}

type ProcessResult struct {
	Duplicate bool
	Dropped   bool
	Reason    string
	Pending   bool
}

type Analyzer interface {
	AnalyzeRelease(ctx context.Context, id domain.ReleaseID) error
}

type Processor struct {
	store     repository.Store
	maxRetry  int
	analyzer  Analyzer
	mu        sync.Mutex
	attempts  map[domain.EventID]int
	dlq       []Envelope
	processed map[domain.EventID]struct{}
}

func NewProcessor(store repository.Store, maxRetry int) *Processor {
	if maxRetry <= 0 {
		maxRetry = 3
	}
	return &Processor{
		store: store, maxRetry: maxRetry,
		attempts:  map[domain.EventID]int{},
		processed: map[domain.EventID]struct{}{},
	}
}

func (p *Processor) SetAnalyzer(a Analyzer) { p.analyzer = a }

// TriggersAnalysis is false for policy.evaluated so an evaluation event
// cannot re-enter the analysis worker and loop.
func TriggersAnalysis(t domain.EventType) bool {
	return t != domain.EventTypePolicyEvaluated
}

func (p *Processor) Handle(ctx context.Context, e Envelope, now time.Time) (ProcessResult, error) {
	norm := Normalize(e, now)
	if norm.Drop {
		p.deadLetter(norm.Envelope)
		return ProcessResult{Dropped: true, Reason: norm.Reason}, nil
	}
	e = norm.Envelope
	dom := ToDomain(e)
	err := p.store.CreateEvent(ctx, dom)
	if domain.IsAlreadyExists(err) {
		if ev, ok := EvidenceFromEnvelope(e); ok {
			if putErr := p.store.PutOperationalEvidence(ctx, ev); putErr != nil && !domain.IsAlreadyExists(putErr) {
				return ProcessResult{}, putErr
			}
		}
		if p.analyzer != nil && e.ReleaseID != nil && TriggersAnalysis(e.EventType) {
			if anErr := p.analyzer.AnalyzeRelease(ctx, *e.ReleaseID); anErr != nil && !domain.IsNotFound(anErr) {
				return ProcessResult{}, anErr
			}
		}
		return ProcessResult{Duplicate: true, Reason: "duplicate event_id"}, nil
	}
	if err != nil {
		p.mu.Lock()
		p.attempts[e.EventID]++
		n := p.attempts[e.EventID]
		p.mu.Unlock()
		if n >= p.maxRetry {
			p.deadLetter(e)
			return ProcessResult{Dropped: true, Reason: "poison after bounded retries"}, err
		}
		return ProcessResult{}, err
	}
	if e.ReleaseID != nil {
		if _, err := p.store.GetRelease(ctx, *e.ReleaseID); domain.IsNotFound(err) {
			return ProcessResult{Pending: true, Reason: "release not found yet; out of order tolerated"}, nil
		}
	}
	if ev, ok := EvidenceFromEnvelope(e); ok {
		if err := p.store.PutOperationalEvidence(ctx, ev); err != nil && !domain.IsAlreadyExists(err) {
			return ProcessResult{}, err
		}
	}
	if !TriggersAnalysis(e.EventType) {
		return ProcessResult{Reason: "timeline only; analysis not dispatched"}, nil
	}
	if p.analyzer != nil && e.ReleaseID != nil {
		if err := p.analyzer.AnalyzeRelease(ctx, *e.ReleaseID); err != nil && !domain.IsNotFound(err) {
			return ProcessResult{}, err
		}
	}
	p.mu.Lock()
	p.processed[e.EventID] = struct{}{}
	p.mu.Unlock()
	return ProcessResult{}, nil
}

func (p *Processor) DLQ() []Envelope {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Envelope, len(p.dlq))
	copy(out, p.dlq)
	return out
}

func (p *Processor) deadLetter(e Envelope) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dlq = append(p.dlq, e)
}

// MemoryBus is a local in-process queue used until EventBridge/SQS exist.
type MemoryBus struct {
	proc *Processor
	now  func() time.Time
}

func NewMemoryBus(proc *Processor) *MemoryBus {
	return &MemoryBus{proc: proc, now: time.Now}
}

func (b *MemoryBus) Publish(ctx context.Context, e Envelope) error {
	_, err := b.proc.Handle(ctx, e, b.now().UTC())
	return err
}
