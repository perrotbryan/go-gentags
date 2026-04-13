package gentags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapByTag(t *testing.T) {
	in := []TaggedPackage{
		{
			Name: "main",
			Files: []TaggedFile{
				{
					Name: "/home/bryan/Projects/bpt/gentags/gentags.go",
					Structs: []TaggedStruct{
						{
							Name: "Test",
							Fields: []TaggedField{
								{
									FieldName: "Name",
									TagName:   "json",
									TagValue:  "name",
								},
								{
									FieldName: "Name",
									TagName:   "bson",
									TagValue:  "name",
								},
								{
									FieldName: "Bar",
									TagName:   "json",
									TagValue:  "name",
								},
								{
									FieldName: "Bar",
									TagName:   "bson",
									TagValue:  "name",
								},
							},
						},
					},
				},
			},
		},
	}

	expected := TagMap{
		"json": TagEntry{
			Packages: []TaggedPackage{
				{
					Name: "main",
					Files: []TaggedFile{
						{
							Name: "/home/bryan/Projects/bpt/gentags/gentags.go",
							Structs: []TaggedStruct{
								{
									Name: "Test",
									Fields: []TaggedField{
										{
											FieldName: "Name",
											TagName:   "json",
											TagValue:  "name",
										},
										{
											FieldName: "Bar",
											TagName:   "json",
											TagValue:  "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		"bson": TagEntry{
			Packages: []TaggedPackage{
				{
					Name: "main",
					Files: []TaggedFile{
						{
							Name: "/home/bryan/Projects/bpt/gentags/gentags.go",
							Structs: []TaggedStruct{
								{
									Name: "Test",
									Fields: []TaggedField{
										{
											FieldName: "Name",
											TagName:   "bson",
											TagValue:  "name",
										},
										{
											FieldName: "Bar",
											TagName:   "bson",
											TagValue:  "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	got := mapByTagName(in)
	assert.Equal(t, expected, got)
}
