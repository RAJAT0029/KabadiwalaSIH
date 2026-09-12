package safety

// Guidance is a shared content contract for future collector vernacular/safety UX.
// This repository does not implement the collector mobile interface.
type Guidance struct {
	ID                 string `bson:"id" json:"id"`
	MaterialCategory   string `bson:"materialCategory" json:"materialCategory"`
	Title              string `bson:"title" json:"title"`
	PictorialReference string `bson:"pictorialReference,omitempty" json:"pictorialReference,omitempty"`
	AudioReference     string `bson:"audioReference,omitempty" json:"audioReference,omitempty"`
	Language           string `bson:"language" json:"language"`
	HazardLevel        string `bson:"hazardLevel" json:"hazardLevel"`
	GuidanceText       string `bson:"guidanceText" json:"guidanceText"`
}
