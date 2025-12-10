package match

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestGjsonExpandPath(t *testing.T) {
	t.Run("simple cases", func(t *testing.T) {
		tests := []struct {
			name     string
			data     []byte
			path     string
			expected []string
			wantErr  bool
		}{
			{
				name:     "simple path without array access",
				data:     []byte(`{"user": {"name": "John"}}`),
				path:     "user.name",
				expected: []string{"user.name"},
				wantErr:  false,
			},
			{
				name:     "single level array access",
				data:     []byte(`{"items": [1, 2, 3]}`),
				path:     "items.#.value",
				expected: []string{"items.0.value", "items.1.value", "items.2.value"},
				wantErr:  false,
			},
			{
				name:     "empty array",
				data:     []byte(`{"items": []}`),
				path:     "items.#.value",
				expected: []string{},
				wantErr:  false,
			},
			{
				name:     "array with single element",
				data:     []byte(`{"items": [42]}`),
				path:     "items.#.value",
				expected: []string{"items.0.value"},
				wantErr:  false,
			},
			{
				name:     "path with # at the end",
				data:     []byte(`{"items": [1, 2, 3]}`),
				path:     "items.#",
				expected: []string{"items.0", "items.1", "items.2"},
				wantErr:  false,
			},
			{
				name:     "path without # returns as-is",
				data:     []byte(`{"a": {"b": {"c": "value"}}}`),
				path:     "a.b.c",
				expected: []string{"a.b.c"},
				wantErr:  false,
			},
			{
				name:     "root level array",
				data:     []byte(`[1, 2, 3]`),
				path:     "#.value",
				expected: []string{"0.value", "1.value", "2.value"},
				wantErr:  false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				test.Equal(t, tt.expected, expandArrayPaths(tt.data, tt.path))
			})
		}
	})

	t.Run("nested arrays", func(t *testing.T) {
		tests := []struct {
			name     string
			data     []byte
			path     string
			expected []string
			wantErr  bool
		}{
			{
				name: "two-level nested arrays",
				data: []byte(`{"results": [{"packages": [1, 2]}, {"packages": [3, 4, 5]}]}`),
				path: "results.#.packages.#.vulnerabilities",
				expected: []string{
					"results.0.packages.0.vulnerabilities",
					"results.0.packages.1.vulnerabilities",
					"results.1.packages.0.vulnerabilities",
					"results.1.packages.1.vulnerabilities",
					"results.1.packages.2.vulnerabilities",
				},
				wantErr: false,
			},
			{
				name: "three levels of nesting",
				data: []byte(`{"a": [{"b": [{"c": [1, 2]}]}]}`),
				path: "a.#.b.#.c.#.d",
				expected: []string{
					"a.0.b.0.c.0.d",
					"a.0.b.0.c.1.d",
				},
				wantErr: false,
			},
			{
				name: "four levels of nesting",
				data: []byte(`{"l1": [{"l2": [{"l3": [{"l4": [1, 2]}, {"l4": [3]}]}]}]}`),
				path: "l1.#.l2.#.l3.#.l4.#.value",
				expected: []string{
					"l1.0.l2.0.l3.0.l4.0.value",
					"l1.0.l2.0.l3.0.l4.1.value",
					"l1.0.l2.0.l3.1.l4.0.value",
				},
				wantErr: false,
			},
			{
				name: "asymmetric nested arrays",
				data: []byte(`{"data": [{"items": [1, 2, 3]}, {"items": []}, {"items": [4]}]}`),
				path: "data.#.items.#.value",
				expected: []string{
					"data.0.items.0.value",
					"data.0.items.1.value",
					"data.0.items.2.value",
					"data.2.items.0.value",
				},
				wantErr: false,
			},
			{
				name: "mixed nested arrays with different lengths at each level",
				data: []byte(`{
					"regions": [
						{
							"cities": [
								{"districts": [1, 2, 3]},
								{"districts": [4]}
							]
						},
						{
							"cities": [
								{"districts": []},
								{"districts": [5, 6]}
							]
						},
						{
							"cities": []
						}
					]
				}`),
				path: "regions.#.cities.#.districts.#.name",
				expected: []string{
					"regions.0.cities.0.districts.0.name",
					"regions.0.cities.0.districts.1.name",
					"regions.0.cities.0.districts.2.name",
					"regions.0.cities.1.districts.0.name",
					"regions.1.cities.1.districts.0.name",
					"regions.1.cities.1.districts.1.name",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				test.Equal(t, tt.expected, expandArrayPaths(tt.data, tt.path))
			})
		}
	})

	t.Run("complex real-world scenarios", func(t *testing.T) {
		tests := []struct {
			name     string
			data     []byte
			path     string
			expected []string
			wantErr  bool
		}{
			{
				name: "GitHub API response structure",
				data: []byte(`{
					"repositories": [
						{
							"name": "repo1",
							"commits": [
								{"sha": "abc123", "files": [{"path": "a.js", "checksum": "c1"}, {"path": "b.js", "checksum": "c2"}]},
								{"sha": "def456", "files": [{"path": "c.js", "checksum": "c3"}]}
							]
						},
						{
							"name": "repo2",
							"commits": [
								{
									"sha": "ghi789", 
									"files": [
										{"path": "d.js", "checksum": "c4"}, 
										{"path": "e.js", "checksum": "c5"}, 
										{"path": "f.js", "checksum": "c6"}
									]
								}
							]
						}
					]
				}`),
				path: "repositories.#.commits.#.files.#.checksum",
				expected: []string{
					"repositories.0.commits.0.files.0.checksum",
					"repositories.0.commits.0.files.1.checksum",
					"repositories.0.commits.1.files.0.checksum",
					"repositories.1.commits.0.files.0.checksum",
					"repositories.1.commits.0.files.1.checksum",
					"repositories.1.commits.0.files.2.checksum",
				},
				wantErr: false,
			},
			{
				name: "vulnerability scanner output",
				data: []byte(`{
					"scan_results": [
						{
							"target": "app1",
							"dependencies": [
								{
									"name": "lodash",
									"vulnerabilities": [
										{"severity": "high", "cves": ["CVE-2021-1", "CVE-2021-2"]},
										{"severity": "medium", "cves": ["CVE-2020-1"]}
									]
								},
								{
									"name": "express",
									"vulnerabilities": []
								}
							]
						},
						{
							"target": "app2",
							"dependencies": [
								{
									"name": "react",
									"vulnerabilities": [
										{"severity": "low", "cves": ["CVE-2019-1"]}
									]
								}
							]
						}
					]
				}`),
				path: "scan_results.#.dependencies.#.vulnerabilities.#.id",
				expected: []string{
					"scan_results.0.dependencies.0.vulnerabilities.0.id",
					"scan_results.0.dependencies.0.vulnerabilities.1.id",
					"scan_results.1.dependencies.0.vulnerabilities.0.id",
				},
				wantErr: false,
			},
			{
				name: "e-commerce order with nested items",
				data: []byte(`{
					"orders": [
						{
							"id": "order1",
							"items": [
								{
									"product": "laptop",
									"variants": [
										{"sku": "LAP-001", "reviews": [{"rating": 5}, {"rating": 4}]},
										{"sku": "LAP-002", "reviews": [{"rating": 3}]}
									]
								},
								{
									"product": "mouse",
									"variants": [
										{"sku": "MOU-001", "reviews": []}
									]
								}
							]
						},
						{
							"id": "order2",
							"items": []
						}
					]
				}`),
				path: "orders.#.items.#.variants.#.reviews.#.verified",
				expected: []string{
					"orders.0.items.0.variants.0.reviews.0.verified",
					"orders.0.items.0.variants.0.reviews.1.verified",
					"orders.0.items.0.variants.1.reviews.0.verified",
				},
				wantErr: false,
			},
			{
				name: "cloud infrastructure with regions and zones",
				data: []byte(`{
					"cloud_providers": [
						{
							"name": "AWS",
							"regions": [
								{
									"code": "us-east-1",
									"zones": [
										{"name": "us-east-1a", "instances": [{"id": "i-1"}, {"id": "i-2"}]},
										{"name": "us-east-1b", "instances": [{"id": "i-3"}]}
									]
								},
								{
									"code": "us-west-2",
									"zones": [
										{"name": "us-west-2a", "instances": []}
									]
								}
							]
						},
						{
							"name": "GCP",
							"regions": [
								{
									"code": "us-central1",
									"zones": [
										{"name": "us-central1-a", "instances": [{"id": "gcp-1"}]}
									]
								}
							]
						}
					]
				}`),
				path: "cloud_providers.#.regions.#.zones.#.instances.#.status",
				expected: []string{
					"cloud_providers.0.regions.0.zones.0.instances.0.status",
					"cloud_providers.0.regions.0.zones.0.instances.1.status",
					"cloud_providers.0.regions.0.zones.1.instances.0.status",
					"cloud_providers.1.regions.0.zones.0.instances.0.status",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				test.Equal(t, tt.expected, expandArrayPaths(tt.data, tt.path))
			})
		}
	})

	t.Run("edge cases and boundary conditions", func(t *testing.T) {
		tests := []struct {
			name          string
			data          []byte
			path          string
			expected      []string
			expectedError error
		}{
			{
				name: "array with many elements (10+)",
				data: []byte(`{"items": [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]}`),
				path: "items.#.value",
				expected: []string{
					"items.0.value", "items.1.value", "items.2.value", "items.3.value",
					"items.4.value", "items.5.value", "items.6.value", "items.7.value",
					"items.8.value", "items.9.value", "items.10.value", "items.11.value",
					"items.12.value", "items.13.value", "items.14.value", "items.15.value",
				},
				expectedError: nil,
			},
			{
				name: "nested array with 100+ total paths",
				data: func() []byte {
					// Generate JSON with 10x10 nested structure = 100 paths
					type Inner struct {
						Values []int `json:"values"`
					}
					type Outer struct {
						Items []Inner `json:"items"`
					}
					root := struct {
						Data []Outer `json:"data"`
					}{
						Data: make([]Outer, 10),
					}
					for i := range root.Data {
						root.Data[i].Items = make([]Inner, 10)
						for j := range root.Data[i].Items {
							root.Data[i].Items[j].Values = []int{1}
						}
					}
					b, _ := json.Marshal(root)
					return b
				}(),
				path: "data.#.items.#.values.#.id",
				expected: func() []string {
					paths := make([]string, 0, 100)
					for i := range 10 {
						for j := range 10 {
							paths = append(paths, fmt.Sprintf("data.%d.items.%d.values.0.id", i, j))
						}
					}
					return paths
				}(),
				expectedError: nil,
			},
			{
				name:          "all empty arrays in nested structure",
				data:          []byte(`{"a": [{"b": []}, {"b": []}]}`),
				path:          "a.#.b.#.c",
				expected:      []string{},
				expectedError: nil,
			},
			{
				name:          "single element at each level",
				data:          []byte(`{"a": [{"b": [{"c": [{"d": [1]}]}]}]}`),
				path:          "a.#.b.#.c.#.d.#.e",
				expected:      []string{"a.0.b.0.c.0.d.0.e"},
				expectedError: nil,
			},
			{
				name: "highly unbalanced tree structure",
				data: []byte(`{
					"items": [
						{"children": [{"grandchildren": [1,2,3,4,5,6,7,8,9,10]}]},
						{"children": [{"grandchildren": [1]}]},
						{"children": [{"grandchildren": []}]},
						{"children": []}
					]
				}`),
				path: "items.#.children.#.grandchildren.#.value",
				expected: []string{
					"items.0.children.0.grandchildren.0.value",
					"items.0.children.0.grandchildren.1.value",
					"items.0.children.0.grandchildren.2.value",
					"items.0.children.0.grandchildren.3.value",
					"items.0.children.0.grandchildren.4.value",
					"items.0.children.0.grandchildren.5.value",
					"items.0.children.0.grandchildren.6.value",
					"items.0.children.0.grandchildren.7.value",
					"items.0.children.0.grandchildren.8.value",
					"items.0.children.0.grandchildren.9.value",
					"items.1.children.0.grandchildren.0.value",
				},
				expectedError: nil,
			},
			{
				name: "multiple # in different branches",
				data: []byte(`{
					"branch_a": [{"data": [1, 2]}],
					"branch_b": [{"data": [3, 4, 5]}]
				}`),
				path:          "branch_a.#.data.#.value",
				expected:      []string{"branch_a.0.data.0.value", "branch_a.0.data.1.value"},
				expectedError: nil,
			},
			{
				name:          "path with non-existent intermediate key",
				data:          []byte(`{"items": []}`),
				path:          "nonexistent.#.value",
				expected:      []string{},
				expectedError: errPathNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				test.Equal(t, tt.expected, expandArrayPaths(tt.data, tt.path))
			})
		}
	})
}
