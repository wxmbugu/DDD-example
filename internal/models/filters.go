package models

import "math"

type (
	Filters struct {
		PageSize int
		Page     int
	}
	Metadata struct {
		CurrentPage  int
		PageSize     int
		FirstPage    int
		LastPage     int
		TotalRecords int
	}
)

func (f Filters) Limit() int {
	return f.PageSize
}
func (f Filters) Offset() int {
	if f.Page == 1 {
		return 0
	} else {
		return (f.Page - 1) * f.PageSize
	}
}

func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Note that we return an empty Metadata struct if there are no records.
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
