package flow

// A tag allows adding a descriptive string to a memo.
type Tag struct {
	Tag string
	Val Memo
}

// Tagged inputs know where each memo comes from.
type TaggedInput <-chan Tag

// Return a list of source ports connected to this tagged input.
func (t *TaggedInput) Sources() []string {
	return nil
}

// Tagged outpus can send a memo to one destination or to all.
type TaggedOutput interface {
	Output
	SendOne(dest string, v Memo)
	CloseOne(dest string)
	Destinations() []string
}
