package utils

type PaginationData struct {
	PreviousPage int
	NextPage     int
	CurrentPage  int
	PrevDisabled string
	NextDisabled string
}

func GetPaginationData(page, pageSize, totalItems int) PaginationData {
	prevDisabled := ""
	if page == 1 {
		prevDisabled = "disabled"
	}

	nextDisabled := ""
	if page*pageSize >= totalItems {
		nextDisabled = "disabled"
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	nextPage := page + 1
	if page*pageSize >= totalItems {
		nextPage = page
	}

	return PaginationData{
		PreviousPage: prevPage,
		NextPage:     nextPage,
		CurrentPage:  page,
		PrevDisabled: prevDisabled,
		NextDisabled: nextDisabled,
	}
}
