package schedule_info

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
)

func GetRoleEligibilityScheduleInfo(
	startDateTime *time.Time,
	endDateTime *time.Time,
) *armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo {
	var expiration armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration

	if endDateTime == nil {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			Type: to.Ptr(armauthorization.TypeNoExpiration),
		}
	} else {
		expiration = armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfoExpiration{
			EndDateTime: endDateTime,
			Type:        to.Ptr(armauthorization.TypeAfterDateTime),
		}
	}

	if startDateTime == nil {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: to.Ptr(time.Now()),
		}
	} else {
		return &armauthorization.RoleEligibilityScheduleRequestPropertiesScheduleInfo{
			Expiration:    &expiration,
			StartDateTime: startDateTime,
		}
	}
}
