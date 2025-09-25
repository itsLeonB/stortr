package entity

type OrphanedBillCleanupTask struct {
	BillObjectKeys []string `json:"billObjectKeys"`
	BucketName     string   `json:"bucketName"`
}

func (obc OrphanedBillCleanupTask) Type() string {
	return "orphaned-bill-cleanup"
}
