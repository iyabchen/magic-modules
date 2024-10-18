package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/GoogleCloudPlatform/magic-modules/tools/diff-processor/detector"
	"github.com/GoogleCloudPlatform/magic-modules/tools/diff-processor/diff"
	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDetectMissingDocs(t *testing.T) {
	cases := []struct {
		name             string
		oldResourceMap   map[string]*schema.Resource
		newResourceMap   map[string]*schema.Resource
		oldDataSourceMap map[string]*schema.Resource
		newDataSourceMap map[string]*schema.Resource
		want             MissingDocsSummary
	}{
		{
			name: "no new fields",
			oldResourceMap: map[string]*schema.Resource{
				"google_x": {
					Schema: map[string]*schema.Schema{
						"field-a": {Description: "beep", Optional: true},
						"field-b": {Description: "beep", Optional: true},
					},
				},
			},
			newResourceMap: map[string]*schema.Resource{
				"google_x": {
					Schema: map[string]*schema.Schema{
						"field-a": {Description: "beep", Optional: true},
						"field-b": {Description: "beep", Optional: true},
					},
				},
			},
			want: MissingDocsSummary{
				Resource:   []detector.MissingDocDetails{},
				DataSource: []detector.MissingDocDetails{},
			},
		},
		{
			name: "doc not exist - multiple new fields missing doc",
			newResourceMap: map[string]*schema.Resource{
				"google_x": {
					Schema: map[string]*schema.Schema{
						"field-a": {Description: "beep"},
						"field-b": {Description: "beep"},
					},
				},
			},
			oldResourceMap: map[string]*schema.Resource{},
			newDataSourceMap: map[string]*schema.Resource{
				"google_data_y": {
					Schema: map[string]*schema.Schema{
						"field-a": {Description: "beep"},
					},
				},
			},
			oldDataSourceMap: map[string]*schema.Resource{},
			want: MissingDocsSummary{
				Resource: []detector.MissingDocDetails{
					{
						Name:     "google_x",
						FilePath: "/website/docs/r/x.html.markdown",
						Fields: []detector.MissingDocField{
							{
								Field:   "field-a",
								Section: "Arguments Reference",
							},
							{
								Field:   "field-b",
								Section: "Arguments Reference",
							},
						},
					},
				},
				DataSource: []detector.MissingDocDetails{
					{
						Name:     "google_data_y",
						FilePath: "/website/docs/d/data_y.html.markdown",
						Fields: []detector.MissingDocField{
							{
								Field: "field-a",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			o := detectMissingDocsOptions{
				computeSchemaDiff: func() diff.SchemaDiff {
					return diff.ComputeSchemaDiff(tc.oldResourceMap, tc.newResourceMap)
				},
				computeDatasourceSchemaDiff: func() diff.SchemaDiff {
					return diff.ComputeSchemaDiff(tc.oldDataSourceMap, tc.newDataSourceMap)
				},
				stdout: &buf,
			}

			err := o.run([]string{t.TempDir()})
			if err != nil {
				t.Fatalf("Error running command: %s", err)
			}

			out := make([]byte, buf.Len())
			buf.Read(out)

			var got MissingDocsSummary
			if err = json.Unmarshal(out, &got); err != nil {
				t.Fatalf("Failed to unmarshall output: %s", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Unexpected result, diff(-want, got) = %s", diff)
			}
		})
	}
}
