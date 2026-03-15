package obsidian

import "context"

// Batch runs a callback for every note returned by the query.
func (q *Query) Batch(ctx context.Context, fn BatchFunc) ([]BatchItemResult, error) {
	notes, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]BatchItemResult, 0, len(notes))
	for _, note := range notes {
		value, itemErr := fn(ctx, note)
		results = append(results, BatchItemResult{
			Path:  note.Path(),
			Value: value,
			Err:   itemErr,
		})
		if itemErr != nil {
			return results, itemErr
		}
	}
	return results, nil
}
