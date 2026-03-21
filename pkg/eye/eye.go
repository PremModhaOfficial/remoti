package eye

import (
	"context"
	"errors"
	"fmt"
	"log"

	"remoti/pkg/client"
	"remoti/pkg/eye/sense"
	"remoti/pkg/eye/sense/atspi"
	"remoti/pkg/eye/sense/niri"
)

// Eye combines sensing (Finder) and acting (client.Client) into a unified interface.
type Eye struct {
	act    *client.Client
	finder *sense.Finder
	opts   Options
	atspi  *atspi.Source // held for cleanup
}

// New creates an Eye from an existing actuator client and finder.
func New(act *client.Client, finder *sense.Finder, opts ...Option) *Eye {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return &Eye{act: act, finder: finder, opts: o}
}

// Connect creates an Eye with a new client connection and auto-detected sources.
func Connect(address string, opts ...Option) (*Eye, error) {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	if address != "" {
		o.ServerAddress = address
	}

	// Connect actuator
	act, err := client.Connect(o.ServerAddress, client.WithNetwork(o.Network))
	if err != nil {
		return nil, fmt.Errorf("eye: connect actuator: %w", err)
	}

	// Auto-detect sensing sources
	var sources []sense.Source

	// Try Niri (also used as WindowLocator for AT-SPI coordinate correction)
	niriSrc := niri.New()
	sources = append(sources, niriSrc)

	// Try AT-SPI, passing Niri as locator for Wayland coordinate fix
	var atspiSrc *atspi.Source
	atspiSrc, err = atspi.New(atspi.WithLocator(niriSrc))
	if err != nil {
		log.Printf("eye: AT-SPI unavailable: %v", err)
	} else {
		sources = append(sources, atspiSrc)
	}

	finder := sense.NewFinder(sources...)

	return &Eye{
		act:    act,
		finder: finder,
		opts:   o,
		atspi:  atspiSrc,
	}, nil
}

// Find locates UI elements matching the query, returning Elements with action methods.
func (e *Eye) Find(ctx context.Context, q sense.Query) ([]*Element, error) {
	matches, err := e.finder.Find(ctx, q)
	if err != nil {
		return nil, err
	}
	elements := make([]*Element, len(matches))
	for i, m := range matches {
		elements[i] = &Element{match: m, client: e}
	}
	return elements, nil
}

// FindOne returns the first matching element or ErrNotFound.
func (e *Eye) FindOne(ctx context.Context, q sense.Query) (*Element, error) {
	q.MaxResults = 1
	elems, err := e.Find(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(elems) == 0 {
		return nil, sense.ErrNotFound
	}
	return elems[0], nil
}

// FindAndClick finds an element by name and clicks it.
func (e *Eye) FindAndClick(ctx context.Context, name string) error {
	elem, err := e.FindOne(ctx, sense.Query{Name: name})
	if err != nil {
		return err
	}
	return elem.Click()
}

// FindAndType finds an element, clicks it, then types text.
func (e *Eye) FindAndType(ctx context.Context, name string, text string) error {
	elem, err := e.FindOne(ctx, sense.Query{Name: name})
	if err != nil {
		return err
	}
	return elem.Type(text)
}

// Act returns the underlying actuator client for raw commands.
func (e *Eye) Act() *client.Client { return e.act }

// Finder returns the underlying Finder for advanced queries.
func (e *Eye) Finder() *sense.Finder { return e.finder }

// Close closes both the finder resources and the actuator connection.
func (e *Eye) Close() error {
	var errs []error
	if e.atspi != nil {
		if err := e.atspi.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := e.act.Close(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
