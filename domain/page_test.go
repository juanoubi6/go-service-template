package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	Limit       = 2
	CursorValue = "someValue"
	NameA       = "A"
	NameB       = "B"
	NameC       = "C"
)

func Test_BuildCursorPage_NextDirectionWithMoreDataThanLimitOnEmptyCursor(t *testing.T) {
	data := []Location{{Name: NameA}, {Name: NameB}, {Name: NameC}}
	filters := CursorPaginationFilters{
		Cursor:    "",
		Direction: NextPage,
		Limit:     Limit,
	}

	page := BuildCursorPage(data, filters)

	assert.Equal(t, NameB, *page.NextPage)
	assert.Nil(t, page.PreviousPage)
	assert.Equal(t, []Location{{Name: NameA}, {Name: NameB}}, page.Data)
	assert.Equal(t, filters.Limit, page.Limit)
}

func Test_BuildCursorPage_NextDirectionWithMoreDataThanLimitWithCursorValue(t *testing.T) {
	data := []Location{{Name: NameA}, {Name: NameB}, {Name: NameC}}
	filters := CursorPaginationFilters{
		Cursor:    CursorValue,
		Direction: NextPage,
		Limit:     Limit,
	}

	page := BuildCursorPage(data, filters)

	assert.Equal(t, NameB, *page.NextPage)
	assert.Equal(t, NameA, *page.PreviousPage)
	assert.Equal(t, []Location{{Name: NameA}, {Name: NameB}}, page.Data)
	assert.Equal(t, filters.Limit, page.Limit)
}

func Test_BuildCursorPage_NextDirectionWithLessDataThanLimitWithCursorValue(t *testing.T) {
	data := []Location{{Name: NameC}}
	filters := CursorPaginationFilters{
		Cursor:    CursorValue,
		Direction: NextPage,
		Limit:     Limit,
	}

	page := BuildCursorPage(data, filters)

	assert.Nil(t, page.NextPage)
	assert.Equal(t, NameC, *page.PreviousPage)
	assert.Equal(t, []Location{{Name: NameC}}, page.Data)
	assert.Equal(t, filters.Limit, page.Limit)
}

func Test_BuildCursorPage_PrevDirectionWithMoreDataThanLimitWithCursorValue(t *testing.T) {
	data := []Location{{Name: NameC}, {Name: NameB}, {Name: NameA}}
	filters := CursorPaginationFilters{
		Cursor:    CursorValue,
		Direction: PreviousPage,
		Limit:     Limit,
	}

	page := BuildCursorPage(data, filters)

	assert.Equal(t, NameC, *page.NextPage)
	assert.Equal(t, NameB, *page.PreviousPage)
	assert.Equal(t, []Location{{Name: NameB}, {Name: NameC}}, page.Data)
	assert.Equal(t, filters.Limit, page.Limit)
}

func Test_BuildCursorPage_PrevDirectionWithLessDataThanLimitWithCursorValue(t *testing.T) {
	data := []Location{{Name: NameB}, {Name: NameA}}
	filters := CursorPaginationFilters{
		Cursor:    CursorValue,
		Direction: PreviousPage,
		Limit:     Limit,
	}

	page := BuildCursorPage(data, filters)

	assert.Equal(t, NameB, *page.NextPage)
	assert.Nil(t, page.PreviousPage)
	assert.Equal(t, []Location{{Name: NameA}, {Name: NameB}}, page.Data)
	assert.Equal(t, filters.Limit, page.Limit)
}

func Test_BuildCursorPage_EmptyPageWhenDataIsEmpty(t *testing.T) {
	page := BuildCursorPage([]Location{}, CursorPaginationFilters{})

	assert.Nil(t, page.PreviousPage)
	assert.Nil(t, page.NextPage)
	assert.Len(t, page.Data, 0)
	assert.Equal(t, 0, page.Limit)
}
