package hcat

import "github.com/hashicorp/hcat/dep"

// Resolver is responsible rendering Templates and invoking Commands.
// Empty but reserving the space for future use.
type Resolver struct{}

// ResolveEvent captures the whether the template dependencies have all been
// resolved and rendered in memory.
type ResolveEvent struct {
	// Complete is true if all dependencies have values and the template
	// is is fully rendered.
	Complete bool

	// Contents is the rendered contents from the template.
	// Only returned when Complete is true.
	Contents []byte

	// for testing, need way to test for missing dependencies case
	missing bool
}

// Basic constructor, here for consistency and future flexibility.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Watcherer is the subset of the Watcher's API that the resolver needs.
// The interface is used to make the used/required API explicit.
type Watcherer interface {
	Recaller
	Add(dep.Dependency) bool
	Changed(tmplID string) bool
	Register(tmplID string, deps ...dep.Dependency)
}

// Templater the interface the Template provides.
// The interface is used to make the used/required API explicit.
type Templater interface {
	Execute(Recaller) (*ExecuteResult, error)
	ID() string
}

// Run the template Execute once. You should repeat calling this until
// output returns Complete as true. It uses the watcher for dependency
// lookup state. The content will be updated each pass until complete.
func (r *Resolver) Run(tmpl Templater, w Watcherer) (ResolveEvent, error) {
	// Check if this dependency has any dependencies that have been change and
	// if not, don't waste time re-rendering it.
	if !w.Changed(tmpl.ID()) {
		return ResolveEvent{}, nil
	}

	// Attempt to render the template, returning any missing dependencies and
	// the rendered contents. If there are any missing dependencies, the
	// contents cannot be rendered or trusted!
	result, err := tmpl.Execute(w)
	if err != nil {
		return ResolveEvent{}, err
	}

	// register all dependencies used
	if l := result.Used.Len(); l > 0 {
		w.Register(tmpl.ID(), result.Used.List()...)
	}

	// add missing dependencies to watcher
	if l := result.Missing.Len(); l > 0 {
		for _, d := range result.Missing.List() {
			w.Add(d)
		}
		// If the template is missing data for some dependencies then we are
		// not ready to render and need to move on to the next one.
		return ResolveEvent{missing: true}, nil
	}

	return ResolveEvent{Complete: true, Contents: result.Output}, nil
}
