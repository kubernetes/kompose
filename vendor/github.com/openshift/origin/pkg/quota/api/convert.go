package api

// ConvertClusterResourceQuotaToAppliedClusterQuota returns back a converted AppliedClusterResourceQuota which is NOT a deep copy.
func ConvertClusterResourceQuotaToAppliedClusterResourceQuota(in *ClusterResourceQuota) *AppliedClusterResourceQuota {
	return &AppliedClusterResourceQuota{
		ObjectMeta: in.ObjectMeta,
		Spec:       in.Spec,
		Status:     in.Status,
	}
}

// ConvertClusterResourceQuotaToAppliedClusterQuota returns back a converted AppliedClusterResourceQuota which is NOT a deep copy.
func ConvertAppliedClusterResourceQuotaToClusterResourceQuota(in *AppliedClusterResourceQuota) *ClusterResourceQuota {
	return &ClusterResourceQuota{
		ObjectMeta: in.ObjectMeta,
		Spec:       in.Spec,
		Status:     in.Status,
	}
}
