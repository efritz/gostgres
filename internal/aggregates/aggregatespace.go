package aggregates

type Aggregatespace struct {
	aggregates map[string]Aggregate
}

func NewDefaultAggregatespace() *Aggregatespace {
	return &Aggregatespace{
		aggregates: DefaultAggregates(),
	}
}

func NewAggregatespace() *Aggregatespace {
	return &Aggregatespace{
		aggregates: map[string]Aggregate{},
	}
}

func (t *Aggregatespace) GetAggregate(name string) (Aggregate, bool) {
	aggregate, ok := t.aggregates[name]
	return aggregate, ok
}

func (t *Aggregatespace) SetFunction(name string, aggregate Aggregate) error {
	t.aggregates[name] = aggregate
	return nil
}
