package api

// Synthetic authorization endpoints
const (
	DockerBuildResource          = "builds/docker"
	SourceBuildResource          = "builds/source"
	CustomBuildResource          = "builds/custom"
	JenkinsPipelineBuildResource = "builds/jenkinspipeline"

	NodeMetricsResource = "nodes/metrics"
	NodeStatsResource   = "nodes/stats"
	NodeLogResource     = "nodes/log"

	RestrictedEndpointsResource = "endpoints/restricted"
)
