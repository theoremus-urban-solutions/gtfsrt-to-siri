package gtfsrtsiri

type SituationExchange struct {
	Situations any `json:"Situations"`
}

func (c *Converter) buildSituationExchange() SituationExchange {
	return SituationExchange{Situations: []any{}}
}
