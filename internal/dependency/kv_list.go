package dependency

import (
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcat/dep"
	"github.com/pkg/errors"
)

var (
	// Ensure implements
	_ isDependency = (*KVListQuery)(nil)

	// KVListQueryRe is the regular expression to use.
	KVListQueryRe = regexp.MustCompile(`\A` + prefixRe + dcRe + `\z`)
)

func init() {
	gob.Register([]*KeyPair{})
}

// KeyPair is a simple Key-Value pair
type KeyPair struct {
	Path  string
	Key   string
	Value string

	// Lesser-used, but still valuable keys from api.KV
	CreateIndex uint64
	ModifyIndex uint64
	LockIndex   uint64
	Flags       uint64
	Session     string
}

// KVListQuery queries the KV store for a single key.
type KVListQuery struct {
	isConsul
	stopCh chan struct{}

	dc     string
	prefix string
	opts   QueryOptions
}

// NewKVListQuery parses a string into a dependency.
func NewKVListQuery(s string) (*KVListQuery, error) {
	if s != "" && !KVListQueryRe.MatchString(s) {
		return nil, fmt.Errorf("kv.list: invalid format: %q", s)
	}

	m := regexpMatch(KVListQueryRe, s)
	return &KVListQuery{
		stopCh: make(chan struct{}, 1),
		dc:     m["dc"],
		prefix: m["prefix"],
	}, nil
}

// Fetch queries the Consul API defined by the given client.
func (d *KVListQuery) Fetch(clients dep.Clients) (interface{}, *dep.ResponseMetadata, error) {
	select {
	case <-d.stopCh:
		return nil, nil, ErrStopped
	default:
	}

	opts := d.opts.Merge(&QueryOptions{
		Datacenter: d.dc,
	})

	//log.Printf("[TRACE] %s: GET %s", d, &url.URL{
	//	Path:     "/v1/kv/" + d.prefix,
	//	RawQuery: opts.String(),
	//})

	list, qm, err := clients.Consul().KV().List(d.prefix, opts.ToConsulOpts())
	if err != nil {
		return nil, nil, errors.Wrap(err, d.String())
	}

	//log.Printf("[TRACE] %s: returned %d pairs", d, len(list))

	pairs := make([]*KeyPair, 0, len(list))
	for _, pair := range list {
		key := strings.TrimPrefix(pair.Key, d.prefix)
		key = strings.TrimLeft(key, "/")

		pairs = append(pairs, &KeyPair{
			Path:        pair.Key,
			Key:         key,
			Value:       string(pair.Value),
			CreateIndex: pair.CreateIndex,
			ModifyIndex: pair.ModifyIndex,
			LockIndex:   pair.LockIndex,
			Flags:       pair.Flags,
			Session:     pair.Session,
		})
	}

	rm := &dep.ResponseMetadata{
		LastIndex:   qm.LastIndex,
		LastContact: qm.LastContact,
	}

	return pairs, rm, nil
}

// CanShare returns a boolean if this dependency is shareable.
func (d *KVListQuery) CanShare() bool {
	return true
}

// String returns the human-friendly version of this dependency.
func (d *KVListQuery) String() string {
	prefix := d.prefix
	if d.dc != "" {
		prefix = prefix + "@" + d.dc
	}
	return fmt.Sprintf("kv.list(%s)", prefix)
}

// Stop halts the dependency's fetch function.
func (d *KVListQuery) Stop() {
	close(d.stopCh)
}

func (d *KVListQuery) SetOptions(opts QueryOptions) {
	d.opts = opts
}
