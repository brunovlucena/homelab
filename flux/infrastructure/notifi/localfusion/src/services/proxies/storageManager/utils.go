package storageManager

import "notifinetwork/localfusion/proto/types"

func deriveStorageId(storageType types.FusionStorageType, key string) string {
	return storageType.String() + key
}
