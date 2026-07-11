package dynamo

// SetRelated sets recipeID's related set and reconciles the symmetric reverse
// relation on each counterpart. Normalizes the request (dedupe, drop self, drop
// non-existent), diffs against the current set, and writes both sides.
//
// DynamoDB has no partial-update helper here (Update == PutItem of the whole
// item), so each side is re-read and re-put. Writes are non-transactional: the
// edited recipe is written first, then counterparts; any error is returned so
// the caller can retry rather than leaving a one-sided relation.
func (s *RecipeStore) SetRelated(recipeID string, requested []string) error {
	self, err := s.GetByID(recipeID)
	if err != nil {
		return err
	}
	old := self.RelatedIDs
	next := s.normalizeRelated(recipeID, requested)

	self.RelatedIDs = next
	if err := s.Create(self); err != nil { // PutItem overwrites the item
		return err
	}
	for _, b := range subtract(next, old) {
		if err := s.mutateCounterpart(b, recipeID, true); err != nil {
			return err
		}
	}
	for _, b := range subtract(old, next) {
		if err := s.mutateCounterpart(b, recipeID, false); err != nil {
			return err
		}
	}
	return nil
}

// normalizeRelated dedupes, drops self, and drops ids that don't resolve to an
// existing recipe.
func (s *RecipeStore) normalizeRelated(self string, requested []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, id := range requested {
		if id == "" || id == self || seen[id] {
			continue
		}
		if _, err := s.GetByID(id); err != nil {
			continue // non-existent → drop
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// mutateCounterpart adds or removes `other` from counterpart `id`'s related set.
func (s *RecipeStore) mutateCounterpart(id, other string, add bool) error {
	r, err := s.GetByID(id)
	if err != nil {
		return err
	}
	out := make([]string, 0, len(r.RelatedIDs)+1)
	present := false
	for _, x := range r.RelatedIDs {
		if x == other {
			present = true
			if !add {
				continue // drop
			}
		}
		out = append(out, x)
	}
	if add && !present {
		out = append(out, other)
	}
	r.RelatedIDs = out
	return s.Create(r)
}

// subtract returns items in a that are not in b.
func subtract(a, b []string) []string {
	inB := map[string]bool{}
	for _, x := range b {
		inB[x] = true
	}
	var out []string
	for _, x := range a {
		if !inB[x] {
			out = append(out, x)
		}
	}
	return out
}
