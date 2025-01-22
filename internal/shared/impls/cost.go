package impls

type NodeCost struct {
	FixedCost    ResourceCost
	VariableCost ResourceCost
	Statistics   RelationStatistics
}

func (nc NodeCost) Scalar() float64 {
	return nc.FixedCost.Add(nc.VariableCost.ScaleUniform(float64(nc.Statistics.RowCount))).Scalar()
}

//
//

type ResourceCost struct {
	CPU    float64
	Memory float64
	IO     float64
}

const (
	cpuMultiplier    = 1
	memoryMultiplier = 1
	ioMultiplier     = 1
)

func (rc ResourceCost) Scalar() float64 {
	return (rc.CPU * cpuMultiplier) + (rc.Memory * memoryMultiplier) + (rc.IO * ioMultiplier)
}

func (rc ResourceCost) Add(other ResourceCost) ResourceCost {
	return SumCosts(rc, other)
}

func (rc ResourceCost) Scale(other ResourceCost) ResourceCost {
	return ResourceCost{
		CPU:    rc.CPU * other.CPU,
		Memory: rc.Memory * other.Memory,
		IO:     rc.IO * other.IO,
	}
}

func (rc ResourceCost) ScaleUniform(multplier float64) ResourceCost {
	return ResourceCost{
		CPU:    rc.CPU * multplier,
		Memory: rc.Memory * multplier,
		IO:     rc.IO * multplier,
	}
}

func SumCosts(costs ...ResourceCost) ResourceCost {
	sum := ResourceCost{}
	for _, other := range costs {
		sum.CPU += other.CPU
		sum.Memory += other.Memory
		sum.IO += other.IO
	}

	return sum
}
