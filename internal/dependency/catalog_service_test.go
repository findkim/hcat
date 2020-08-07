package dependency

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCatalogServiceQuery(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		i    string
		exp  *CatalogServiceQuery
		err  bool
	}{
		{
			"empty",
			"",
			nil,
			true,
		},
		{
			"dc_only",
			"@dc1",
			nil,
			true,
		},
		{
			"near_only",
			"~near",
			nil,
			true,
		},
		{
			"tag_only",
			"tag.",
			nil,
			true,
		},
		{
			"name",
			"name",
			&CatalogServiceQuery{
				name: "name",
			},
			false,
		},
		{
			"name_dc",
			"name@dc1",
			&CatalogServiceQuery{
				dc:   "dc1",
				name: "name",
			},
			false,
		},
		{
			"name_dc_near",
			"name@dc1~near",
			&CatalogServiceQuery{
				dc:   "dc1",
				name: "name",
				near: "near",
			},
			false,
		},
		{
			"name_near",
			"name~near",
			&CatalogServiceQuery{
				name: "name",
				near: "near",
			},
			false,
		},
		{
			"tag_name",
			"tag.name",
			&CatalogServiceQuery{
				name: "name",
				tag:  "tag",
			},
			false,
		},
		{
			"tag_name_dc",
			"tag.name@dc",
			&CatalogServiceQuery{
				dc:   "dc",
				name: "name",
				tag:  "tag",
			},
			false,
		},
		{
			"tag_name_near",
			"tag.name~near",
			&CatalogServiceQuery{
				name: "name",
				near: "near",
				tag:  "tag",
			},
			false,
		},
		{
			"tag_name_dc_near",
			"tag.name@dc~near",
			&CatalogServiceQuery{
				dc:   "dc",
				name: "name",
				near: "near",
				tag:  "tag",
			},
			false,
		},
		{
			"tag_name_with_colon",
			"tag:value.name",
			&CatalogServiceQuery{
				name: "name",
				tag:  "tag:value",
			},
			false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			act, err := NewCatalogServiceQuery(tc.i)
			if (err != nil) != tc.err {
				t.Fatal(err)
			}

			if act != nil {
				act.stopCh = nil
			}

			assert.Equal(t, tc.exp, act)
		})
	}
}

func TestCatalogServiceQuery_Fetch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		i    string
		exp  []*CatalogService
	}{
		{
			"consul",
			"consul",
			[]*CatalogService{
				&CatalogService{
					Node:       testConsul.Config.NodeName,
					Address:    testConsul.Config.Bind,
					Datacenter: "dc1",
					TaggedAddresses: map[string]string{
						"lan": "127.0.0.1",
						"wan": "127.0.0.1",
					},
					NodeMeta: map[string]string{
						"consul-network-segment": "",
					},
					ServiceID:      "consul",
					ServiceName:    "consul",
					ServiceAddress: "",
					ServiceTags:    ServiceTags([]string{}),
					ServiceMeta:    map[string]string{},
					ServicePort:    testConsul.Config.Ports.Server,
				},
			},
		},
		{
			"service-meta",
			"service-meta",
			[]*CatalogService{
				&CatalogService{
					Node:       testConsul.Config.NodeName,
					Address:    testConsul.Config.Bind,
					Datacenter: "dc1",
					TaggedAddresses: map[string]string{
						"lan": "127.0.0.1",
						"wan": "127.0.0.1",
					},
					NodeMeta: map[string]string{
						"consul-network-segment": "",
					},
					ServiceID:      "service-meta",
					ServiceName:    "service-meta",
					ServiceAddress: "",
					ServiceTags:    ServiceTags([]string{"tag1"}),
					ServiceMeta:    map[string]string{"meta1": "value1"},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			d, err := NewCatalogServiceQuery(tc.i)
			if err != nil {
				t.Fatal(err)
			}

			act, _, err := d.Fetch(testClients)
			if err != nil {
				t.Fatal(err)
			}

			if act != nil {
				for _, s := range act.([]*CatalogService) {
					s.ID = ""
					s.TaggedAddresses = filterAddresses(s.TaggedAddresses)
				}
			}

			// delete any version data from ServiceMeta
			act_list := act.([]*CatalogService)
			for i := range act_list {
				act_list[i].ServiceMeta = filterVersionMeta(
					act_list[i].ServiceMeta)
			}

			assert.Equal(t, tc.exp, act)
		})
	}
}

func TestCatalogServiceQuery_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		i    string
		exp  string
	}{
		{
			"name",
			"name",
			"catalog.service(name)",
		},
		{
			"name_dc",
			"name@dc",
			"catalog.service(name@dc)",
		},
		{
			"name_near",
			"name~near",
			"catalog.service(name~near)",
		},
		{
			"name_dc_near",
			"name@dc~near",
			"catalog.service(name@dc~near)",
		},
		{
			"tag_name",
			"tag.name",
			"catalog.service(tag.name)",
		},
		{
			"tag_name_dc",
			"tag.name@dc",
			"catalog.service(tag.name@dc)",
		},
		{
			"tag_name_near",
			"tag.name~near",
			"catalog.service(tag.name~near)",
		},
		{
			"tag_name_dc_near",
			"tag.name@dc~near",
			"catalog.service(tag.name@dc~near)",
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			d, err := NewCatalogServiceQuery(tc.i)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.exp, d.String())
		})
	}
}
