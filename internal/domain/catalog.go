package domain

import "time"

const (
	maxNameLen        = 120
	maxDescriptionLen = 500
)

type Service struct {
	ID          ServiceID
	Name        string
	Description string
	Criticality Criticality
	Source      DataSource
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s Service) Validate() error {
	if err := s.ID.Validate(); err != nil {
		return err
	}
	if err := requireName("name", s.Name, maxNameLen); err != nil {
		return err
	}
	if len(s.Description) > maxDescriptionLen {
		return ValidationError{Field: "description", Message: "is too long"}
	}
	if err := s.Criticality.Validate(); err != nil {
		return err
	}
	if err := s.Source.Validate(); err != nil {
		return err
	}
	if err := requireTime("created_at", s.CreatedAt); err != nil {
		return err
	}
	if err := requireTime("updated_at", s.UpdatedAt); err != nil {
		return err
	}
	if s.UpdatedAt.Before(s.CreatedAt) {
		return ValidationError{Field: "updated_at", Message: "must not precede created_at"}
	}
	return nil
}

func (s Service) Normalized() Service {
	s.CreatedAt = UTC(s.CreatedAt)
	s.UpdatedAt = UTC(s.UpdatedAt)
	return s
}

type Dependency struct {
	From      ServiceID
	To        ServiceID
	Kind      DependencyKind
	CreatedAt time.Time
}

func (d Dependency) Key() string {
	return d.From.String() + "->" + d.To.String()
}

func (d Dependency) Validate() error {
	if err := d.From.Validate(); err != nil {
		return err
	}
	if err := d.To.Validate(); err != nil {
		return err
	}
	if d.From == d.To {
		return ValidationError{Field: "to", Message: "service cannot depend on itself"}
	}
	if err := d.Kind.Validate(); err != nil {
		return err
	}
	if err := requireTime("created_at", d.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (d Dependency) Normalized() Dependency {
	d.CreatedAt = UTC(d.CreatedAt)
	return d
}
