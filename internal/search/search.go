package search

func SearchCodebase(dir, query string, limit int) ([]SearchResult, error) {
	idx := NewIndex()
	if err := idx.Build(dir); err != nil {
		return nil, err
	}
	return idx.Search(query, limit), nil
}
