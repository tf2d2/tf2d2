package aws

var (
	// Nodes represents supported AWS node resources
	Nodes = map[string]bool{
		"aws_alb":         true,
		"aws_ecs_cluster": true,
		"aws_ecs_service": true,
		"aws_elb":         true,
		"aws_lb":          true,
		"aws_subnet":      true,
		"aws_vpc":         true,
	}
)
